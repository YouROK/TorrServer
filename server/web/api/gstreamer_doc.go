package api

import "server/gstreamer"

type gstreamerSettingsDocResponse struct {
	BuiltIn  bool        `json:"built_in"`
	Config   interface{} `json:"config,omitempty"`
	Defaults interface{} `json:"defaults,omitempty"`
}

type gstreamerSettingsDocRequest struct {
	Action string            `json:"action,omitempty"`
	Config *gstreamer.Config `json:"config,omitempty"`
}

type gstreamerEchoDocResponse struct {
	GSTDiscoverer gstreamerComponentDocStatus `json:"gst_discoverer"`
	GStreamer     gstreamerComponentDocStatus `json:"gstreamer"`
}

type gstreamerComponentDocStatus struct {
	Found     bool   `json:"found"`
	Available bool   `json:"available"`
	Works     bool   `json:"works"`
	Error     string `json:"error,omitempty"`
}

// GetGStreamerSettingsDoc godoc
// @Summary Get GStreamer configuration
// @Description On `-gst` builds returns built_in, config and defaults. On standard builds returns built_in: false only.
// @Tags GStreamer
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gstreamerSettingsDocResponse "GStreamer settings"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /gst/settings [get]
func GetGStreamerSettingsDoc() {}

// UpdateGStreamerSettingsDoc godoc
// @Summary Update GStreamer configuration
// @Description Available on `-gst` builds only; standard builds return 404.
// @Tags GStreamer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body gstreamerSettingsDocRequest true "GStreamer settings request"
// @Success 200 {object} map[string]string "Update successful"
// @Failure 400 {object} map[string]string "Invalid input data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Read-only mode"
// @Failure 404 {object} map[string]string "GStreamer is not built in"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /gst/settings [post]
func UpdateGStreamerSettingsDoc() {}

// GetGStreamerEchoDoc godoc
// @Summary GStreamer health check
// @Description Requires a `-gst` build. Not available in standard binaries.
// @Tags GStreamer
// @Produce json
// @Success 200 {object} gstreamerEchoDocResponse "GStreamer runtime status"
// @Router /gst/echo [get]
func GetGStreamerEchoDoc() {}

// RemoveGStreamerTaskDoc godoc
// @Summary Stop GStreamer transcode task
// @Description Requires a `-gst` build. Not available in standard binaries.
// @Tags GStreamer
// @Produce json
// @Param hash query string false "Torrent infohash"
// @Param id query string false "Torrent infohash (alias)"
// @Success 200 {object} map[string]bool "Task removed"
// @Failure 400 {object} map[string]string "Invalid identifier"
// @Failure 404 "Task not found"
// @Router /gst/remove [get]
func RemoveGStreamerTaskDoc() {}

// GetGStreamerHeartbeatDoc godoc
// @Summary GStreamer task heartbeat
// @Description Requires a `-gst` build. Not available in standard binaries.
// @Tags GStreamer
// @Produce json
// @Param hash path string true "Torrent infohash"
// @Success 200 {object} map[string]interface{} "Torrent heartbeat state"
// @Failure 404 "Task not found"
// @Router /gst/{hash}/heartbeat [get]
func GetGStreamerHeartbeatDoc() {}

// ProbeGStreamerSourceDoc godoc
// @Summary Probe torrent media tracks
// @Description Requires a `-gst` build. Uses gst-discoverer to inspect container and codec information.
// @Tags GStreamer
// @Produce json
// @Param hash path string true "Torrent infohash"
// @Param index query string false "File index in torrent"
// @Param id query string false "File index (alias)"
// @Param fileID query string false "File index (alias)"
// @Success 200 {object} gstreamer.ProbeInfo "Media probe result"
// @Failure 400 {object} map[string]string "Invalid source"
// @Failure 502 {string} string "Probe failed"
// @Failure 504 {string} string "Probe timed out"
// @Router /gst/{hash}/probe [get]
func ProbeGStreamerSourceDoc() {}

// GetGStreamerMasterPlaylistDoc godoc
// @Summary HLS master playlist
// @Description Requires a `-gst` build. Returns an HLS VOD master playlist for transcoded playback.
// @Tags GStreamer
// @Produce application/vnd.apple.mpegurl
// @Param hash path string true "Torrent infohash"
// @Param index query string false "File index in torrent"
// @Param id query string false "File index (alias)"
// @Param fileID query string false "File index (alias)"
// @Param audio query int false "Audio track index" default(0)
// @Param seconds query int false "Start offset in seconds" default(0)
// @Success 200 {string} string "application/vnd.apple.mpegurl playlist"
// @Failure 502 {string} string "Pipeline error"
// @Router /gst/{hash}/master.m3u8 [get]
func GetGStreamerMasterPlaylistDoc() {}

// GetGStreamerInitSegmentDoc godoc
// @Summary HLS initialization segment
// @Description Requires a `-gst` build. Returns the fMP4 init segment for the transcode session.
// @Tags GStreamer
// @Produce video/mp4
// @Param hash path string true "Torrent infohash"
// @Param audio query int false "Audio track index"
// @Param seconds query int false "Start offset in seconds" default(0)
// @Success 200 {file} file "video/mp4 init segment"
// @Failure 404 "Task not found"
// @Failure 502 {string} string "Pipeline error"
// @Router /gst/{hash}/init.mp4 [get]
func GetGStreamerInitSegmentDoc() {}

// GetGStreamerMediaSegmentDoc godoc
// @Summary HLS media segment
// @Description Requires a `-gst` build. Returns an fMP4 media segment for HLS playback.
// @Tags GStreamer
// @Produce video/mp4
// @Param hash path string true "Torrent infohash"
// @Param segment path string true "Segment index (e.g. 0.m4s)"
// @Param audio query int false "Audio track index"
// @Success 200 {file} file "video/mp4 media segment"
// @Success 206 {file} file "Partial content"
// @Failure 400 {object} map[string]string "Invalid segment"
// @Failure 404 "Task not found"
// @Failure 502 {string} string "Pipeline error"
// @Router /gst/{hash}/seg/{segment} [get]
func GetGStreamerMediaSegmentDoc() {}
