package torr

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkServeStreamRange(b *testing.B) {
	body := bytes.Repeat([]byte("x"), 1<<20)
	src := bytes.NewReader(body)
	req := httptest.NewRequest(http.MethodGet, "/stream?play=1", nil)
	req.Header.Set("Range", "bytes=4096-65535")
	w := &streamDiscard{}
	b.ReportAllocs()
	b.SetBytes(64 << 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := src.Seek(0, 0); err != nil {
			b.Fatal(err)
		}
		ServeStreamContent(w, req, "movie.mkv", time.Unix(0, 0), src)
	}
}

type streamDiscard struct{ h http.Header }

func (w *streamDiscard) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}
	return w.h
}
func (w *streamDiscard) Write(p []byte) (int, error) { return len(p), nil }
func (w *streamDiscard) WriteHeader(int)             {}
