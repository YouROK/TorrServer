package gstreamer

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Service) SetupRoute(route gin.IRouter) {
	route.GET("/gst/remove", s.remove)
	route.GET("/gst/echo", s.echo)
	route.GET("/gst/:hash/heartbeat", s.heartbeat)
	route.GET("/gst/:hash/probe", s.probe)
	route.GET("/gst/:hash/master.m3u8", s.master)
	route.GET("/gst/:hash/video.m3u8", s.video)
	route.GET("/gst/:hash/init.mp4", s.initMP4)
	route.GET("/gst/:hash/seg/*segment", s.segment)
	route.GET("/gst/:hash/subs/:subtitle", s.subtitlePlaylist)
	route.GET("/gst/:hash/subs/:subtitle/:segment", s.subtitleSegment)
}

func (s *Service) remove(c *gin.Context) {
	id := firstNonEmpty(c.Query("hash"), c.Query("id"))
	if id == "" {
		c.AbortWithError(http.StatusBadRequest, ErrInvalidIdentifier)
		return
	}

	if !s.TryRemove(id) {
		c.Status(http.StatusNotFound)
		return
	}

	dropTorrentForGStreamer(id)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Service) heartbeat(c *gin.Context) {
	hash := c.Param("hash")
	if s.Get(hash) == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, torrentHeartbeatState(hash))
}

func (s *Service) probe(c *gin.Context) {
	noCache(c)

	hash := c.Param("hash")
	fileID := firstNonEmpty(c.Query("index"), c.Query("id"), c.Query("fileID"))
	if fileID == "" {
		c.AbortWithError(http.StatusBadRequest, ErrBadSource)
		return
	}

	probe, err := s.Probe(hash, fileID)
	if err != nil {
		abortWithSourceError(c, err)
		return
	}

	c.JSON(http.StatusOK, probe)
}

func (s *Service) master(c *gin.Context) {
	started := time.Now()
	noCache(c)

	hash := c.Param("hash")
	fileID := firstNonEmpty(c.Query("index"), c.Query("id"), c.Query("fileID"))
	audio := parseQueryInt(c, "audio", 0)
	seconds := parseQueryInt(c, "seconds", 0)
	gstDebugf("HTTP master start hash=%s file=%s audio=%d seconds=%d range=%q", hash, fileID, audio, seconds, c.GetHeader("Range"))

	task, err := s.GetOrAdd(hash, fileID, audio)
	if err != nil {
		logHTTPError("HTTP master", hash, fileID, audio, -1, err, started)
		abortWithSourceError(c, err)
		return
	}

	if len(subtitleTracksForPlaylist(task.Probe)) == 0 {
		segmentSeconds := task.Config.SegmentSeconds
		rawDuration := task.Probe.DurationSeconds()
		duration, fallbackDuration := playlistDurationSeconds(rawDuration, segmentSeconds)
		if fallbackDuration {
			gstDebugf("HTTP master duration fallback task=%s file=%s audio=%d seconds=%d rawDuration=%d segmentSeconds=%d playlistDuration=%d", task.ID, task.FileID, audio, seconds, rawDuration, segmentSeconds, duration)
		}

		count := duration / segmentSeconds
		startIndex := startSegmentIndex(seconds, segmentSeconds, count)
		startSeconds := startIndex * segmentSeconds

		playlist := buildMediaPlaylist(segmentSeconds, startIndex, startSeconds, count, audio)
		c.Data(http.StatusOK, "application/vnd.apple.mpegurl; charset=utf-8", []byte(playlist))
		gstDebugf("HTTP master media completed task=%s file=%s audio=%d startIndex=%d seconds=%d status=%d bytes=%d duration=%s", task.ID, task.FileID, audio, startIndex, seconds, c.Writer.Status(), len(playlist), time.Since(started))
		return
	}

	playlist := buildMasterPlaylist(hash, task.Probe, task.Config, audio, seconds)
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl; charset=utf-8", []byte(playlist))
	gstDebugf("HTTP master completed task=%s file=%s audio=%d seconds=%d status=%d bytes=%d duration=%s", task.ID, task.FileID, audio, seconds, c.Writer.Status(), len(playlist), time.Since(started))
}

