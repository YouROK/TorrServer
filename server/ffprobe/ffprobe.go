package ffprobe

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/vansante/go-ffprobe.v2"
)

var binFile = "ffprobe"

func init() {
	// Try to find ffprobe in PATH first
	path, err := exec.LookPath("ffprobe")
	if err == nil {
		ffprobe.SetFFProbeBinPath(path)
		binFile = path
		return
	}

	// Fallback: check for the binary in the application's directory
	appDir := filepath.Dir(os.Args[0])
	fallbackPath := filepath.Join(appDir, "ffprobe")

	if _, err := os.Stat(fallbackPath); err == nil {
		ffprobe.SetFFProbeBinPath(fallbackPath)
		binFile = fallbackPath
	}
}

func Exists() bool {
	_, err := os.Stat(binFile)
	return !os.IsNotExist(err)
}

func ProbeUrl(link string) (*ffprobe.ProbeData, error) {
	data, err := ffprobe.ProbeURL(getCtx(), link)
	return data, err
}

func ProbeReader(reader io.Reader) (*ffprobe.ProbeData, error) {
	data, err := ffprobe.ProbeReader(getCtx(), reader)
	return data, err
}

func getCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1 * time.Minute)
		cancel()
	}()
	return ctx
}
