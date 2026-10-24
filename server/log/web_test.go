package log

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWebLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	old := webLog
	webLog = log.New(&buf, "", 0)
	t.Cleanup(func() { webLog = old })

	r := gin.New()
	r.Use(WebLogger())
	r.POST("/x", func(c *gin.Context) { c.String(http.StatusCreated, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/x?q=1", strings.NewReader("hi"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatal(w.Code)
	}
	if !strings.Contains(buf.String(), "hi") || !strings.Contains(buf.String(), "/x?q=1") || !strings.Contains(buf.String(), "201") {
		t.Fatalf("log %q", buf.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/x", strings.NewReader("secret"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=x")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if !strings.Contains(buf.String(), "body hidden") {
		t.Fatalf("multipart log %q", buf.String())
	}
}

func TestWebLoggerNilSkips(t *testing.T) {
	gin.SetMode(gin.TestMode)
	old := webLog
	webLog = nil
	t.Cleanup(func() { webLog = old })
	r := gin.New()
	r.Use(WebLogger())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w.Code != http.StatusNoContent {
		t.Fatal(w.Code)
	}
}

func TestWebLogConcurrent(t *testing.T) {
	var buf bytes.Buffer
	old := webLog
	webLog = log.New(&buf, "", 0)
	t.Cleanup(func() { webLog = old })
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WebLogln("line")
		}()
	}
	wg.Wait()
	if strings.Count(buf.String(), "line") != 40 {
		t.Fatalf("count %d", strings.Count(buf.String(), "line"))
	}
}
