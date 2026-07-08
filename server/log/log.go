package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
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
