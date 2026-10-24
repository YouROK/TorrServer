package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"server/torr"
	"server/torr/state"
)

func BenchmarkPlaylistGeneration(b *testing.B) {
	st := &state.TorrentStatus{Hash: "0123456789abcdef0123456789abcdef01234567", Title: "Show"}
	for i := 0; i < 40; i++ {
		name := fmt.Sprintf("Season/episode%02d", i)
		st.FileStats = append(st.FileStats, &state.TorrentFileStat{
			Id: i + 1, Path: name + ".mkv", Length: 1 << 30,
		})
		st.FileStats = append(st.FileStats, &state.TorrentFileStat{
			Id: 1000 + i, Path: name + ".en.srt", Length: 4096,
		})
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = getM3uList(st, "http://127.0.0.1:8090", false, "")
	}
	if sink == "" {
		b.Fatal("empty playlist")
	}
}

func BenchmarkStreamRangeHeader(b *testing.B) {
	body := bytes.Repeat([]byte("x"), 1<<20)
	src := bytes.NewReader(body)
	req := httptest.NewRequest(http.MethodGet, "/stream?play=1", nil)
	req.Header.Set("Range", "bytes=4096-8191")
	w := &discardRW{}
	b.ReportAllocs()
	b.SetBytes(4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			b.Fatal(err)
		}
		torr.ServeStreamContent(w, req, "movie.mkv", time.Unix(0, 0), src)
	}
}

type discardRW struct{ h http.Header }

func (w *discardRW) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}
	return w.h
}
func (w *discardRW) Write(p []byte) (int, error) { return len(p), nil }
func (w *discardRW) WriteHeader(int)             {}
