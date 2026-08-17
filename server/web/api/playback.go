package api

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	pathpkg "path"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/playback"
)

var playbackHashPattern = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

type playbackRequest struct {
	DeviceID string `json:"device_id,omitempty"`
	Path     string `json:"path" binding:"required"`
	Hash     string `json:"hash" binding:"required"`
	Index    int    `json:"index"`
}

func listPlaybackDevices(c *gin.Context) {
	c.JSON(http.StatusOK, playback.GetPublicConfig())
}

func listPlaybackDeviceConfigs(c *gin.Context) {
	c.JSON(http.StatusOK, playback.GetManagedConfig())
}

func updatePlaybackSettings(c *gin.Context) {
	var input playback.Settings
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	settings, err := playback.UpdateSettings(input)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func upsertPlaybackDevice(c *gin.Context) {
	var input playback.DeviceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	device, err := playback.Upsert(input)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, device)
}

func deletePlaybackDevice(c *gin.Context) {
	if err := playback.Delete(c.Param("id")); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func testPlaybackDevice(c *gin.Context) {
	if err := playback.Test(c.Param("id")); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func playOnDevice(c *gin.Context) {
	var req playbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fileName, err := validatePlaybackRequest(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	deviceID, err := playback.ResolveTarget(req.DeviceID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	streamURL, err := playbackStreamURL(c, deviceID, fileName, req.Hash, req.Index)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := playback.Play(deviceID, playback.AgentPlayRequest{
		Path:      fileName,
		Hash:      strings.ToLower(req.Hash),
		Index:     req.Index,
		StreamURL: streamURL,
	}); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"ok": true, "device_id": deviceID})
}

func validatePlaybackRequest(req playbackRequest) (string, error) {
	if !playbackHashPattern.MatchString(req.Hash) {
		return "", errors.New("invalid torrent hash")
	}
	if req.Index < 0 || req.Index > 100000 {
		return "", errors.New("invalid file index")
	}
	fileName := pathpkg.Base(strings.ReplaceAll(strings.TrimSpace(req.Path), "\\", "/"))
	if fileName == "" || fileName == "." || fileName == ".." || len(fileName) > 1024 {
		return "", errors.New("invalid file path")
	}
	return fileName, nil
}

func playbackStreamURL(c *gin.Context, deviceID, fileName, hash string, index int) (string, error) {
	device, ok := playback.GetPrivate(deviceID)
	if !ok {
		return "", os.ErrNotExist
	}

	baseURL := device.StreamBaseURL
	if baseURL == "" {
		scheme := forwardedScheme(c)
		baseURL = scheme + "://" + c.Request.Host
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/stream/" + fileName
	query := parsed.Query()
	query.Set("link", strings.ToLower(hash))
	query.Set("index", strconv.Itoa(index))
	query.Set("play", "")
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func forwardedScheme(c *gin.Context) string {
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		if scheme := strings.TrimSpace(strings.Split(forwarded, ",")[0]); scheme == "http" || scheme == "https" {
			return scheme
		}
	}
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}
