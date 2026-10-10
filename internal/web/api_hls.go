package web

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"silo/internal/ffmpeg"
	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/hls"
	"silo/internal/ffmpeg/job"
	"silo/internal/ffmpeg/media"
	"silo/internal/ffmpeg/profile"
	"silo/internal/log"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

const (
	// hlsForwardGap - на сколько сегментов вперёд ждём без перезапуска процесса
	hlsForwardGap = 12
	// hlsIdleTimeout - сколько сессия живёт без запросов сегментов
	hlsIdleTimeout = 2 * time.Minute
)

// hlsSession описывает сессию сегментированного транскодирования.
// Одна сессия обслуживает много запросов плеера: плейлист и сегменты.
type hlsSession struct {
	ID      string
	UserID  string
	Hash    string
	FileIdx int

	// key хранит ключ реестра на момент создания
	key string

	Profile profile.Profile
	Info    *media.MediaInfo
	LeaseID string

	mu         sync.Mutex
	job        *job.Job
	startIndex int       // Номер сегмента, с которого запущен текущий процесс
	lastAccess time.Time // Время последнего обращения плеера
	stopped    bool
}

// handleHLSMaster отдаёт главный плейлист, создавая сессию при необходимости.
func (s *Server) handleHLSMaster(c *gin.Context) {
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
	prof.Protocol = args.ProtocolHLS

	session, err := s.hlsSessionFor(c.Request.Context(), currentUser, hashHex, fileIdx, prof)
	if err != nil {
		writeTranscodeError(c, err)
		return
	}

	bandwidth := session.bandwidth()
	// Ссылка абсолютная: плеер запрашивает её с того же узла
	uri := fmt.Sprintf("/api/transcode/hls/session/%s/main.m3u8?token=%s", session.ID, c.Query("token"))

	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.Header("Cache-Control", "no-store")
	c.String(http.StatusOK, hls.BuildMaster(uri, bandwidth, codecsFor(session.Profile)))
}

// handleHLSPlaylist отдаёт медиаплейлист сессии с путями наших сегментов.
func (s *Server) handleHLSPlaylist(c *gin.Context) {
	session, ok := s.registry().ByID(c.Param("sessionID"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	session.touch()

	// Префикс абсолютный: плеер запрашивает сегменты у того же узла.
	// Токен переносится в каждый адрес, иначе плеер получит отказ доступа
	prefix := fmt.Sprintf("/api/transcode/hls/session/%s/segment/", session.ID)
	suffix := ""
	if token := c.Query("token"); token != "" {
		// Плеер обязан повторить токен в каждом запросе сегмента
		suffix = "?token=" + token
	}

	p, err := hls.Parse(session.playlistPath(s))
	if err != nil {
		if errors.Is(err, hls.ErrNotReady) || errors.Is(err, hls.ErrNoSegments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "playlist is not ready"})
			return
		}
		log.Errorf("[Web] Failed to read playlist for %s: %v", session.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read playlist"})
		return
	}

	session.cleanup(s)

	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.Header("Cache-Control", "no-store")
	c.String(http.StatusOK, p.Rewrite(prefix, suffix))
}

// handleHLSSegment отдаёт сегмент, дожидаясь его готовности.
func (s *Server) handleHLSSegment(c *gin.Context) {
	session, ok := s.registry().ByID(c.Param("sessionID"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	name := c.Param("name")
	dir := session.workDir(s)

	path, err := hls.SafeSegmentPath(dir, name)
	if err != nil {
		log.Warnf("[Web] Rejected segment name %q for session %s", name, session.ID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment name"})
		return
	}

	// Начальный сегмент бывает только у fmp4: у mpegts его нет и ждать нечего
	if name == session.initSegmentName() {
		if session.Profile.HLS.SegmentType != "fmp4" {
			c.JSON(http.StatusNotFound, gin.H{"error": "session has no init segment"})
			return
		}
		if err := waitForFile(path, s.hlsWaitTimeout()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "init segment is not available"})
			return
		}
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Type", segmentContentType(name, session.Profile))
		http.ServeFile(c.Writer, c.Request, path)
		return
	}

	index, ok := hls.SegmentIndex(name)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment name"})
		return
	}

	session.touch()

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.hlsWaitTimeout())
	defer cancel()

	if err := session.ensure(ctx, s, index); err != nil {
		log.Warnf("[Web] Segment %s of session %s is unavailable: %v", name, session.ID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "segment is not available"})
		return
	}

	session.markServed()

	// Плейлист скользит, поэтому сегменты не кэшируем
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Type", segmentContentType(name, session.Profile))
	http.ServeFile(c.Writer, c.Request, path)
}

