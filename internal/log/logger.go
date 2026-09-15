package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LevelNone отключает абсолютно все логи
const LevelNone = slog.Level(1000)

var (
	defaultLogger *slog.Logger
	levelVar      = new(slog.LevelVar)
	logFile       *os.File
	ringBuf       *ringBuffer
	mu            sync.RWMutex
)

func init() {
	ringBuf = newRingBuffer(500)
	levelVar.Set(slog.LevelInfo)
	handler := newCustomHandler(io.MultiWriter(os.Stdout, ringBuf), levelVar)
	defaultLogger = slog.New(handler)
}

func Init(levelStr, filePath string) error {
	mu.Lock()
	defer mu.Unlock()

	// 1. Устанавливаем уровень (поддерживаем отключение через none / off / disabled)
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "none", "off", "disabled", "mute":
		levelVar.Set(LevelNone)
	case "debug":
		levelVar.Set(slog.LevelDebug)
	case "warn", "warning":
		levelVar.Set(slog.LevelWarn)
	case "error":
		levelVar.Set(slog.LevelError)
	default:
		levelVar.Set(slog.LevelInfo)
	}

	// Если логи полностью выключены, даже файл открывать не нужно
	if levelVar.Level() == LevelNone {
		if logFile != nil {
			_ = logFile.Close()
			logFile = nil
		}
		defaultLogger = slog.New(newCustomHandler(io.Discard, levelVar))
		return nil
	}

	// 2. Настраиваем вывод
	writers := []io.Writer{os.Stdout, ringBuf}

	if filePath != "" {
		dir := filepath.Dir(filePath)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("не удалось создать папку для логов: %w", err)
			}
		}

		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("ошибка открытия файла логов (%s): %w", filePath, err)
		}

		if logFile != nil {
			_ = logFile.Close()
		}
		logFile = f
		writers = append(writers, f)
	}

	mw := io.MultiWriter(writers...)
	defaultLogger = slog.New(newCustomHandler(mw, levelVar))
	return nil
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		_ = logFile.Close()
	}
}

func GetRecentLogs() []string {
	return ringBuf.GetAll()
}

// ============================================================================
// Обертки логирования
// ============================================================================

func logMsg(level slog.Level, msg string, args ...any) {
	// Если уровень выключен (None) — выходим мгновенно с нулевым оверхедом
	if !defaultLogger.Enabled(context.Background(), level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])

	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = defaultLogger.Handler().Handle(context.Background(), r)
}

func Debug(msg string, args ...any) { logMsg(slog.LevelDebug, msg, args...) }
func Info(msg string, args ...any)  { logMsg(slog.LevelInfo, msg, args...) }
func Warn(msg string, args ...any)  { logMsg(slog.LevelWarn, msg, args...) }
func Error(msg string, args ...any) { logMsg(slog.LevelError, msg, args...) }

func Debugf(format string, args ...any) { logMsg(slog.LevelDebug, fmt.Sprintf(format, args...)) }
func Infof(format string, args ...any)  { logMsg(slog.LevelInfo, fmt.Sprintf(format, args...)) }
func Warnf(format string, args ...any)  { logMsg(slog.LevelWarn, fmt.Sprintf(format, args...)) }
func Errorf(format string, args ...any) { logMsg(slog.LevelError, fmt.Sprintf(format, args...)) }

// ============================================================================
// Форматтер
// ============================================================================

type customHandler struct {
	w  io.Writer
	lv *slog.LevelVar
	mu sync.Mutex
}

func newCustomHandler(w io.Writer, lv *slog.LevelVar) *customHandler {
	return &customHandler{w: w, lv: lv}
}

func (h *customHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.lv.Level()
}

func (h *customHandler) Handle(_ context.Context, r slog.Record) error {
	buf := bytes.NewBuffer(make([]byte, 0, 128))

	// Время [2026-09-04 15:04:05.000]
	buf.WriteString(r.Time.Format("[2006-01-02 15:04:05.000] "))

	// Уровень без пробела: [INFO] / [DEBUG] / [ERROR]
	levelStr := fmt.Sprintf("[%s] ", r.Level.String())
	buf.WriteString(levelStr)

	// Для DEBUG и ERROR добавляем стек
	if r.PC != 0 && (r.Level == slog.LevelDebug || r.Level >= slog.LevelError) {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()

		fnName := f.Function
		if lastSlash := strings.LastIndex(fnName, "/"); lastSlash >= 0 {
			fnName = fnName[lastSlash+1:]
		}

		callerStr := fmt.Sprintf("[%s:%d %s] ", filepath.Base(f.File), f.Line, fnName)
		buf.WriteString(callerStr)
	}

	// Сообщение
	buf.WriteString(r.Message)

	// Дополнительные параметры
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value))
		return true
	})

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	return err
}

func (h *customHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *customHandler) WithGroup(_ string) slog.Handler      { return h }

// ============================================================================
// Кольцевой буфер строк
// ============================================================================

type ringBuffer struct {
	mu    sync.RWMutex
	lines []string
	size  int
	pos   int
	full  bool
}

func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		lines: make([]string, size),
		size:  size,
	}
}

func (r *ringBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	line := strings.TrimRight(string(p), "\r\n")
	r.lines[r.pos] = line
	r.pos = (r.pos + 1) % r.size
	if r.pos == 0 {
		r.full = true
	}
	return len(p), nil
}

func (r *ringBuffer) GetAll() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.full {
		res := make([]string, r.pos)
		copy(res, r.lines[:r.pos])
		return res
	}

	res := make([]string, r.size)
	copy(res, r.lines[r.pos:])
	copy(res[r.size-r.pos:], r.lines[:r.pos])
	return res
}
