//go:build gst

package gstreamer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
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
	route.GET("/gst/:hash/video.m3u8", s.videoPlaylist)
	route.GET("/gst/:hash/init.mp4", s.initMP4)
	route.GET("/gst/:hash/seg/*segment", s.segment)
	route.GET("/gst/:hash/subs/*subtitle", s.subtitle)
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
		gstSourceFailure(hash, fileID, 0, "probe request", err)
		abortWithSourceError(c, err)
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
		gstSourceFailure(hash, fileID, audio, "master task creation", err)
		abortWithSourceError(c, err)
		return
	}

	seconds := parseQueryInt(c, "seconds", 0)
	if err := task.EnsureInit(c.Request.Context(), audio, task.startIndexForSeconds(seconds)); err != nil {
		gstTaskFailure(task, "master init", err)
		abortWithRequestError(c, http.StatusBadGateway, err)
		return
	}

	playlist := buildVariantPlaylist(task, audio, seconds)
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl; charset=utf-8", []byte(playlist))
}

func (s *Service) videoPlaylist(c *gin.Context) {
	noCache(c)
	task := s.Get(c.Param("hash"))
	if task == nil {
		c.Status(http.StatusNotFound)
		return
	}
	audio := parseQueryInt(c, "audio", task.Audio)
	startIndex := task.startIndexForSeconds(parseQueryInt(c, "seconds", 0))
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl; charset=utf-8", []byte(buildTaskPlaylist(task, startIndex, audio)))
}

func buildVariantPlaylist(task *Task, audio int, seconds int) string {
	var playlist strings.Builder
	playlist.WriteString("#EXTM3U\n#EXT-X-VERSION:7\n\n")
	taskID := url.PathEscape(task.ID)
	hasSubtitles := false
	if task.Config.Subtitles {
		for _, track := range task.Probe.Tracks {
			if !supportedSubtitleTrack(track) {
				continue
			}
			hasSubtitles = true
			language := hlsQuoted(firstNonEmpty(track.Language, "und"))
			name := track.Title
			if name == "" {
				name = "Subtitle " + strconv.Itoa(track.Index)
			}
			playlist.WriteString("#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\",NAME=\"")
			playlist.WriteString(hlsQuoted(name))
			playlist.WriteString("\",LANGUAGE=\"")
			playlist.WriteString(language)
			playlist.WriteString("\",DEFAULT=NO,AUTOSELECT=YES,FORCED=NO,URI=\"/gst/")
			playlist.WriteString(taskID)
			playlist.WriteString("/subs/")
			playlist.WriteString(strconv.Itoa(track.Index))
			playlist.WriteString(".m3u8\"\n")
		}
		if hasSubtitles {
			playlist.WriteByte('\n')
		}
	}

	bandwidth, average := task.hlsBandwidth()
	variant, hasVariant := task.hlsVariant()
	playlist.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=")
	playlist.WriteString(strconv.FormatInt(bandwidth, 10))
	if average > 0 {
		playlist.WriteString(",AVERAGE-BANDWIDTH=")
		playlist.WriteString(strconv.FormatInt(average, 10))
	}
	if video := task.Probe.Video(); video != nil {
		width, height := video.Width, video.Height
		frameRate := 0.0
		codecs := task.hlsCodecs()
		videoRange := strings.ToUpper(video.VideoTransfer)
		if hasVariant {
			if variant.Width > 0 {
				width = variant.Width
			}
			if variant.Height > 0 {
				height = variant.Height
			}
			frameRate = variant.FrameRate
			if variant.Codecs != "" {
				codecs = variant.Codecs
			}
			if variant.VideoRange != "" {
				videoRange = strings.ToUpper(variant.VideoRange)
			}
		}
		if width > 0 && height > 0 {
			playlist.WriteString(",RESOLUTION=")
			playlist.WriteString(strconv.Itoa(width))
			playlist.WriteByte('x')
			playlist.WriteString(strconv.Itoa(height))
		}
		if frameRate <= 0 && video.FrameRateNum > 0 && video.FrameRateDen > 0 {
			frameRate = float64(video.FrameRateNum) / float64(video.FrameRateDen)
		}
		if frameRate > 0 {
			playlist.WriteString(",FRAME-RATE=")
			playlist.WriteString(strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", frameRate), "0"), "."))
		}
		if codecs != "" {
			playlist.WriteString(",CODECS=\"")
			playlist.WriteString(codecs)
			playlist.WriteByte('"')
		}
		if task.Config.HDRToSDR && video.IsHDRVideo() {
			videoRange = "SDR"
		}
		if videoRange == "PQ" || videoRange == "HLG" || videoRange == "SDR" {
			playlist.WriteString(",VIDEO-RANGE=")
			playlist.WriteString(videoRange)
		}
	}
	if hasSubtitles {
		playlist.WriteString(",SUBTITLES=\"subs\"")
	}
	playlist.WriteString("\n/gst/")
	playlist.WriteString(taskID)
	playlist.WriteString("/video.m3u8?audio=")
	playlist.WriteString(strconv.Itoa(audio))
	if seconds > 0 {
		playlist.WriteString("&seconds=")
		playlist.WriteString(strconv.Itoa(seconds))
	}
	playlist.WriteByte('\n')
	return playlist.String()
}

func (t *Task) hlsBandwidth() (int64, int64) {
	audio := int64(max(t.Config.AACBitrateKbps, 1) * 1000)
	if track := t.Probe.AudioTrack(t.Audio); effectiveAACChannels(t.Config, track) > 2 {
		audio *= 2
	}
	if videoIsTranscoded(t.Config, t.Probe) {
		average := int64(max(t.Config.VideoBitrate, 1)*1000) + audio
		return average * 125 / 100, average
	}
	if t.Probe.FileSize > 0 && t.Probe.DurationNS > 0 {
		durationSeconds := float64(t.Probe.DurationNS) / float64(time.Second)
		averageValue := float64(t.Probe.FileSize) * 8 / durationSeconds
		if !math.IsNaN(averageValue) && !math.IsInf(averageValue, 0) &&
			averageValue > 0 && averageValue <= float64(math.MaxInt64)/2 {
			average := int64(math.Ceil(averageValue))
			return average * 150 / 100, average
		}
	}
	fallback := int64(max(t.Config.VideoBitrate, 1)*1000) + audio
	return max(int64(4_000_000), fallback*150/100), 0
}

func (t *Task) hlsCodecs() string {
	videoCodec := ""
	if videoIsTranscoded(t.Config, t.Probe) {
		videoCodec = "avc1.4d401f"
	} else {
		switch {
		case t.Probe.IsH264():
			videoCodec = "avc1.4d401f"
		case t.Probe.IsH265():
			videoCodec = "hvc1"
		case t.Probe.IsAV1():
			videoCodec = "av01.0.08M.08"
		case t.Probe.IsVP9():
			videoCodec = "vp09.00.10.08"
		}
	}
	if t.Probe.HasAudio() {
		return strings.Trim(videoCodec+",mp4a.40.2", ",")
	}
	return videoCodec
}

func hlsQuoted(value string) string {
	value = strings.Map(func(char rune) rune {
		if char < 0x20 || char == 0x7f {
			return ' '
		}
		return char
	}, value)
	return strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "\"", "\\\"")
}

