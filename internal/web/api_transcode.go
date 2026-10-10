package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"silo/internal/ffmpeg"
	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/job"
	"silo/internal/ffmpeg/media"
	"silo/internal/ffmpeg/probe"
	"silo/internal/ffmpeg/profile"
	"silo/internal/ffmpeg/source"
	"silo/internal/log"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// SessionHeader - заголовок с идентификатором сессии транскодирования.
const SessionHeader = "X-Silo-Transcode-Session"

// handleTranscodeInfo отдаёт сведения о файле, полученные через ffprobe.
func (s *Server) handleTranscodeInfo(c *gin.Context) {
	hashHex := c.Param("hash")
	fileIdx, ok := parseFileIdx(c)
	if !ok {
		return
	}

	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if !s.transcodeEnabled() {
		writeTranscodeUnavailable(c, s.transcodeReason())
		return
	}

	lease, info, err := s.probeSource(c.Request.Context(), currentUser, hashHex, fileIdx)
	if err != nil {
		writeTranscodeError(c, err)
		return
	}
	defer s.sourceRegistry.Release(lease.ID)

	c.JSON(http.StatusOK, gin.H{
		"hash":        hashHex,
		"file_idx":    fileIdx,
		"container":   info.Container,
		"format":      info.FormatName,
		"duration":    info.Duration,
		"duration_h":  media.FormatDuration(info.Duration),
		"size":        info.Size,
		"bitrate":     info.Bitrate,
		"streams":     info.Streams,
		"has_video":   info.HasVideo(),
		"has_audio":   info.HasAudio(),
		"video_codec": codecOf(info.Video()),
		"audio_codec": codecOf(info.Audio()),
		"is_hdr":      info.IsHDR(),
		"tonemap":     info.IsHDR(),
	})
}

// handleTranscodeStream запускает транскодирование и отдаёт поток клиенту.
func (s *Server) handleTranscodeStream(c *gin.Context) {
	hashHex := c.Param("hash")
	fileIdx, ok := parseFileIdx(c)
	if !ok {
		return
	}

	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if !s.transcodeEnabled() {
		writeTranscodeUnavailable(c, s.transcodeReason())
		return
	}

	prof, err := s.resolveProfile(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if prof.Protocol == args.ProtocolHLS {
		// Для HLS клиент работает через плейлисты, поток отдавать не нужно.
		// Параметры запроса переносятся целиком: иначе теряются позиция и настройки профиля
		c.Redirect(http.StatusFound, masterPlaylistURL(hashHex, fileIdx, c.Request.URL.RawQuery))
		return
	}

	lease, info, err := s.probeSource(c.Request.Context(), currentUser, hashHex, fileIdx)
	if err != nil {
		writeTranscodeError(c, err)
		return
	}

	options, err := prof.Resolve(s.SourceURL(lease.ID), info)
	if err != nil {
		s.sourceRegistry.Release(lease.ID)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	options.Seek = parseSeek(c)

	session, err := s.transcoder.Start(c.Request.Context(), ffmpeg.StartRequest{
		UserID:    currentUser.ID,
		Hash:      hashHex,
		FileIdx:   fileIdx,
		ProfileID: prof.ID,
		Options:   options,
		Duration:  secondsToDuration(info.Duration),
	})
	if err != nil {
		s.sourceRegistry.Release(lease.ID)
		writeTranscodeError(c, err)
		return
	}

	log.Infof("[Web] Transcode session %s started for %s/%d", session.ID, hashHex, fileIdx)

	// Доступ живёт до конца потока: ffmpeg читает файл раздачи всё это время
	defer s.sourceRegistry.Release(lease.ID)
	s.serveProgressiveStream(c, session)
}

// serveProgressiveStream отдаёт поток ffmpeg клиенту по мере появления данных.
func (s *Server) serveProgressiveStream(c *gin.Context, session *job.Job) {
	// Разрыв соединения останавливает процесс через контекст запроса,
	// здесь достаточно освободить ресурсы, если клиент ушёл молча
	defer session.Stop(s.stopTimeout())

	out := session.Output()
	if out == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "stream is unavailable"})
		return
	}

	c.Header(SessionHeader, session.ID)
	c.Header("Content-Type", streamContentType(session))
	// Поток незавершён, поэтому длину заранее отдать нельзя и Range не поддерживается
	c.Header("Accept-Ranges", "none")
	c.Header("Cache-Control", "no-store")
	c.Header("transferMode.dlna.org", "Streaming")

	handle := s.streamTracker.Open(StreamOpen{
		UserID:    session.UserID,
		Hash:      session.Hash,
		FileIdx:   session.FileIdx,
		FileName:  session.ID,
		ClientIP:  c.RemoteIP(),
		UserAgent: c.Request.UserAgent(),
	})
	defer handle.Close()

	c.Writer = &countingWriter{ResponseWriter: c.Writer, handle: handle}
	c.Status(http.StatusOK)

	if _, err := io.Copy(newFlushWriter(c), out); err != nil && !isClientGone(err) {
		log.Warnf("[Web] Session %s stream ended: %v", session.ID, err)
	}

	select {
	case <-session.Done():
	case <-time.After(s.stopTimeout()):
	}
}

