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
	route.GET("/gst/:hash/init.mp4", s.initMP4)
	route.GET("/gst/:hash/seg/*segment", s.segment)
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

	segmentSeconds := task.Config.SegmentSeconds
	rawDuration := task.Probe.DurationSeconds()
	duration, fallbackDuration := playlistDurationSeconds(rawDuration, segmentSeconds)
	if fallbackDuration {
		gstDebugf("HTTP master duration fallback task=%s file=%s audio=%d seconds=%d rawDuration=%d segmentSeconds=%d playlistDuration=%d", task.ID, task.FileID, audio, seconds, rawDuration, segmentSeconds, duration)
	}

	count := duration / segmentSeconds
	startIndex := startSegmentIndex(seconds, segmentSeconds, count)
	startSeconds := startIndex * segmentSeconds

	playlist := buildMasterPlaylist(segmentSeconds, startIndex, startSeconds, count, audio)
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl; charset=utf-8", []byte(playlist))
	gstDebugf("HTTP master completed task=%s file=%s audio=%d startIndex=%d seconds=%d status=%d bytes=%d duration=%s", task.ID, task.FileID, audio, startIndex, seconds, c.Writer.Status(), len(playlist), time.Since(started))
}

func playlistDurationSeconds(rawDuration int, segmentSeconds int) (int, bool) {
	if rawDuration <= 0 || rawDuration < segmentSeconds {
		return unknownPlaylistDurationSeconds, true
	}
	return rawDuration, false
}

func buildMasterPlaylist(segmentSeconds int, startIndex int, startSeconds int, count int, audio int) string {
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
