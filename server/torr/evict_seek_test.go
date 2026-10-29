package torr

import (
	"bytes"
	"io"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/anacrolix/torrent"

	"server/leakcheck"
	"server/settings"
	"server/torr/offlinetest"
)

// Random seeks through a cache much smaller than the torrent evict partial
// pieces. Without Piece.Evict the client finished them over zeroed buffers,
// failed the hash and dropped the only seeder. The race needs a loaded CPU:
// run several copies at once with GOMAXPROCS=2 to reproduce it locally.
func TestEvictedPartialPiecesDoNotFailHash(t *testing.T) {
	baseG := runtime.NumGoroutine()
	offlinetest.Prepare(t)
	// Stray cleanPieces and AdjustRA goroutines read settings.BTsets; let them
	// exit before Prepare's cleanup restores it.
	t.Cleanup(func() { leakcheck.AssertNoGoroutineLeak(t, baseG) })
	settings.BTsets.CacheSize = 8 << 20
	show := offlinetest.WriteSized(t, "seek.mkv", 64<<20, 256<<10)
	seed := offlinetest.Seed(t, show)
	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)
	tor, err := AddTorrent(show.Spec, "seek", "", "", "")
	if err != nil || !tor.GotInfo() {
		t.Fatal("no info", err)
	}
	peer := seed.Peer()
	tor.AddPeers([]torrent.Peer{peer})
	file := tor.Files()[0]
	r := tor.NewReader(file)
	defer tor.CloseReader(r)
	r.SetResponsive()

	badHashes := func() int64 {
		st := tor.Stats()
		return st.PiecesDirtiedBad.Int64()
	}
	rnd := rand.New(rand.NewSource(1))
	buf := make([]byte, 256<<10)
	deadline := time.Now().Add(3 * time.Second)
	seeks := 0
	for time.Now().Before(deadline) {
		off := rnd.Int63n(file.Length() - int64(len(buf)))
		if _, err := r.Seek(off, io.SeekStart); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { _, err := io.ReadFull(r, buf); done <- err }()
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("read stalled after %d seeks, hash failures %d", seeks, badHashes())
		}
		if !bytes.Equal(buf, show.Data[off:off+int64(len(buf))]) {
			t.Fatalf("seek %d at %d read bytes the seeder never sent", seeks, off)
		}
		seeks++
	}
	if bad := badHashes(); bad != 0 {
		t.Fatalf("%d hash failures in %d seeks", bad, seeks)
	}
	if seeks < 20 {
		t.Fatalf("only %d seeks in 3s", seeks)
	}
}
