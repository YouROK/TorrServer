package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
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
	webLogPath = webpath
	logPath = path

	if webpath != "" {
		ff, err := os.OpenFile(webLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			TLogln("Error create web log file:", err)
		} else {
			webLogFile = ff
			webLog = log.New(ff, " ", log.LstdFlags)
		}
	}

	if path != "" {
		if fi, err := os.Lstat(path); err == nil {
			if fi.Size() >= 100*1024*1024 { // 100MB
				os.Remove(path)
			}
		}
		ff, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			TLogln("Error create log file:", err)
			return
		}
		logFile = ff
		os.Stdout = ff
		os.Stderr = ff
		// var timeFmt string
		// var ok bool
		// timeFmt, ok = os.LookupEnv("GO_LOG_TIME_FMT")
		// if !ok {
		// 	timeFmt = "2006-01-02T15:04:05-0700"
		// }
		// log.SetFlags(log.Lmsgprefix)
		// log.SetPrefix(time.Now().Format(timeFmt) + " TSM ")
		log.SetFlags(log.LstdFlags | log.LUTC | log.Lmsgprefix)
		log.SetPrefix("UTC0 ")
		log.SetOutput(ff)
	}
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
	if webLogFile != nil {
		webLogFile.Close()
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
			body, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
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
			string(body),
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
			// "github.com/anacrolix/torrent",
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

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "TORRENT [%s] %s", record.Level.String(), msg)

	record.Attrs(func(attr slog.Attr) bool {
		fmt.Fprintf(&buf, " %s=%v", attr.Key, attr.Value.Any())
		return true
	})

	buf.WriteByte('\n')

	if logFile != nil {
		logFile.Write(buf.Bytes())
	}

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