// ensureSegment дожидается готовности сегмента, перезапуская процесс при отставании.
func (h *hlsSession) ensure(ctx context.Context, s *Server, index int) error {
	dir := h.workDir(s)
	name := h.segmentName(index)

	if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
		return nil
	}

	// Процесс отстал: запрошен сегмент до его начала или далеко впереди
	h.mu.Lock()
	start := h.startIndex
	stopped := h.stopped
	h.mu.Unlock()

	if stopped {
		return errors.New("session is stopped")
	}

	if index < start || index > start+hlsForwardGap {
		if err := h.restart(ctx, s, index); err != nil {
			return err
		}
	}

	done := h.done()
	return hls.WaitForSegment(h.playlistPath(s), dir, name, s.hlsWaitTimeout(), done)
}

// restart перезапускает процесс с сегмента, который запросил плеер.
func (h *hlsSession) restart(ctx context.Context, s *Server, index int) error {
	h.mu.Lock()
	instance := h.job
	h.stopped = false
	h.mu.Unlock()

	if instance != nil {
		instance.Stop(s.stopTimeout())
		<-instance.Done()
	}

	startAt := float64(index) * float64(h.segmentLength())
	if h.Info != nil && h.Info.Duration > 0 && startAt > h.Info.Duration {
		startAt = h.Info.Duration
	}

	log.Debugf("[Web] Session %s restarts at %.3fs for segment %d", h.ID, startAt, index)
	return h.start(ctx, s, startAt, index)
}

// start запускает процесс с указанной позиции.
func (h *hlsSession) start(ctx context.Context, s *Server, startAt float64, startIndex int) error {
	h.mu.Lock()
	stopped := h.stopped
	h.mu.Unlock()
	if stopped {
		return errors.New("session is stopped")
	}

	options, err := h.Profile.Resolve(s.SourceURL(h.LeaseID), h.Info)
	if err != nil {
		return err
	}
	options.Seek = startAt
	options.Output = ""

	dir := h.workDir(s)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create session dir: %w", err)
	}

	// При старте с нуля остатки прошлого запуска не нужны
	if startIndex == 0 {
		removeSegments(dir)
		_ = os.Remove(h.playlistPath(s))
	}

	// Сессия живёт дольше одного запроса, поэтому процесс не привязан к контексту плеера
	instance, err := s.transcoder.Start(context.WithoutCancel(ctx), ffmpeg.StartRequest{
		ID:        h.ID,
		UserID:    h.UserID,
		Hash:      h.Hash,
		FileIdx:   h.FileIdx,
		ProfileID: h.Profile.ID,
		Options:   options,
		Duration:  secondsToDuration(h.Info.Duration),
	})
	if err != nil {
		return err
	}

	h.mu.Lock()
	h.job = instance
	h.startIndex = startIndex
	h.mu.Unlock()
	return nil
}