func playlistDurationSeconds(rawDuration int, segmentSeconds int) (int, bool) {
	if rawDuration <= 0 || rawDuration < segmentSeconds {
		return unknownPlaylistDurationSeconds, true
	}
	return rawDuration, false
}

func buildMediaPlaylist(segmentSeconds int, startIndex int, startSeconds int, count int, audio int) string {
	var playlist strings.Builder
	playlist.WriteString("#EXTM3U\n")
	playlist.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	playlist.WriteString("#EXT-X-VERSION:7\n")
	playlist.WriteString("#EXT-X-TARGETDURATION:")
	playlist.WriteString(strconv.Itoa(segmentSeconds))
	playlist.WriteByte('\n')
	playlist.WriteString("#EXT-X-MEDIA-SEQUENCE:")
	playlist.WriteString(strconv.Itoa(startIndex))
	playlist.WriteByte('\n')
	playlist.WriteString("#EXT-X-MAP:URI=\"init.mp4?audio=")
	playlist.WriteString(strconv.Itoa(audio))
	if startSeconds > 0 {
		playlist.WriteString("&seconds=")
		playlist.WriteString(strconv.Itoa(startSeconds))
	}
	playlist.WriteString("\"\n")

	for i := startIndex; i < count; i++ {
		playlist.WriteString("#EXTINF:")
		playlist.WriteString(strconv.Itoa(segmentSeconds))
		playlist.WriteString(".00,\n")
		playlist.WriteString("seg/")
		playlist.WriteString(strconv.Itoa(i))
		playlist.WriteString(".m4s\n")
	}

	playlist.WriteString("#EXT-X-ENDLIST\n")
	return playlist.String()
}

func (s *Service) initMP4(c *gin.Context) {
	started := time.Now()
	noCache(c)

	hash := c.Param("hash")
	task := s.Get(hash)
	if task == nil {
		gstDebugf("HTTP init start hash=%s file= audio=0 seconds=0 startIndex=0 range=%q", hash, c.GetHeader("Range"))
		gstErrorf("HTTP init failed hash=%s file= audio=0 segment=-1 status=%d err=%v duration=%s", hash, http.StatusNotFound, ErrTaskNotFound, time.Since(started))
		c.Status(http.StatusNotFound)
		return
	}

	audio := parseQueryInt(c, "audio", task.Audio)
	seconds := parseQueryInt(c, "seconds", 0)
	segmentSeconds := task.Config.SegmentSeconds
	rawDuration := task.Probe.DurationSeconds()
	duration, fallbackDuration := playlistDurationSeconds(rawDuration, segmentSeconds)
	if fallbackDuration {
		gstDebugf("HTTP init duration fallback task=%s file=%s audio=%d seconds=%d rawDuration=%d segmentSeconds=%d playlistDuration=%d", task.ID, task.FileID, audio, seconds, rawDuration, segmentSeconds, duration)
	}
	startIndex := startSegmentIndex(seconds, segmentSeconds, duration/segmentSeconds)
	gstDebugf("HTTP init start task=%s file=%s audio=%d seconds=%d startIndex=%d range=%q", task.ID, task.FileID, audio, seconds, startIndex, c.GetHeader("Range"))
	if err := task.EnsureInit(c.Request.Context(), audio, startIndex); err != nil {
		logHTTPError("HTTP init EnsureInit", task.ID, task.FileID, audio, -1, err, started)
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}
	gstDebugf("HTTP init EnsureInit completed task=%s file=%s audio=%d startIndex=%d elapsed=%s", task.ID, task.FileID, audio, startIndex, time.Since(started))

	var sent int
	if err := task.WithInitMP4(func(init []byte) error {
		sent = len(init)
		c.Header("Content-Length", strconv.Itoa(len(init)))
		c.Data(http.StatusOK, "video/mp4", init)
		return nil
	}); err != nil {
		logHTTPError("HTTP init write", task.ID, task.FileID, audio, -1, err, started)
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}
	gstDebugf("HTTP init completed task=%s file=%s audio=%d startIndex=%d status=%d bytes=%d duration=%s", task.ID, task.FileID, audio, startIndex, c.Writer.Status(), sent, time.Since(started))
}