// handleTranscodeSessions отдаёт список активных сессий транскодирования.
func (s *Server) handleTranscodeSessions(c *gin.Context) {
	if !s.transcodeEnabled() {
		writeTranscodeUnavailable(c, s.transcodeReason())
		return
	}

	sessions := s.transcoder.Jobs()
	if sessions == nil {
		sessions = []job.Snapshot{}
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// handleTranscodeStop останавливает сессию транскодирования.
func (s *Server) handleTranscodeStop(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if !s.transcodeEnabled() {
		writeTranscodeUnavailable(c, s.transcodeReason())
		return
	}

	id := c.Param("id")
	for _, snap := range s.transcoder.Jobs() {
		if snap.ID != id {
			continue
		}
		if snap.UserID != currentUser.ID && currentUser.Rank < user.RankAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}
		if !s.transcoder.Stop(id) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "stopped"})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
}

// flushWriter отдаёт данные клиенту сразу, не накапливая их в буфере.
type flushWriter struct {
	c *gin.Context
}

// newFlushWriter создаёт писатель, сбрасывающий буфер после каждой записи.
func newFlushWriter(c *gin.Context) *flushWriter {
	return &flushWriter{c: c}
}

func (w *flushWriter) Write(p []byte) (int, error) {
	n, err := w.c.Writer.Write(p)
	if err == nil {
		w.c.Writer.Flush()
	}
	return n, err
}

// resolveProfile выбирает параметры транскодирования для запроса.
// Базой служит профиль из базы, а параметры запроса его переопределяют.
func (s *Server) resolveProfile(c *gin.Context) (profile.Profile, error) {
	prof, err := s.baseProfile(profileActor(c), strings.TrimSpace(c.Query("profile")))
	if err != nil {
		return profile.Profile{}, err
	}

	if v := strings.TrimSpace(c.Query("container")); v != "" {
		prof.Container = v
	}
	if v := strings.TrimSpace(c.Query("video_codec")); v != "" {
		prof.Video.Codec = v
	}
	if v := strings.TrimSpace(c.Query("audio_codec")); v != "" {
		prof.Audio.Codec = v
	}
	if v := strings.TrimSpace(c.Query("preset")); v != "" {
		prof.Video.Preset = v
	}
	if v := strings.TrimSpace(c.Query("hwaccel")); v != "" {
		prof.Video.HWAccel = v
	}
	if v := strings.TrimSpace(c.Query("protocol")); v != "" {
		prof.Protocol = args.Protocol(v)
	}
	if v := strings.TrimSpace(c.Query("segment_type")); v != "" {
		prof.HLS.SegmentType = v
	}
	if v := strings.TrimSpace(c.Query("transpose")); v != "" {
		prof.Video.Transpose = v
	}
	if v := strings.TrimSpace(c.Query("subtitle_mode")); v != "" {
		prof.Subtitles.Mode = args.SubtitleMode(v)
	}
	if v, ok := parseIntQuery(c, "subtitle_index"); ok {
		prof.Subtitles.StreamIndex = &v
	}
	// tonemap_mode задаёт режим явно, tonemap остаётся понятным сокращением
	if v := strings.TrimSpace(c.Query("tonemap_mode")); v != "" {
		prof.Video.TonemapMode = args.ParseTonemapMode(v)
	}
	if v, ok := boolQuery(c, "tonemap"); ok {
		if v {
			prof.Video.TonemapMode = args.TonemapOn
		} else {
			prof.Video.TonemapMode = args.TonemapOff
		}
	}
	if v := strings.TrimSpace(c.Query("tonemap_algorithm")); v != "" {
		prof.Video.TonemapAlgorithm = v
	}
	if v, ok := boolQuery(c, "hw_decode"); ok {
		prof.Video.HWDecode = v
	}
	if v, ok := parseIntQuery(c, "segment_length"); ok {
		prof.HLS.SegmentLength = v
	}

	if v, ok := parseIntQuery(c, "crf"); ok {
		prof.Video.CRF = v
		prof.Video.BitrateKbps = 0
	}
	if v, ok := parseIntQuery(c, "video_bitrate"); ok {
		prof.Video.BitrateKbps = v
	}
	if v, ok := parseIntQuery(c, "audio_bitrate"); ok {
		prof.Audio.BitrateKbps = v
	}
	if v, ok := parseIntQuery(c, "audio_channels"); ok {
		prof.Audio.Channels = v
	}
	if v, ok := parseIntQuery(c, "max_width"); ok {
		prof.Video.MaxWidth = v
	}
	if v, ok := parseIntQuery(c, "max_height"); ok {
		prof.Video.MaxHeight = v
	}

	if err := prof.Validate(); err != nil {
		return profile.Profile{}, err
	}
	return prof, nil
}

