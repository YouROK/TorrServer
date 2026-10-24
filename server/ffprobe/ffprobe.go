package ffprobe

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"server/log"

	goffprobe "gopkg.in/vansante/go-ffprobe.v2"
)

const (
	// Four ffprobe processes at once. Each one is a process plus a /play reader.
	// The cap is concurrent processes, not probes per day.
	maxConcurrentProbes = 4

	// One socket read, in microseconds. Stock ffprobe waits forever on a silent socket.
	rwTimeoutArg = "30000000"
)

var (
	binFile      = "ffprobe"
	probeTimeout = time.Minute

	probeSlots       = make(chan struct{}, maxConcurrentProbes)
	probeInFlight    atomic.Int32
	probeInFlightMax atomic.Int32
)

func init() {
	path, err := exec.LookPath("ffprobe")
	if err == nil {
		goffprobe.SetFFProbeBinPath(path)
		binFile = path
		return
	}

	fallbackPath := filepath.Join(filepath.Dir(os.Args[0]), "ffprobe")
	if _, err := os.Stat(fallbackPath); err == nil {
		goffprobe.SetFFProbeBinPath(fallbackPath)
		binFile = fallbackPath
	}
}

func Exists() bool {
	_, err := os.Stat(binFile)
	return !os.IsNotExist(err)
}

func ProbeUrl(link string) (*goffprobe.ProbeData, error) {
	return ProbeURLContext(context.Background(), link)
}

// ProbeURLContext probes link. At most maxConcurrentProbes ffprobe processes run.
// ctx cancels the wait and the process. The probe itself gets probeTimeout after a slot is free.
func ProbeURLContext(ctx context.Context, link string) (*goffprobe.ProbeData, error) {
	return run(ctx, func(ctx context.Context) (*goffprobe.ProbeData, error) {
		return goffprobe.ProbeURL(ctx, link, "-rw_timeout", rwTimeoutArg)
	})
}

func ProbeReader(reader io.Reader) (*goffprobe.ProbeData, error) {
	return run(context.Background(), func(ctx context.Context) (*goffprobe.ProbeData, error) {
		return goffprobe.ProbeReader(ctx, reader, "-rw_timeout", rwTimeoutArg)
	})
}

func run(ctx context.Context, probe func(context.Context) (*goffprobe.ProbeData, error)) (*goffprobe.ProbeData, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := acquireProbe(ctx); err != nil {
		log.TLogln("ffprobe: canceled waiting for a probe slot:", err)
		return nil, err
	}
	defer func() { <-probeSlots }()

	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	noteInFlight(1)
	defer noteInFlight(-1)
	data, err := probe(ctx)
	if err != nil {
		log.TLogln("ffprobe:", err)
	}
	return data, err
}

func acquireProbe(ctx context.Context) error {
	select {
	case probeSlots <- struct{}{}:
		return nil
	default:
	}
	log.TLogln("ffprobe: concurrency limit reached, waiting")
	select {
	case probeSlots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func noteInFlight(delta int32) {
	n := probeInFlight.Add(delta)
	if delta < 0 {
		return
	}
	for {
		old := probeInFlightMax.Load()
		if n <= old || probeInFlightMax.CompareAndSwap(old, n) {
			return
		}
	}
}