func (s *Service) segment(c *gin.Context) {
	started := time.Now()
	noCache(c)

	hash := c.Param("hash")
	task := s.Get(hash)
	if task == nil {
		gstDebugf("HTTP segment start hash=%s file= audio=0 segment=-1 range=%q", hash, c.GetHeader("Range"))
		gstErrorf("HTTP segment failed hash=%s file= audio=0 segment=-1 status=%d err=%v duration=%s", hash, http.StatusNotFound, ErrTaskNotFound, time.Since(started))
		c.Status(http.StatusNotFound)
		return
	}

	index, err := parseSegmentIndex(c.Param("segment"))
	if err != nil {
		logHTTPError("HTTP segment parse", task.ID, task.FileID, task.Audio, -1, err, started)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	audio := parseQueryInt(c, "audio", task.Audio)
	gstDebugf("HTTP segment start task=%s file=%s audio=%d segment=%d range=%q uri=%q ua=%q", task.ID, task.FileID, audio, index, c.GetHeader("Range"), c.Request.RequestURI, c.GetHeader("User-Agent"))
	if !task.hasInitMP4() {
		if err := task.EnsureInit(c.Request.Context(), audio, index); err != nil {
			logHTTPError("HTTP segment EnsureInit", task.ID, task.FileID, audio, index, err, started)
			c.AbortWithError(http.StatusBadGateway, err)
			return
		}
		gstDebugf("HTTP segment EnsureInit completed task=%s file=%s audio=%d segment=%d elapsed=%s", task.ID, task.FileID, audio, index, time.Since(started))
	}

	var sent int64
	err = task.WithSegment(c.Request.Context(), index, audio, func(seg Segment) error {
		if seg.Empty() {
			return ErrSegmentNotReady
		}
		gstDebugf("HTTP segment GetSegment completed task=%s file=%s audio=%d segment=%d size=%d elapsed=%s", task.ID, task.FileID, audio, index, seg.Len(), time.Since(started))
		var writeErr error
		sent, writeErr = writeSegment(c, seg)
		return writeErr
	})
	if err != nil {
		logHTTPError("HTTP segment GetSegment", task.ID, task.FileID, audio, index, err, started)
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}
	gstDebugf("HTTP segment response written task=%s file=%s audio=%d segment=%d status=%d bytes=%d duration=%s", task.ID, task.FileID, audio, index, c.Writer.Status(), sent, time.Since(started))
}

const unknownPlaylistDurationSeconds = 6*60*60 + 6*60 + 6

func writeSegment(c *gin.Context, seg Segment) (int64, error) {
	totalLength := int64(seg.Len())

	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader == "" {
		c.Header("Content-Length", strconv.FormatInt(totalLength, 10))
		return totalLength, seg.WriteTo(c.Writer)
	}

	start, end, ok := parseSingleRange(rangeHeader, totalLength)
	if !ok {
		c.Header("Content-Range", "bytes */"+strconv.FormatInt(totalLength, 10))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return 0, nil
	}

	length := end - start + 1
	c.Status(http.StatusPartialContent)
	c.Header("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(totalLength, 10))
	c.Header("Content-Length", strconv.FormatInt(length, 10))

	return length, seg.WriteRange(c.Writer, start, length)
}

func logHTTPError(prefix string, task string, fileID string, audio int, segment int, err error, started time.Time) {
	if errors.Is(err, context.Canceled) {
		gstErrorf("%s context canceled task=%s file=%s audio=%d segment=%d err=%v duration=%s", prefix, task, fileID, audio, segment, err, time.Since(started))
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		gstErrorf("%s deadline exceeded task=%s file=%s audio=%d segment=%d err=%v duration=%s", prefix, task, fileID, audio, segment, err, time.Since(started))
		return
	}
	gstErrorf("%s failed task=%s file=%s audio=%d segment=%d err=%v duration=%s", prefix, task, fileID, audio, segment, err, time.Since(started))
}

func parseSingleRange(header string, totalLength int64) (int64, int64, bool) {
	const prefix = "bytes="
	if totalLength <= 0 || !strings.HasPrefix(header, prefix) {
		return 0, 0, false
	}

	rangeText := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if strings.Contains(rangeText, ",") {
		return 0, 0, false
	}

	startText, endText, ok := strings.Cut(rangeText, "-")
	if !ok {
		return 0, 0, false
	}

	startText = strings.TrimSpace(startText)
	endText = strings.TrimSpace(endText)
	if startText == "" {
		suffix, err := strconv.ParseInt(endText, 10, 64)
		if err != nil || suffix <= 0 {
			return 0, 0, false
		}
		if suffix > totalLength {
			suffix = totalLength
		}
		return totalLength - suffix, totalLength - 1, true
	}

	start, err := strconv.ParseInt(startText, 10, 64)
	if err != nil || start < 0 || start >= totalLength {
		return 0, 0, false
	}

	end := totalLength - 1
	if endText != "" {
		parsedEnd, err := strconv.ParseInt(endText, 10, 64)
		if err != nil || parsedEnd < start {
			return 0, 0, false
		}
		if parsedEnd < end {
			end = parsedEnd
		}
	}

	return start, end, true
}

func (s *Service) video(c *gin.Context) {
	started := time.Now()
	noCache(c)

	hash := c.Param("hash")
	fileID := firstNonEmpty(c.Query("index"), c.Query("id"), c.Query("fileID"))
	audio := parseQueryInt(c, "audio", 0)
	seconds := parseQueryInt(c, "seconds", 0)

	task := s.Get(hash)
	var err error
	if task == nil && fileID != "" {
		task, err = s.GetOrAdd(hash, fileID, audio)
	}
	if err != nil {
		logHTTPError("HTTP video", hash, fileID, audio, -1, err, started)
		abortWithSourceError(c, err)
		return
	}
	if task == nil {
		c.Status(http.StatusNotFound)
		return
	}

	segmentSeconds := task.Config.normalized().SegmentSeconds
	rawDuration := task.Probe.DurationSeconds()
	duration, fallbackDuration := playlistDurationSeconds(rawDuration, segmentSeconds)
	if fallbackDuration {
		gstDebugf("HTTP video duration fallback task=%s file=%s audio=%d seconds=%d rawDuration=%d segmentSeconds=%d playlistDuration=%d", task.ID, task.FileID, audio, seconds, rawDuration, segmentSeconds, duration)
	}

	count := duration / segmentSeconds
	startIndex := startSegmentIndex(seconds, segmentSeconds, count)
	startSeconds := startIndex * segmentSeconds

	playlist := buildMediaPlaylist(segmentSeconds, startIndex, startSeconds, count, audio)
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl; charset=utf-8", []byte(playlist))
	gstDebugf("HTTP video completed task=%s file=%s audio=%d startIndex=%d seconds=%d status=%d bytes=%d duration=%s", task.ID, task.FileID, audio, startIndex, seconds, c.Writer.Status(), len(playlist), time.Since(started))
}

func (s *Service) subtitlePlaylist(c *gin.Context) {
	noCache(c)

	task := s.Get(c.Param("hash"))
	if task == nil {
		c.Status(http.StatusNotFound)
		return
	}

	subtitle, err := parseSubtitleIndex(c.Param("subtitle"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	track := task.Probe.SubtitleTrack(subtitle)
	if track == nil || !track.IsSupportedSubtitle() {
		c.Status(http.StatusNotFound)
		return
	}

	segmentSeconds := task.Config.normalized().SegmentSeconds
	duration, _ := playlistDurationSeconds(task.Probe.DurationSeconds(), segmentSeconds)
	playlist := buildSubtitlePlaylist(segmentSeconds, duration/segmentSeconds, subtitle)
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl; charset=utf-8", []byte(playlist))
}

func (s *Service) subtitleSegment(c *gin.Context) {
	started := time.Now()
	noCache(c)

	task := s.Get(c.Param("hash"))
	if task == nil {
		c.Status(http.StatusNotFound)
		return
	}

	subtitle, err := parseSubtitleIndex(c.Param("subtitle"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	segment, err := parseSubtitleSegmentIndex(c.Param("segment"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	data, err := task.SubtitleSegment(c.Request.Context(), subtitle, segment)
	if err != nil {
		logHTTPError("HTTP subtitle segment", task.ID, task.FileID, task.Audio, segment, err, started)
		if errors.Is(err, ErrSubtitleNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrUnsupportedSubtitle) {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		data = emptyWebVTTSegment()
		c.Data(http.StatusOK, "text/vtt; charset=utf-8", data)
		gstDebugf("HTTP subtitle segment fallback empty task=%s file=%s subtitle=%d segment=%d status=%d bytes=%d duration=%s", task.ID, task.FileID, subtitle, segment, c.Writer.Status(), len(data), time.Since(started))
		return
	}

	c.Data(http.StatusOK, "text/vtt; charset=utf-8", data)
	gstDebugf("HTTP subtitle segment completed task=%s file=%s subtitle=%d segment=%d status=%d bytes=%d duration=%s", task.ID, task.FileID, subtitle, segment, c.Writer.Status(), len(data), time.Since(started))
}

func buildMasterPlaylist(hash string, probe ProbeInfo, conf Config, audio int, seconds int) string {
	tracks := subtitleTracksForPlaylist(probe)
	bandwidth := conf.normalized().VideoBitrate*1000 + conf.normalized().AACBitrateKbps*1000
	if bandwidth <= 0 {
		bandwidth = 4_000_000
	}

	var playlist strings.Builder
	playlist.WriteString("#EXTM3U\n")
	playlist.WriteString("#EXT-X-VERSION:7\n")
	playlist.WriteByte('\n')

	for _, track := range tracks {
		playlist.WriteString("#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\",NAME=\"")
		playlist.WriteString(hlsQuote(subtitleTrackName(track)))
		playlist.WriteString("\",LANGUAGE=\"")
		playlist.WriteString(hlsQuote(subtitleTrackLanguage(track)))
		playlist.WriteString("\",DEFAULT=NO,AUTOSELECT=NO,FORCED=NO,URI=\"/gst/")
		playlist.WriteString(hlsQuote(hash))
		playlist.WriteString("/subs/")
		playlist.WriteString(strconv.Itoa(track.Index))
		playlist.WriteString(".m3u8\"\n")
	}
	if len(tracks) > 0 {
		playlist.WriteByte('\n')
	}

	playlist.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=")
	playlist.WriteString(strconv.Itoa(bandwidth))
	if video := probe.Video(); video != nil && video.Width > 0 && video.Height > 0 {
		playlist.WriteString(",RESOLUTION=")
		playlist.WriteString(strconv.Itoa(video.Width))
		playlist.WriteByte('x')
		playlist.WriteString(strconv.Itoa(video.Height))
	}
	if codecs := hlsMasterCodecs(probe, conf); codecs != "" {
		playlist.WriteString(",CODECS=\"")
		playlist.WriteString(codecs)
		playlist.WriteByte('"')
	}
	if len(tracks) > 0 {
		playlist.WriteString(",SUBTITLES=\"subs\"")
	}

	playlist.WriteByte('\n')
	playlist.WriteString("/gst/")
	playlist.WriteString(hlsQuote(hash))
	playlist.WriteString("/video.m3u8?audio=")
	playlist.WriteString(strconv.Itoa(audio))
	if seconds > 0 {
		playlist.WriteString("&seconds=")
		playlist.WriteString(strconv.Itoa(seconds))
	}
	playlist.WriteByte('\n')
	return playlist.String()
}

func hlsMasterCodecs(probe ProbeInfo, conf Config) string {
	conf = conf.normalized()
	videoIsH264 := probe.IsH264() ||
		(probe.IsH265() && conf.TranscodeH265) ||
		(probe.IsAV1() && conf.TranscodeAV1) ||
		(probe.IsVP9() && conf.TranscodeVP9)
	if videoIsH264 {
		return "avc1.640028,mp4a.40.2"
	}

	return ""
}