// handleHLSStop останавливает сессию по запросу клиента.
func (s *Server) handleHLSStop(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	session, ok := s.registry().ByID(c.Param("sessionID"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if session.UserID != currentUser.ID && currentUser.Rank < user.RankAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	s.stopHLSSession(session)
	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
}

// hlsSessionFor возвращает активную сессию для файла или создаёт новую.
func (s *Server) hlsSessionFor(ctx context.Context, u *user.User, hash string, fileIdx int, prof profile.Profile) (*hlsSession, error) {
	key := hlsKey(hash, fileIdx, u.ID, profileFingerprint(prof))
	if session, ok := s.registry().ByKey(key); ok {
		session.touch()
		return session, nil
	}

	lease, info, err := s.probeSource(ctx, u, hash, fileIdx)
	if err != nil {
		return nil, err
	}

	session := &hlsSession{
		ID:         newSessionToken(),
		key:        key,
		UserID:     u.ID,
		Hash:       hash,
		FileIdx:    fileIdx,
		Profile:    prof,
		Info:       info,
		LeaseID:    lease.ID,
		lastAccess: time.Now(),
	}

	if err := session.start(ctx, s, 0, 0); err != nil {
		s.sourceRegistry.Release(lease.ID)
		return nil, err
	}

	// Прежние сессии того же файла с другими настройками больше не нужны:
	// пользователь сменил параметры, а старая сессия только занимает место
	s.dropOtherSessions(u.ID, hash, fileIdx, key)

	s.registry().Add(key, session)
	log.Infof("[Web] HLS session %s started for %s/%d", session.ID, hash, fileIdx)
	return session, nil
}

// dropOtherSessions останавливает сессии того же файла с другими настройками.
func (s *Server) dropOtherSessions(userID, hash string, fileIdx int, keepKey string) {
	for _, other := range s.registry().List() {
		if other.sessionKey() == keepKey {
			continue
		}
		if other.UserID == userID && other.Hash == hash && other.FileIdx == fileIdx {
			log.Debugf("[Web] Session %s replaced by new settings", other.ID)
			s.stopHLSSession(other)
		}
	}
}

// profileFingerprint возвращает отпечаток настроек профиля.
// Ключ сессии строится по нему целиком, а не по отдельным полям:
// любые различия в параметрах запроса дают отдельную сессию,
// и плеер получает поток ровно с запрошенными настройками.
func profileFingerprint(prof profile.Profile) string {
	// Поля времени и владельца не влияют на результат кодирования
	prof.CreatedAt = time.Time{}
	prof.UpdatedAt = time.Time{}
	prof.Name = ""
	prof.OwnerID = ""
	prof.IsDefault = false

	data, err := json.Marshal(prof)
	if err != nil {
		// Отпечаток нужен только для сравнения, поэтому при ошибке
		// достаточно отличить сессии по идентификатору профиля
		return prof.ID
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}

// hlsKey собирает ключ сессии: файл, пользователь и отпечаток настроек.
func hlsKey(hash string, fileIdx int, userID, fingerprint string) string {
	return fmt.Sprintf("%s|%d|%s|%s", hash, fileIdx, userID, fingerprint)
}

// sessionKey возвращает ключ сессии для этого экземпляра.
func (h *hlsSession) sessionKey() string {
	return h.key
}

// stopHLSSession останавливает сессию и освобождает её ресурсы.
func (s *Server) stopHLSSession(session *hlsSession) {
	if session == nil {
		return
	}

	s.registry().Remove(session.sessionKey())
	session.stop(s)
}

// StopAllHLS останавливает все сессии сегментирования.
func (s *Server) StopAllHLS() {
	for _, session := range s.registry().List() {
		s.stopHLSSession(session)
	}
}

// startHLSJanitor убирает сессии, к которым давно не обращались.
func (s *Server) startHLSJanitor() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			for _, session := range s.registry().List() {
				if session.idleFor() > hlsIdleTimeout {
					log.Infof("[Web] HLS session %s stopped after inactivity", session.ID)
					s.stopHLSSession(session)
				}
			}
		}
	}()
}

// markServed отмечает сегмент, отданный плееру: по этому счётчику
// ограничитель понимает, насколько процесс убежал вперёд.
func (h *hlsSession) markServed() {
	h.mu.Lock()
	instance := h.job
	h.mu.Unlock()

	if instance != nil {
		instance.AddServedSegment()
	}
}