// baseProfile возвращает профиль из базы или встроенный, если сервис недоступен.
// Пустой идентификатор означает выбор пользователя, затем профиль по умолчанию.
func (s *Server) baseProfile(actor profile.Actor, id string) (profile.Profile, error) {
	if s.profiles == nil || actor.ID == "" {
		if id != "" && id != profile.DefaultID {
			return profile.Profile{}, errors.New("unknown profile: " + id)
		}
		return profile.Default(), nil
	}

	// Пользователь не указал профиль: берём его предпочтение
	if id == "" && actor.ID != "" {
		if pref := s.preferredProfile(actor.ID); pref != "" {
			id = pref
		}
	}

	p, err := s.profiles.ResolveProfile(actor, id)
	if err != nil {
		return profile.Profile{}, err
	}
	return *p, nil
}

// probeSource выдаёт доступ к файлу раздачи и разбирает его через ffprobe.
func (s *Server) probeSource(ctx context.Context, u *user.User, hashHex string, fileIdx int) (*source.Lease, *media.MediaInfo, error) {
	lease := s.sourceRegistry.Issue(u.ID, hashHex, fileIdx)

	info, err := s.transcoder.Probe(ctx, s.SourceURL(lease.ID))
	if err != nil {
		s.sourceRegistry.Release(lease.ID)
		return nil, nil, err
	}
	return lease, info, nil
}

// parseFileIdx разбирает индекс файла из пути.
func parseFileIdx(c *gin.Context) (int, bool) {
	idx, err := strconv.Atoi(c.Param("fileIdx"))
	if err != nil || idx < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file index"})
		return 0, false
	}
	return idx, true
}

// parseSeek разбирает позицию старта в секундах.
func parseSeek(c *gin.Context) float64 {
	raw := strings.TrimSpace(c.Query("t"))
	if raw == "" {
		raw = strings.TrimSpace(c.Query("start"))
	}
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return 0
	}
	return v
}

// masterPlaylistURL собирает адрес мастер-плейлиста с исходными параметрами.
func masterPlaylistURL(hash string, fileIdx int, rawQuery string) string {
	base := fmt.Sprintf("/api/transcode/hls/%s/%d/master.m3u8", hash, fileIdx)
	if rawQuery == "" {
		return base
	}
	return base + "?" + rawQuery
}

// boolQuery возвращает логический параметр запроса.
func boolQuery(c *gin.Context, name string) (bool, bool) {
	raw := strings.TrimSpace(strings.ToLower(c.Query(name)))
	if raw == "" {
		return false, false
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

// parseIntQuery возвращает целочисленный параметр запроса.
func parseIntQuery(c *gin.Context, name string) (int, bool) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return 0, false
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, false
	}
	return v, true
}

// streamContentType определяет тип потока по контейнеру сессии.
func streamContentType(session *job.Job) string {
	switch session.Protocol {
	case string(args.ProtocolHLS):
		return "application/vnd.apple.mpegurl"
	default:
		return "video/mp4"
	}
}

