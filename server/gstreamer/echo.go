package gstreamer

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type echoResponse struct {
	FFProbe   componentStatus `json:"ffprobe"`
	GStreamer componentStatus `json:"gstreamer"`
}

type componentStatus struct {
	Found     bool `json:"found"`
	Available bool `json:"available"`
	Works     bool `json:"works"`
}

func (s *Service) echo(c *gin.Context) {
	c.JSON(http.StatusOK, echoResponse{
		FFProbe:   checkFFProbe(),
		GStreamer: checkGStreamer(s.conf),
	})
}

func checkFFProbe() componentStatus {
	var status componentStatus

	path, ok := findFFProbeBinary()
	if !ok {
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

	if err := exec.CommandContext(ctx, path, "-version").Run(); err == nil {
		status.Works = true
	}
	return status
}

func findFFProbeBinary() (string, bool) {
	if path, err := exec.LookPath("ffprobe"); err == nil {
		return path, true
	}

	dirs := []string{"."}
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}

	seen := make(map[string]struct{}, len(dirs))
	for _, dir := range dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if _, ok := seen[absDir]; ok {
			continue
		}
		seen[absDir] = struct{}{}

		for _, name := range []string{"ffprobe", "ffprobe.exe"} {
			path := filepath.Join(absDir, name)
			if _, err := os.Stat(path); err == nil {
				return path, true
			}
		}
	}

	return "", false
}
