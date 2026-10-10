package job

import (
	"strings"
	"sync"
)

// ringBuffer хранит последние строки вывода процесса.
type ringBuffer struct {
	mu    sync.Mutex
	lines []string
	size  int
}

const defaultStderrLines = 50

func newRingBuffer(size int) *ringBuffer {
	if size <= 0 {
		size = defaultStderrLines
	}
	return &ringBuffer{
		lines: make([]string, 0, size),
		size:  size,
	}
}

// Write добавляет порцию текста, разбивая её на строки.
func (r *ringBuffer) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, line := range strings.Split(string(p), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if len(r.lines) == r.size {
			copy(r.lines, r.lines[1:])
			r.lines = r.lines[:r.size-1]
		}
		r.lines = append(r.lines, line)
	}
	return len(p), nil
}

// Lines возвращает копию сохранённых строк.
func (r *ringBuffer) Lines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]string, len(r.lines))
	copy(out, r.lines)
	return out
}

// joinCommand собирает читаемое представление команды.
func joinCommand(args []string) string {
	return strings.Join(args, " ")
}