// segmentContentType задаёт тип сегмента.
// Расширение .ts закреплено и за форматом переводов, поэтому тип задаётся явно.
func segmentContentType(name string, prof profile.Profile) string {
	lower := strings.ToLower(name)

	// Начальный сегмент fmp4 это обычный mp4 с описанием дорожек
	if lower == strings.ToLower(prof.HLS.InitFilename) || strings.HasSuffix(lower, ".mp4") {
		return "video/mp4"
	}
	if strings.HasSuffix(lower, ".m4s") {
		return "video/iso.segment"
	}
	return "video/mp2t"
}

// codecOf возвращает название кодека потока или пустую строку.
func codecOf(stream *media.Stream) string {
	if stream == nil {
		return ""
	}
	return stream.Codec
}

// secondsToDuration переводит секунды в длительность.
func secondsToDuration(seconds float64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}

// writeTranscodeUnavailable отвечает отказом, когда транскодинг выключен.
func writeTranscodeUnavailable(c *gin.Context, reason string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error":  "transcoding is unavailable",
		"reason": reason,
	})
}

// writeTranscodeError переводит ошибку транскодирования в HTTP-ответ.
func writeTranscodeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, job.ErrTooManySessions):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case errors.Is(err, job.ErrNoBinary):
		writeTranscodeUnavailable(c, "ffmpeg is not configured")
	case errors.Is(err, job.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "transcoding session not found"})
	case errors.Is(err, job.ErrPauseUnsupported):
		c.JSON(http.StatusNotImplemented, gin.H{"error": "this ffmpeg build does not support pause"})
	case errors.Is(err, probe.ErrTimeout):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "failed to read media info"})
	case isSourceUnavailable(err):
		// Источник не открылся: раздачи нет, файл недоступен или поток битый
		c.JSON(http.StatusNotFound, gin.H{"error": "media source is unavailable"})
	default:
		log.Errorf("[Web] Transcode request failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// isSourceUnavailable проверяет, что ошибка означает недоступный источник.
func isSourceUnavailable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	markers := []string{
		"not found",
		"no such file",
		"connection refused",
		"connection timed out",
		"server returned 404",
		"invalid data found",
		"could not open",
		"failed to open",
	}
	for _, m := range markers {
		if strings.Contains(msg, m) {
			return true
		}
	}
	return false
}

// isClientGone проверяет, что ошибка вызвана отключением клиента.
func isClientGone(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "context canceled")
}

// preferredProfile возвращает профиль, выбранный пользователем в его карточке.
func (s *Server) preferredProfile(userID string) string {
	u, err := s.userSvc.GetUserByID(userID)
	if err != nil {
		return ""
	}
	return u.TranscodeProfile
}

// handleTranscodeModule отдаёт состояние модуля транскодирования.
func (s *Server) handleTranscodeModule(c *gin.Context) {
	if s.transcoder == nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"reason":  "transcoding module is not linked",
		})
		return
	}
	c.JSON(http.StatusOK, s.transcoder.Info())
}

// ProbeFile разбирает файл раздачи и возвращает сведения о потоках.
// Метод нужен другим модулям, например плагинам.
func (s *Server) ProbeFile(ctx context.Context, userID, hash string, fileIdx int) (*media.MediaInfo, error) {
	if !s.transcodeEnabled() {
		return nil, errors.New("transcoding is unavailable")
	}

	u, err := s.userSvc.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	lease, info, err := s.probeSource(ctx, u, hash, fileIdx)
	if err != nil {
		return nil, err
	}
	s.sourceRegistry.Release(lease.ID)
	return info, nil
}

// handleTranscodePause приостанавливает вывод активной сессии.
func (s *Server) handleTranscodePause(c *gin.Context) {
	s.setSessionPaused(c, true)
}

// handleTranscodeResume возобновляет вывод активной сессии.
func (s *Server) handleTranscodeResume(c *gin.Context) {
	s.setSessionPaused(c, false)
}

// setSessionPaused управляет паузой сессии.
func (s *Server) setSessionPaused(c *gin.Context, paused bool) {
	if !s.transcodeEnabled() {
		writeTranscodeUnavailable(c, s.transcodeReason())
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id is required"})
		return
	}

	var err error
	if paused {
		err = s.transcoder.Pause(id)
	} else {
		err = s.transcoder.Resume(id)
	}
	if err != nil {
		writeTranscodeError(c, err)
		return
	}

	action := "resumed"
	if paused {
		action = "paused"
	}
	log.Infof("[Web] Transcode session %s %s", id, action)
	c.JSON(http.StatusOK, gin.H{"status": action})
}
