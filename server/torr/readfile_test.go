package torr

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"server/leakcheck"
	"server/torr/offlinetest"
)

func fileIDByPath(t *testing.T, tor *Torrent, suffix string) int {
	t.Helper()
	for _, f := range tor.Status().FileStats {
		if strings.HasSuffix(f.Path, suffix) {
			return f.Id
		}
	}
	t.Fatalf("no file %q", suffix)
	return 0
}

func TestReadFileReturnsExactlyTheFile(t *testing.T) {
	tor, _ := connectShow(t)

	// The srt sits between two larger files in piece order; ReadFull on the
	// reader would run into the neighbour if ReadFile did not cap at its size.
	id := fileIDByPath(t, tor, "episode1.srt")
	got, err := tor.ReadFile(t.Context(), id, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if want := "1\n00:00:00,000 --> 00:00:01,000\nHi\n"; string(got) != want {
		t.Fatalf("ReadFile = %q, want %q", got, want)
	}

	// The poster is last in the torrent: ReadFile must stop at its last byte.
	poster, err := tor.ReadFile(t.Context(), fileIDByPath(t, tor, "poster.jpg"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{0xff, 0xd8, 0xff, 0xd9, 0x00, 0x01}; string(poster) != string(want) {
		t.Fatalf("poster = %v, want %v", poster, want)
	}
}

func TestReadFileRejects(t *testing.T) {
	tor, _ := connectShow(t)
	id := fileIDByPath(t, tor, "episode1.srt")

	if _, err := tor.ReadFile(t.Context(), id, 8); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Errorf("size cap: %v", err)
	}
	// The cap is inclusive: a file of exactly maxSize is allowed.
	size := int64(len("1\n00:00:00,000 --> 00:00:01,000\nHi\n"))
	if _, err := tor.ReadFile(t.Context(), id, size); err != nil {
		t.Errorf("file of exactly maxSize rejected: %v", err)
	}
	if _, err := tor.ReadFile(t.Context(), id, size-1); err == nil {
		t.Error("file of maxSize+1 accepted")
	}
	if _, err := tor.ReadFile(t.Context(), 9999, 1<<20); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("unknown id: %v", err)
	}
}

func TestReadFileNoInfo(t *testing.T) {
	tor := &Torrent{}
	if _, err := tor.ReadFile(t.Context(), 1, 1); err == nil {
		t.Fatal("torrent without info must fail, not panic")
	}
}

// With no peer the read blocks forever. ReadFile must return when the context
// ends and must not leave the reader goroutine behind (the read runs on a context
// that ReadFile cancels on return).
func TestReadFileHonoursContextAndDoesNotLeak(t *testing.T) {
	offlinetest.Prepare(t)
	show := offlinetest.WriteShow(t)
	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)
	tor, err := AddTorrent(show.Spec, "Show", "", "", "tv")
	if err != nil {
		t.Fatal(err)
	}
	id := fileIDByPath(t, tor, "episode1.srt")

	base := runtime.NumGoroutine()
	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = tor.ReadFile(ctx, id, 1<<20)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Fatalf("ReadFile ignored the context for %v", d)
	}
	leakcheck.AssertNoGoroutineLeak(t, base)
	if n := tor.cache.GetUseReaders(); n != 0 {
		t.Errorf("reader left registered after ctx expiry: %d", n)
	}
}

// Every expired ReadFile must release its reader goroutine, not just the first:
// with no peer the read can only end through the context.
func TestReadFileCancelReleasesEveryGoroutine(t *testing.T) {
	offlinetest.Prepare(t)
	show := offlinetest.WriteShow(t)
	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)
	tor, err := AddTorrent(show.Spec, "Show", "", "", "tv")
	if err != nil {
		t.Fatal(err)
	}
	id := fileIDByPath(t, tor, "episode1.srt")

	base := runtime.NumGoroutine()
	for i := 0; i < 12; i++ {
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
		_, err := tor.ReadFile(ctx, id, 1<<20)
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("round %d: err = %v, want DeadlineExceeded", i, err)
		}
	}
	leakcheck.AssertNoGoroutineLeak(t, base)
	if n := tor.cache.GetUseReaders(); n != 0 {
		t.Errorf("%d readers left registered", n)
	}
}