// touch отмечает обращение плеера к сессии.
func (h *hlsSession) touch() {
	h.mu.Lock()
	h.lastAccess = time.Now()
	h.mu.Unlock()
}

// idleFor возвращает время с последнего обращения.
func (h *hlsSession) idleFor() time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	return time.Since(h.lastAccess)
}

// stop останавливает процесс, освобождает доступ и удаляет файлы сессии.
func (h *hlsSession) stop(s *Server) {
	h.mu.Lock()
	if h.stopped {
		h.mu.Unlock()
		return
	}
	h.stopped = true
	instance := h.job
	dir := h.workDir(s)
	h.mu.Unlock()

	if instance != nil {
		instance.Stop(s.stopTimeout())
	}
	s.sourceRegistry.Release(h.LeaseID)
	_ = os.RemoveAll(dir)
}

// done возвращает канал завершения текущего процесса.
func (h *hlsSession) done() <-chan struct{} {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.job == nil {
		return nil
	}
	return h.job.Done()
}

// cleanup удаляет сегменты, выпавшие из окна плейлиста.
func (h *hlsSession) cleanup(s *Server) {
	entries, err := os.ReadDir(h.workDir(s))
	if err != nil {
		return
	}

	var indexes []int
	for _, e := range entries {
		if e.IsDir() || !isSegmentName(e.Name()) {
			continue
		}
		if num, ok := hls.SegmentIndex(e.Name()); ok {
			indexes = append(indexes, num)
		}
	}
	tail := s.segmentTail()
	if len(indexes) <= tail {
		return
	}

	max := indexes[0]
	for _, num := range indexes {
		if num > max {
			max = num
		}
	}
	removeSegmentsBefore(h.workDir(s), max-tail+1)
}

// workDir возвращает каталог сессии.
func (h *hlsSession) workDir(s *Server) string {
	return s.transcoder.SessionDir(h.ID)
}

// playlistPath возвращает путь плейлиста сессии.
func (h *hlsSession) playlistPath(s *Server) string {
	return filepath.Join(h.workDir(s), "index.m3u8")
}

// segmentName собирает имя сегмента по номеру.
func (h *hlsSession) segmentName(index int) string {
	if h.Profile.HLS.SegmentType == "fmp4" {
		return fmt.Sprintf("seg%d.m4s", index)
	}
	return fmt.Sprintf("seg%d.ts", index)
}

// initSegmentName возвращает имя начального сегмента fmp4.
func (h *hlsSession) initSegmentName() string {
	if h.Profile.HLS.InitFilename != "" {
		return h.Profile.HLS.InitFilename
	}
	return args.DefaultInitSegment
}

// segmentLength возвращает длительность сегмента профиля.
func (h *hlsSession) segmentLength() int {
	if h.Profile.HLS.SegmentLength > 0 {
		return h.Profile.HLS.SegmentLength
	}
	return args.DefaultHLSOptions().SegmentLength
}

// bandwidth оценивает битрейт результата для главного плейлиста.
func (h *hlsSession) bandwidth() int {
	if h.Profile.Video.BitrateKbps > 0 {
		return h.Profile.Video.BitrateKbps * 1000
	}
	if h.Info != nil && h.Info.Bitrate > 0 {
		return int(h.Info.Bitrate)
	}
	return 2000000
}

// hlsRegistry хранит активные сессии сегментирования.
type hlsRegistry struct {
	mu       sync.RWMutex
	sessions map[string]*hlsSession
}

func newHLSRegistry() *hlsRegistry {
	return &hlsRegistry{sessions: make(map[string]*hlsSession)}
}

// registry возвращает реестр HLS, создавая его при первом обращении.
func (s *Server) registry() *hlsRegistry {
	s.hlsOnce.Do(func() {
		if s.hlsRegistry == nil {
			s.hlsRegistry = newHLSRegistry()
		}
	})
	return s.hlsRegistry
}

// Add регистрирует сессию под ключом.
func (r *hlsRegistry) Add(key string, s *hlsSession) {
	r.mu.Lock()
	r.sessions[key] = s
	r.mu.Unlock()
}

