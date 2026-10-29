package torr

import (
	"runtime"
	"testing"
	"time"

	"server/settings"
	"server/torr/offlinetest"
	"server/torr/state"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func addDBRow(t *testing.T, title string) metainfo.Hash {
	t.Helper()
	hash := metainfo.NewHashFromHex("0123456789abcdef0123456789abcdef01234567")
	AddTorrentDB(&Torrent{
		TorrentSpec: &torrent.TorrentSpec{InfoHash: hash, Trackers: [][]string{{"http://t1/announce"}}},
		Title:       title,
		Size:        1 << 20,
	})
	return hash
}

// GotInfo on a library row closes it, and NewTorrent appends a tracker tier to the
// spec it is given. Neither may leak into the next caller's row: /stream only adds
// the torrent when the row still reads TorrentInDB.
func TestDBRowsAreNotShared(t *testing.T) {
	offlinetest.Prepare(t)
	hash := addDBRow(t, "Movie")

	if ListTorrentsDB()[hash].GotInfo() {
		t.Fatal("GotInfo on a library-only row reported info")
	}
	if st := ListTorrentsDB()[hash].Status().Stat; st != state.TorrentInDB {
		t.Fatalf("row stat after GotInfo on another copy = %v, want TorrentInDB", st)
	}

	row := GetTorrentDB(hash)
	addExtraTrackers(row.TorrentSpec, []string{"http://t2/announce"})
	row.Trackers[0] = nil
	if got := GetTorrentDB(hash).Trackers; len(got) != 1 || len(got[0]) != 1 {
		t.Fatalf("cached spec trackers changed through a returned row: %v", got)
	}
}

func TestDBRowsFollowWrites(t *testing.T) {
	offlinetest.Prepare(t)
	hash := addDBRow(t, "Movie")
	if GetTorrentDB(hash).Title != "Movie" {
		t.Fatal("row missing after add")
	}

	addDBRow(t, "Renamed")
	if got := ListTorrentsDB()[hash].Title; got != "Renamed" {
		t.Fatalf("title after update = %q", got)
	}

	RemTorrentDB(hash)
	if GetTorrentDB(hash) != nil || len(ListTorrentsDB()) != 0 {
		t.Fatal("row still listed after remove")
	}
}

// A list poll must not unmarshal rows or zlib-pack the torrs token again.
// flate's BestCompression writer alone allocates hundreds of KiB.
func TestDBListPollAllocs(t *testing.T) {
	offlinetest.Prepare(t)
	hash := addDBRow(t, "Movie")
	if ListTorrentsDB()[hash].Status().TorrsHash == "" {
		t.Fatal("no torrs token")
	}

	const polls = 50
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for range polls {
		ListTorrentsDB()[hash].Status()
	}
	runtime.ReadMemStats(&after)
	if perPoll := (after.TotalAlloc - before.TotalAlloc) / polls; perPoll > 32<<10 {
		t.Fatalf("list poll allocates %d B, want <= 32KiB (token repacked or rows decoded per poll)", perPoll)
	}
}

func TestExpiredWaitsForPendingInfo(t *testing.T) {
	offlinetest.Prepare(t)
	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(bt.Disconnect)

	// No peers, no trackers, no DHT: metadata never arrives.
	tor, err := NewTorrent(&torrent.TorrentSpec{InfoHash: metainfo.NewHashFromHex("89abcdef0123456789abcdef0123456789abcdef")}, bt)
	if err != nil {
		t.Fatal(err)
	}
	tor.muTorrent.Lock()
	tor.expiredTime = time.Now().Add(-time.Second)
	tor.muTorrent.Unlock()

	if !tor.expired() {
		t.Fatal("idle TorrentAdded torrent with nil cache past its expiry must expire")
	}

	done := make(chan bool)
	go func() { done <- tor.WaitInfo() }()
	for tor.infoWaiters.Load() == 0 {
		runtime.Gosched()
	}
	if tor.expired() {
		t.Fatal("torrent expired while WaitInfo was still waiting for metadata")
	}

	tor.Close()
	if <-done {
		t.Fatal("WaitInfo reported info after Close")
	}
	if !tor.expired() {
		t.Fatal("closed torrent past its expiry must expire once nobody waits")
	}

	tor.muTorrent.Lock()
	tor.Stat = state.TorrentPreload
	tor.muTorrent.Unlock()
	if tor.expired() {
		t.Fatal("preloading torrent must not expire")
	}
}

// One add or remove patches the cached rows instead of decoding the library again,
// and the cache stays equal to what settings.ListTorrent reads from the DB.
func TestDBRowsPatchedInPlace(t *testing.T) {
	offlinetest.Prepare(t)
	row := func(b byte, ts int64) *settings.TorrentDB {
		return &settings.TorrentDB{TorrentSpec: &torrent.TorrentSpec{InfoHash: metainfo.Hash{b}}, Title: string('a' + b), Timestamp: ts}
	}
	settings.AddTorrent(row(1, 10))
	settings.AddTorrent(row(2, 30))
	before := listDBRows()
	if len(before) != 2 {
		t.Fatalf("rows %d, want 2", len(before))
	}

	settings.AddTorrent(row(3, 20))
	edited := row(1, 10)
	edited.Title = "edited"
	settings.AddTorrent(edited)
	after := listDBRows()
	if after[0] != before[0] {
		t.Fatal("an untouched row was decoded again")
	}
	want := settings.ListTorrent()
	if len(after) != len(want) {
		t.Fatalf("cache has %d rows, DB has %d", len(after), len(want))
	}
	for i := range want {
		if after[i].InfoHash != want[i].InfoHash || after[i].Title != want[i].Title {
			t.Fatalf("row %d: cache %v %q, DB %v %q", i, after[i].InfoHash, after[i].Title, want[i].InfoHash, want[i].Title)
		}
	}

	settings.RemTorrent(metainfo.Hash{2})
	if got := ListTorrentsDB(); len(got) != 2 || got[metainfo.Hash{2}] != nil {
		t.Fatalf("removed row still listed: %d rows", len(got))
	}
}
