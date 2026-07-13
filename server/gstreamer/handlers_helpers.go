package gstreamer

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

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

func noCache(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}
