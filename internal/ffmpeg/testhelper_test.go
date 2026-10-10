package ffmpeg

import (
	"bytes"
	"os/exec"
	"time"
)

// newTestCommand создаёт команду для вспомогательных вызовов в тестах.
func newTestCommand(bin string, args ...string) *exec.Cmd {
	return exec.Command(bin, args...)
}

// newReaderAt отдаёт данные из памяти как io.ReadSeeker для http.ServeContent.
func newReaderAt(data []byte) *bytes.Reader {
	return bytes.NewReader(data)
}

// secondsToDuration переводит секунды в длительность.
func secondsToDuration(seconds float64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}