// ByKey возвращает активную сессию по ключу.
func (r *hlsRegistry) ByKey(key string) (*hlsSession, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.sessions[key]
	if !ok || s.isStopped() {
		return nil, false
	}
	return s, true
}

// ByID возвращает сессию по идентификатору.
func (r *hlsRegistry) ByID(id string) (*hlsSession, bool) {
	if id == "" {
		return nil, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, s := range r.sessions {
		if s.ID == id && !s.isStopped() {
			return s, true
		}
	}
	return nil, false
}

// Remove убирает сессию из реестра.
func (r *hlsRegistry) Remove(key string) {
	r.mu.Lock()
	delete(r.sessions, key)
	r.mu.Unlock()
}

// List возвращает все зарегистрированные сессии.
func (r *hlsRegistry) List() []*hlsSession {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*hlsSession, 0, len(r.sessions))
	for _, s := range r.sessions {
		out = append(out, s)
	}
	return out
}

// isStopped сообщает, остановлена ли сессия.
func (h *hlsSession) isStopped() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.stopped
}

// waitForFile ждёт появления файла.
func waitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("file did not appear in time")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// segmentTail возвращает число сегментов, остающихся на диске.
func (s *Server) segmentTail() int {
	if s.cfg != nil && s.cfg.FFmpeg.CacheSegments > 0 {
		return s.cfg.FFmpeg.CacheSegments
	}
	// Значение по умолчанию живёт рядом с остальными настройками транскодирования
	return ffmpeg.DefaultCacheSegments
}

// hlsWaitTimeout возвращает время ожидания готовности сегмента.
func (s *Server) hlsWaitTimeout() time.Duration {
	return 60 * time.Second
}

// removeSegments удаляет все сегменты в каталоге.
func removeSegments(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !isSegmentName(e.Name()) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, e.Name()))
	}
}

// removeSegmentsBefore удаляет сегменты с номером меньше указанного.
func removeSegmentsBefore(dir string, index int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !isSegmentName(e.Name()) {
			continue
		}
		num, ok := hls.SegmentIndex(e.Name())
		if !ok || num >= index {
			continue
		}
		_ = os.Remove(filepath.Join(dir, e.Name()))
	}
}

// isSegmentName проверяет, что имя похоже на сегмент.
// Начальный сегмент fmp4 тоже считается сегментом: плеер запрашивает его отдельно.
func isSegmentName(name string) bool {
	if strings.HasPrefix(name, ".") {
		return false
	}
	if name == args.DefaultInitSegment {
		return true
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".ts", ".m4s":
		return true
	}
	return false
}

// codecsFor собирает строку кодеков для главного плейлиста.
// Кодеки берутся из профиля: источник может быть в другом формате,
// а плеер выбирает вариант по объявленным кодекам.
func codecsFor(prof profile.Profile) string {
	var codecs []string

	if codec := prof.Video.Codec; codec != "" && codec != "none" {
		codecs = append(codecs, videoCodecString(codec))
	}
	if codec := prof.Audio.Codec; codec != "" && codec != "none" {
		codecs = append(codecs, audioCodecString(codec))
	}
	return strings.Join(codecs, ",")
}

// videoCodecString переводит имя видеокодека в строку RFC 6381.
func videoCodecString(codec string) string {
	switch codec {
	case "h264":
		return "avc1.42e01e"
	case "hevc":
		return "hvc1.1.6.L93.B0"
	case "av1":
		return "av01.0.04M.08"
	case "vp9":
		return "vp09.00.10.08"
	default:
		return codec
	}
}

// audioCodecString переводит имя аудиокодека в строку RFC 6381.
func audioCodecString(codec string) string {
	switch codec {
	case "aac":
		return "mp4a.40.2"
	case "ac3":
		return "ac-3"
	case "eac3":
		return "ec-3"
	case "opus":
		return "opus"
	default:
		return codec
	}
}

// newSessionToken создаёт непредсказуемый идентификатор сессии.
func newSessionToken() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