func buildTaskPlaylist(task *Task, startIndex int, audio int) string {
	segmentSeconds := max(task.Config.SegmentSeconds, 1)
	durationNS := task.Probe.DurationNS
	if durationNS <= 0 {
		durationNS = int64(200 * 60 * 1_000_000_000)
	}

	segmentDurationNS := int64(segmentSeconds) * 1_000_000_000
	count := int(1 + (durationNS-1)/segmentDurationNS)
	targetDuration := segmentSeconds
	if task.Cue != nil {
		count = len(task.Cue.Segments)
		targetDuration = int((task.Cue.MaxDurationNS + 1_000_000_000 - 1) / 1_000_000_000)
	}
	startIndex = min(max(startIndex, 0), count)

	var playlist strings.Builder
	playlist.WriteString("#EXTM3U\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-VERSION:7\n")
	playlist.WriteString("#EXT-X-TARGETDURATION:")
	playlist.WriteString(strconv.Itoa(max(targetDuration, 1)))
	playlist.WriteString("\n#EXT-X-MEDIA-SEQUENCE:")
	playlist.WriteString(strconv.Itoa(startIndex))
	playlist.WriteString("\n#EXT-X-MAP:URI=\"init.mp4?audio=")
	playlist.WriteString(strconv.Itoa(audio))
	if startIndex > 0 {
		playlist.WriteString("&seconds=")
		playlist.WriteString(strconv.Itoa(int(task.segmentStartNS(startIndex) / 1_000_000_000)))
	}
	playlist.WriteString("\"\n")

	segmentNS := uint64(segmentSeconds) * 1_000_000_000
	for i := startIndex; i < count; i++ {
		itemDurationNS := segmentNS
		if cue, ok := task.Cue.Segment(i); ok {
			itemDurationNS = cue.DurationNS()
		} else if i == count-1 {
			startNS := uint64(i) * segmentNS
			if uint64(durationNS) > startNS {
				itemDurationNS = uint64(durationNS) - startNS
			}
		}
		playlist.WriteString("#EXTINF:")
		playlist.WriteString(strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", float64(itemDurationNS)/1_000_000_000), "0"), "."))
		playlist.WriteString(",\nseg/")
		playlist.WriteString(strconv.Itoa(i))
		playlist.WriteString(".m4s\n")
	}
	playlist.WriteString("#EXT-X-ENDLIST\n")
	return playlist.String()
}

func (s *Service) initMP4(c *gin.Context) {
	noCache(c)

	task := s.Get(c.Param("hash"))
	if task == nil {
		c.Status(http.StatusNotFound)
		return
	}

	audio := parseQueryInt(c, "audio", task.Audio)
	startIndex := task.startIndexForSeconds(parseQueryInt(c, "seconds", 0))
	if err := task.EnsureInit(c.Request.Context(), audio, startIndex); err != nil {
		gstTaskFailure(task, "init.mp4 preparation", err)
		abortWithRequestError(c, http.StatusBadGateway, err)
		return
	}
	if c.Request.Context().Err() != nil {
		return
	}

	if err := task.WithInitMP4(func(init []byte) error {
		c.Header("Content-Length", strconv.Itoa(len(init)))
		c.Data(http.StatusOK, "video/mp4", init)
		return nil
	}); err != nil {
		gstTaskFailure(task, "init.mp4 response", err)
		abortWithRequestError(c, http.StatusBadGateway, err)
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
			gstTaskFailure(task, fmt.Sprintf("segment %d init", index), err)
			abortWithRequestError(c, http.StatusBadGateway, err)
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
		gstTaskFailure(task, fmt.Sprintf("segment %d response", index), err)
		abortWithRequestError(c, http.StatusBadGateway, err)
		return
	}
}

func (s *Service) subtitle(c *gin.Context) {
	noCache(c)
	task := s.Get(c.Param("hash"))
	if task == nil {
		c.Status(http.StatusNotFound)
		return
	}
	path := strings.TrimPrefix(c.Param("subtitle"), "/")
	if strings.HasSuffix(path, ".m3u8") && !strings.Contains(strings.TrimSuffix(path, ".m3u8"), "/") {
		trackIndex, err := strconv.Atoi(strings.TrimSuffix(path, ".m3u8"))
		if err != nil || trackIndex < 0 {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Data(http.StatusOK, "application/vnd.apple.mpegurl; charset=utf-8", []byte(buildSubtitlePlaylist(task, trackIndex)))
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) != 2 || !strings.HasSuffix(parts[1], ".vtt") {
		c.Status(http.StatusBadRequest)
		return
	}
	trackIndex, trackErr := strconv.Atoi(parts[0])
	segmentIndex, segmentErr := strconv.Atoi(strings.TrimSuffix(parts[1], ".vtt"))
	if trackErr != nil || segmentErr != nil || trackIndex < 0 || segmentIndex < 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	value, err := task.WaitSubtitleVTT(c.Request.Context(), trackIndex, segmentIndex, subtitleWaitTimeout)
	if err != nil {
		return
	}
	c.Data(http.StatusOK, "text/vtt; charset=utf-8", []byte(value))
}

func buildSubtitlePlaylist(task *Task, trackIndex int) string {
	videoPlaylist := buildTaskPlaylist(task, 0, task.Audio)
	lines := strings.Split(videoPlaylist, "\n")
	var result strings.Builder
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "#EXT-X-MAP:"):
			continue
		case strings.HasPrefix(line, "seg/"):
			index := strings.TrimSuffix(strings.TrimPrefix(line, "seg/"), ".m4s")
			result.WriteString(strconv.Itoa(trackIndex))
			result.WriteByte('/')
			result.WriteString(index)
			result.WriteString(".vtt\n")
		default:
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	return result.String()
}

func writeSegment(c *gin.Context, seg Segment) error {
	totalLength := int64(seg.Len())

	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader == "" {
		c.Header("Content-Length", strconv.FormatInt(totalLength, 10))
		_, err := seg.WriteTo(c.Writer)
		return err
	}

	start, end, ok := parseSingleRange(rangeHeader, totalLength)
	if !ok {
		c.Header("Content-Range", "bytes */"+strconv.FormatInt(totalLength, 10))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return nil
	}

	length := end - start + 1
	c.Header("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(totalLength, 10))
	c.Header("Content-Length", strconv.FormatInt(length, 10))
	c.Status(http.StatusPartialContent)

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
	if count > 0 && index > count {
		return count
	}
	return index
}

func abortWithSourceError(c *gin.Context, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		c.String(http.StatusGatewayTimeout, err.Error())
		return
	}
	c.String(http.StatusBadGateway, err.Error())
}

func abortWithRequestError(c *gin.Context, status int, err error) {
	if errors.Is(err, context.Canceled) || c.Request.Context().Err() != nil {
		return
	}
	if c.Writer.Written() {
		_ = c.Error(err)
		return
	}
	_ = c.AbortWithError(status, err)
}

func noCache(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}
