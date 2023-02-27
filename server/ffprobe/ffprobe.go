package ffprobe

import (
	"context"
	"gopkg.in/vansante/go-ffprobe.v2"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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
		time.Sleep(5 * time.Minute)
		cancel()
	}()
	return ctx
}
