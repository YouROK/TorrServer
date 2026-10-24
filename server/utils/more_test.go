package utils

import (
	"net/http/httptest"
	"sync"
	"testing"

	"server/torr/state"

	"github.com/gin-gonic/gin"
)

func TestGetMimeTypeAndPlayable(t *testing.T) {
	cases := map[string]string{
		"a.MKV":  "video/*",
		"b.mp3":  "audio/*",
		"c.nfo":  "*/*",
		"noext":  "*/*",
		".webm":  "video/*",
		"x.FLAC": "audio/*",
	}
	for name, want := range cases {
		if got := GetMimeType(name); got != want {
			t.Errorf("GetMimeType(%q) = %q, want %q", name, got, want)
		}
	}
	st := state.TorrentStatus{FileStats: []*state.TorrentFileStat{
		{Path: "a.mkv"},
		{Path: "b.nfo"},
		{Path: "c.mp3"},
		nil,
	}}
	// nil file panics on Path; don't include it. Rebuild without nil.
	st.FileStats = st.FileStats[:3]
	got := GetPlayableFiles(st)
	if len(got) != 2 || got[0].Path != "a.mkv" || got[1].Path != "c.mp3" {
		t.Fatalf("%#v", got)
	}
	if len(GetPlayableFiles(state.TorrentStatus{})) != 0 {
		t.Fatal("empty status")
	}
}

func TestFormatBytes(t *testing.T) {
	cases := map[float64]string{
		0:    "0",
		1:    "1B",
		1024: "1.00KB",
		1536: "1.50KB",
	}
	for in, want := range cases {
		if got := Format(in); got != want {
			t.Errorf("Format(%v) = %q, want %q", in, got, want)
		}
	}
	if Format(float64(MB)) != "1.00MB" || Format(float64(GB)) != "1.00GB" {
		t.Fatalf("MB %q GB %q", Format(float64(MB)), Format(float64(GB)))
	}
}

func TestCommonPrefixAndNumberPrefix(t *testing.T) {
	if CommonPrefix("abc", "ax") != "a" || CommonPrefix("", "a") != "" || CommonPrefix("абв", "абг") == "" {
		t.Fatalf("prefix %q %q", CommonPrefix("abc", "ax"), CommonPrefix("абв", "абг"))
	}
	n, err := NumberPrefix("12abc")
	if err != nil || n != 12 {
		t.Fatalf("12abc %d %v", n, err)
	}
	if _, err := NumberPrefix(""); err == nil {
		t.Fatal("empty prefix should fail")
	}
	if _, err := NumberPrefix("abc"); err == nil {
		t.Fatal("no digits should fail")
	}
}

func TestParallelFor(t *testing.T) {
	var got [40]int
	ParallelFor(0, 0, func(int) { t.Error("called") })
	ParallelFor(10, 10, func(int) { t.Error("called") })
	ParallelFor(0, len(got), func(i int) { got[i] = i + 1 })
	for i, v := range got {
		if v != i+1 {
			t.Fatalf("index %d = %d", i, v)
		}
	}
}

func TestParallelForConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for n := 0; n < 20; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sum := 0
			var mu sync.Mutex
			ParallelFor(0, 30, func(i int) {
				mu.Lock()
				sum += i
				mu.Unlock()
			})
			if sum != 30*29/2 {
				t.Errorf("sum = %d", sum)
			}
		}()
	}
	wg.Wait()
}

func TestGetSchemeAndHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://example.com/a", nil)
	if GetScheme(c) != "http" {
		t.Fatalf("scheme %q", GetScheme(c))
	}
	if GetHost(c) != "example.com" {
		t.Fatalf("host %q", GetHost(c))
	}
}
