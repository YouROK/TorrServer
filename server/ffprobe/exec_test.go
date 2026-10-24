package ffprobe

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	goffprobe "gopkg.in/vansante/go-ffprobe.v2"
)

func writeMock(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("ffprobe mock is a shell script")
	}
	path := filepath.Join(t.TempDir(), "ffprobe")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func useMock(t *testing.T, path string) {
	t.Helper()
	oldBin, oldTimeout := binFile, probeTimeout
	goffprobe.SetFFProbeBinPath(path)
	binFile = path
	t.Cleanup(func() {
		binFile = oldBin
		probeTimeout = oldTimeout
		goffprobe.SetFFProbeBinPath(oldBin)
	})
}

const okScript = "#!/bin/sh\ncat >/dev/null\nprintf '%s' '{\"format\":{\"duration\":\"1.500000\",\"bit_rate\":\"1000\",\"size\":\"10\"}}'\n"

func TestProbeReaderMock(t *testing.T) {
	useMock(t, writeMock(t, okScript))
	data, err := ProbeReader(strings.NewReader("media-bytes"))
	if err != nil {
		t.Fatal(err)
	}
	if data.Format == nil || data.Format.DurationSeconds != 1.5 {
		t.Fatalf("format = %#v", data.Format)
	}
	if !Exists() {
		t.Fatal("Exists() false for mock binary")
	}
}

func TestProbeURLMock(t *testing.T) {
	useMock(t, writeMock(t, okScript))
	data, err := ProbeUrl(filepath.Join(t.TempDir(), "no-such-file"))
	if err != nil {
		t.Fatal(err)
	}
	if data.Format.BitRate != "1000" || data.Format.Size != "10" {
		t.Fatalf("format = %#v", data.Format)
	}
}

func TestProbeURLPassesSocketTimeout(t *testing.T) {
	out := filepath.Join(t.TempDir(), "args")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" > %q\nprintf '%%s' '{\"format\":{\"duration\":\"1.500000\",\"bit_rate\":\"1000\",\"size\":\"10\"}}'\n", out)
	useMock(t, writeMock(t, script))
	if _, err := ProbeUrl("http://127.0.0.1/media"); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "-rw_timeout\n"+rwTimeoutArg) {
		t.Fatalf("args missing socket timeout:\n%s", body)
	}
}

func TestProbeReaderBadJSON(t *testing.T) {
	useMock(t, writeMock(t, "#!/bin/sh\ncat >/dev/null\nprintf 'not-json'\n"))
	if _, err := ProbeReader(bytes.NewReader(nil)); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestProbeReaderNonZeroExit(t *testing.T) {
	useMock(t, writeMock(t, "#!/bin/sh\ncat >/dev/null\nexit 2\n"))
	if _, err := ProbeReader(strings.NewReader("x")); err == nil {
		t.Fatal("expected non-zero exit")
	}
}

func TestProbeReaderTimeout(t *testing.T) {
	useMock(t, writeMock(t, "#!/bin/sh\nexec sleep 30\n"))
	probeTimeout = 200 * time.Millisecond
	start := time.Now()
	_, err := ProbeReader(strings.NewReader("x"))
	if err == nil {
		t.Fatal("expected timeout")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatalf("probe took %s, context cancel did not stop ffprobe", time.Since(start))
	}
}

func TestExistsMissing(t *testing.T) {
	useMock(t, filepath.Join(t.TempDir(), "missing-ffprobe"))
	if Exists() {
		t.Fatal("Exists() true for missing binary")
	}
}

func TestProbeReaderConcurrent(t *testing.T) {
	useMock(t, writeMock(t, "#!/bin/sh\ncat >/dev/null\nsleep 0.3\nprintf '%s' '{\"format\":{\"duration\":\"1.500000\",\"bit_rate\":\"1000\",\"size\":\"10\"}}'\n"))
	probeInFlightMax.Store(0)
	const n = 30
	var wg sync.WaitGroup
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := ProbeReader(strings.NewReader("x")); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
	if got := probeInFlightMax.Load(); got != int32(maxConcurrentProbes) {
		t.Fatalf("peak ffprobe processes = %d, want %d", got, maxConcurrentProbes)
	}
	if got := probeInFlight.Load(); got != 0 {
		t.Fatalf("ffprobe still in flight: %d", got)
	}
}
