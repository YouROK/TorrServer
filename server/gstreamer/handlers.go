package gstreamer

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

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

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Service) heartbeat(c *gin.Context) {
	hash := c.Param("hash")
	if s.Get(hash) == nil {
		c.Status(http.StatusNotFound)
		return
	}

	touchTorrent(hash)
	c.Status(http.StatusOK)
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
		if errors.Is(err, context.DeadlineExceeded) {
			c.String(http.StatusGatewayTimeout, err.Error())
			return
		}
		c.String(http.StatusBadGateway, err.Error())
		return
	}

	c.JSON(http.StatusOK, probe)
}

func (s *Service) master(c *gin.Context) {
	noCache(c)

	hash := c.Param("hash")
	fileID := firstNonEmpty(c.Query("index"), c.Query("id"), c.Query("fileID"))
	audio := parseQueryInt(c, "audio", 0)

	task, err := s.GetOrAdd(hash, fileID, audio)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			c.String(http.StatusGatewayTimeout, err.Error())
			return
		}
		c.String(http.StatusBadGateway, err.Error())
		return
	}

	duration := task.Probe.DurationSeconds()
	if duration <= 0 {
		duration = 200 * 60
	}

	segmentSeconds := task.Config.SegmentSeconds
	count := duration / segmentSeconds
	startIndex := startSegmentIndex(parseQueryInt(c, "seconds", 0), segmentSeconds, count)
	startSeconds := startIndex * segmentSeconds

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

	c.Data(http.StatusOK, "application/vnd.apple.mpegurl; charset=utf-8", []byte(playlist.String()))
}

func (s *Service) initMP4(c *gin.Context) {
	noCache(c)

	task := s.Get(c.Param("hash"))
	if task == nil {
		c.Status(http.StatusNotFound)
		return
	}

	audio := parseQueryInt(c, "audio", task.Audio)
	startIndex := startSegmentIndex(parseQueryInt(c, "seconds", 0), task.Config.SegmentSeconds, task.Probe.DurationSeconds()/task.Config.SegmentSeconds)
	if err := task.EnsureInit(c.Request.Context(), audio, startIndex); err != nil {
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	if err := task.WithInitMP4(func(init []byte) error {
		c.Header("Content-Length", strconv.Itoa(len(init)))
		c.Data(http.StatusOK, "video/mp4", init)
		return nil
	}); err != nil {
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}
}

func (s *Service) segment(c *gin.Context) {
	noCache(c)

	task := s.Get(c.Param("hash"))
	if task == nil {
		c.Status(http.StatusNotFound)
		return
	}

	index, err := parseSegmentIndex(c.Param("segment"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	audio := parseQueryInt(c, "audio", task.Audio)
	if !task.hasInitMP4() {
		if err := task.EnsureInit(c.Request.Context(), audio, index); err != nil {
			c.AbortWithError(http.StatusBadGateway, err)
			return
		}
	}

	err = task.WithSegment(c.Request.Context(), index, audio, func(seg Segment) error {
		if seg.Empty() {
			return ErrSegmentNotReady
		}
		return writeSegment(c, seg)
	})
	if err != nil {
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}
}

func writeSegment(c *gin.Context, seg Segment) error {
	totalLength := int64(seg.Len())

	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader == "" {
		c.Header("Content-Length", strconv.FormatInt(totalLength, 10))
		return seg.WriteTo(c.Writer)
	}

	start, end, ok := parseSingleRange(rangeHeader, totalLength)
	if !ok {
		c.Header("Content-Range", "bytes */"+strconv.FormatInt(totalLength, 10))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return nil
	}

	length := end - start + 1
	c.Status(http.StatusPartialContent)
	c.Header("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(totalLength, 10))
	c.Header("Content-Length", strconv.FormatInt(length, 10))

	return seg.WriteRange(c.Writer, start, length)
}

func parseSingleRange(header string, totalLength int64) (int64, int64, bool) {
	const prefix = "bytes="
	if totalLength <= 0 || !strings.HasPrefix(header, prefix) {
		return 0, 0, false
	}

	spec := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if spec == "" || strings.Contains(spec, ",") {
		return 0, 0, false
	}

	left, right, ok := strings.Cut(spec, "-")
	if !ok {
		return 0, 0, false
	}

	var start int64
	var end int64

	if left != "" {
		parsedStart, err := strconv.ParseInt(left, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		start = parsedStart

		if right == "" {
			end = totalLength - 1
		} else {
			parsedEnd, err := strconv.ParseInt(right, 10, 64)
			if err != nil {
				return 0, 0, false
			}
			end = parsedEnd
		}
	} else {
		suffixLength, err := strconv.ParseInt(right, 10, 64)
		if err != nil || suffixLength <= 0 {
			return 0, 0, false
		}
		if suffixLength > totalLength {
			suffixLength = totalLength
		}
		start = totalLength - suffixLength
		end = totalLength - 1
	}

	if start < 0 || end < start || start >= totalLength {
		return 0, 0, false
	}
	if end >= totalLength {
		end = totalLength - 1
	}

	return start, end, true
}

func parseSegmentIndex(value string) (int, error) {
	value = strings.TrimPrefix(value, "/")
	value = strings.TrimSuffix(value, ".m4s")
	if value == "" || strings.Contains(value, "/") {
		return 0, errors.New("invalid segment index")
	}
	index, err := strconv.Atoi(value)
	if err != nil || index < 0 {
		return 0, errors.New("invalid segment index")
	}
	return index, nil
}

func parseQueryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func startSegmentIndex(seconds int, segmentSeconds int, count int) int {
	if seconds <= 0 || segmentSeconds <= 0 {
		return 0
	}

	index := seconds / segmentSeconds
	if index < 0 {
		return 0
	}
	if count > 0 && index > count {
		return count
	}
	return index
}

func noCache(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}
