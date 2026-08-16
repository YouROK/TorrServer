package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	logPath    = ""
	webLogPath = ""
)

var webLog *log.Logger

var (
	logFile    *os.File
	webLogFile *os.File
)

func Init(path, webpath string) {
	logPath = path
	webLogPath = webpath

	shared := path != "" && webpath != "" && filepath.Clean(path) == filepath.Clean(webpath)
	if path != "" {
		path = filepath.Clean(path)
	}
	if webpath != "" {
		webpath = filepath.Clean(webpath)
	}

	if shared {
		ff, err := openLogFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error create log file: %v\n", err)
			return
		}
		logFile = ff
		webLogFile = ff
		webLog = log.New(ff, " ", log.LstdFlags)
		applyServerLog(ff)
		return
	}

	if webpath != "" {
		ff, err := os.OpenFile(webpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error create web log file: %v\n", err)
		} else {
			webLogFile = ff
			webLog = log.New(ff, " ", log.LstdFlags)
		}
	}

	if path != "" {
		ff, err := openLogFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error create log file: %v\n", err)
			return
		}
		logFile = ff
		applyServerLog(ff)
	}
}

func openLogFile(path string) (*os.File, error) {
	if fi, err := os.Lstat(path); err == nil {
		if fi.Size() >= 100*1024*1024 { // 100MB
			os.Remove(path)
		}
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
}

func applyServerLog(ff *os.File) {
	if err := redirectStdFDs(ff); err != nil {
		fmt.Fprintf(os.Stderr, "Error redirect stdout/stderr: %v\n", err)
	}
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("UTC0 ")
	log.SetOutput(ff)
}

func Close() {
	if logFile != nil {
		logFile.Close()
		if webLogFile == logFile {
			webLogFile = nil
			webLog = nil
		}
		logFile = nil
	}
	if webLogFile != nil {
		webLogFile.Close()
		webLogFile = nil
		webLog = nil
	}
}

func TLogln(v ...interface{}) {
	log.Println(v...)
}

func WebLogln(v ...interface{}) {
	if webLog != nil {
		webLog.Println(v...)
	}
}

func WebLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if webLog == nil {
			c.Next()
			return
		}
		body := ""
		// save body if not form or file
		if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			body = string(bodyBytes)
		} else {
			body = "body hidden, too large"
		}
		c.Next()

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		logStr := fmt.Sprintf("%3d | %12s | %-7s %#v %v",
			statusCode,
			clientIP,
			method,
			path,
			body,
		)
		WebLogln(logStr)
	}
}

// TorrentLogHandler implements filtered slog.Handler in a minimal way
type TorrentLogHandler struct {
	level          slog.Level
	bannedPrefixes []string
}

func NewTorrentLogHandler(level slog.Level) *TorrentLogHandler {
	return &TorrentLogHandler{
		level: level,
		bannedPrefixes: []string{
			"reader",
			"readAt",
		},
	}
}

// slog levels:
// LevelDebug Level = -4
// LevelInfo  Level = 0
// LevelWarn  Level = 4
// LevelError Level = 8
func (h *TorrentLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *TorrentLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	msg := record.Message

	for _, prefix := range h.bannedPrefixes {
		if strings.Contains(msg, prefix) {
			return nil
		}
	}

	// Build the message
	var builder strings.Builder
	builder.WriteString("[")
	builder.WriteString(record.Level.String())
	builder.WriteString("] ")
	builder.WriteString(record.Message)

	// Add attributes
	record.Attrs(func(attr slog.Attr) bool {
		builder.WriteString(fmt.Sprintf(" %s=%v", attr.Key, attr.Value.Any()))
		return true
	})

	TLogln("TORRENT", builder.String())

	return nil
}

func (h *TorrentLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, return same handler. Attributes will be added at log time.
	return h
}

func (h *TorrentLogHandler) WithGroup(name string) slog.Handler {
	// For simplicity, ignore groups
	return h
}

// TorrentLogger creates a slog.Logger for torrent client
func DebugTorrentLogger() *slog.Logger {
	handler := NewTorrentLogHandler(slog.LevelDebug)
	return slog.New(handler)
}

func TorrentLogger() *slog.Logger {
	handler := NewTorrentLogHandler(slog.LevelInfo)
	return slog.New(handler)
}

func WarnTorrentLogger() *slog.Logger {
	handler := NewTorrentLogHandler(slog.LevelWarn)
	return slog.New(handler)
}

func ErrorTorrentLogger() *slog.Logger {
	handler := NewTorrentLogHandler(slog.LevelError)
	return slog.New(handler)
}

// FIXME: recursion in
// log/slog.AnyValue(...)
// log/slog.argsToAttr(...)
// log/slog.(*Record).Add(...)
// log/slog.(*Logger).log(...)
// github.com/anacrolix/torrent.(*reader).readAt(...)
func NullTorrentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError + 1000, // Log nothing
	}))
}
