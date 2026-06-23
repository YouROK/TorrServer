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
	path, err := exec.LookPath("ffprobe")
	if err == nil {
		ffprobe.SetFFProbeBinPath(path)
		binFile = path
	} else {
		// working dir
		if _, err := os.Stat("ffprobe"); os.IsNotExist(err) {
			ffprobe.SetFFProbeBinPath(filepath.Dir(os.Args[0]) + "/ffprobe")
			binFile = filepath.Dir(os.Args[0]) + "/ffprobe"
		}
	}
}

func Exists() bool {
	_, err := os.Stat(binFile)
	return !os.IsNotExist(err)
}

func ProbeUrl(link string) (*ffprobe.ProbeData, error) {
	data, err := ProbeUrlWithTimeout(link, 5*time.Minute)
	return data, err
}

func ProbeUrlWithTimeout(link string, timeout time.Duration) (*ffprobe.ProbeData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ffprobe.ProbeURL(ctx, link)
}

func ProbeReader(reader io.Reader) (*ffprobe.ProbeData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	data, err := ffprobe.ProbeReader(ctx, reader)
	return data, err
}
