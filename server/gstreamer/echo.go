package gstreamer

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

type echoResponse struct {
	GSTDiscoverer componentStatus `json:"gst_discoverer"`
	GStreamer     componentStatus `json:"gstreamer"`
}

type componentStatus struct {
	Found     bool `json:"found"`
	Available bool `json:"available"`
	Works     bool `json:"works"`
}

func (s *Service) echo(c *gin.Context) {
	c.JSON(http.StatusOK, echoResponse{
		GSTDiscoverer: checkGSTDiscoverer(s.conf),
		GStreamer:     checkGStreamer(s.conf),
	})
}

func checkGSTDiscoverer(conf Config) componentStatus {
	var status componentStatus

	path, err := gstDiscovererPath(conf)
	if err != nil {
		return status
	}
	status.Found = true

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return status
	}
	status.Available = true

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "-h")
	cmd.Env = gstDiscovererEnv(conf)
	if err := cmd.Run(); err == nil {
		status.Works = true
	}
	return status
}
