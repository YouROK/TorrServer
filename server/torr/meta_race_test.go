package torr

import (
	"strconv"
	"sync"
	"testing"

	"github.com/anacrolix/torrent"

	"server/torr/offlinetest"
)

// Stream, play and subs read a torrent's metadata through Meta while the add
// handler's goroutine, SetTorrent, AddTorrent and AddTorrentDB write it. Run under
// -race: any unsynchronised access to Title, Poster, Category or Data is reported.
func TestMetaIsSafeAgainstConcurrentWriters(t *testing.T) {
	offlinetest.Prepare(t)
	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)

	show := offlinetest.WriteShow(t)
	seed := offlinetest.Seed(t, show)
	tor, err := AddTorrent(show.Spec, "Show", "", "", "tv")
	if err != nil || !tor.GotInfo() {
		t.Fatal(err)
	}
	tor.AddPeers([]torrent.Peer{seed.Peer()})
	hash := tor.Hash().HexString()

	const rounds = 40
	var wg sync.WaitGroup
	run := func(f func(i int)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < rounds; i++ {
				f(i)
			}
		}()
	}
	run(func(i int) { SetTorrent(hash, "title "+strconv.Itoa(i), "", "tv", "") })
	run(func(i int) { SaveTorrentToDB(tor) })
	run(func(i int) { _, _ = AddTorrent(show.Spec, "again", "", "", "movie") })
	run(func(i int) { tor.EnsureTitle(func() string { return "fallback" }) })
	for r := 0; r < 3; r++ {
		run(func(i int) {
			title, poster, category, data := tor.Meta()
			_, _, _, _ = title, poster, category, data
			_ = tor.Status()
		})
	}
	wg.Wait()

	title, _, category, _ := tor.Meta()
	if title == "" || category == "" {
		t.Fatalf("metadata lost: title=%q category=%q", title, category)
	}
}
