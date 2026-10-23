package log

import (
	"bytes"
	"io"
	"log/slog"
	"sync"
)

// syncBuffer - потокобезопасный буфер для перехвата вывода логгера в тестах.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// newLoggerForTest собирает логгер с тем же форматтером, что и рабочий.
func newLoggerForTest(w io.Writer) *slog.Logger {
	return slog.New(newCustomHandler(w, levelVar))
}
