package settings

import (
	"encoding/binary"
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	bolt "go.etcd.io/bbolt"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func TestInitSetsCRUDAndReadOnly(t *testing.T) {
	dir := t.TempDir()
	oldPath, oldSets, oldRO, oldWA, oldArgs := Path, BTsets, ReadOnly, SearchWA, Args
	t.Cleanup(func() {
		CloseDB()
		Path, BTsets, ReadOnly, SearchWA, Args = oldPath, oldSets, oldRO, oldWA, oldArgs
	})
	Path = dir
	ReadOnly = false
	if err := InitSets(false, false); err != nil {
		t.Fatal(err)
	}
	if BTsets == nil || BTsets.CacheSize == 0 {
		t.Fatalf("defaults %#v", BTsets)
	}

	hash := strings.Repeat("ab", 20)
	spec := &torrent.TorrentSpec{InfoHash: metainfo.NewHashFromHex(hash), DisplayName: "Show"}
	AddTorrent(&TorrentDB{TorrentSpec: spec, Title: "Show", Category: "tv", Timestamp: 1})
	list := ListTorrent()
	if len(list) != 1 || list[0].Title != "Show" {
		t.Fatalf("list %#v", list)
	}
	SetViewed(&Viewed{Hash: hash, FileIndex: 1, TimeCode: 9})
	SetViewed(&Viewed{Hash: hash, FileIndex: 2, TimeCode: 3})
	got := ListViewed(hash)
	if len(got) != 2 {
		t.Fatalf("viewed %#v", got)
	}
	if len(ListViewed("")) < 1 {
		t.Fatal("all viewed")
	}
	RemViewed(&Viewed{Hash: hash, FileIndex: 1})
	if len(ListViewed(hash)) != 1 {
		t.Fatal(ListViewed(hash))
	}
	RemViewed(&Viewed{Hash: hash, FileIndex: -1})
	if len(ListViewed(hash)) != 0 {
		t.Fatal("viewed left")
	}
	RemTorrent(spec.InfoHash)
	if len(ListTorrent()) != 0 {
		t.Fatal("torrent left")
	}

	saved := BTsets.CacheSize
	ReadOnly = true
	changed := *BTsets
	changed.CacheSize = 1
	SetBTSets(&changed)
	ReadOnly = false
	if BTsets.CacheSize != saved {
		t.Fatalf("readonly write landed: %d", BTsets.CacheSize)
	}
	SetBTSets(BTsets)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			h := hash
			SetViewed(&Viewed{Hash: h, FileIndex: i, TimeCode: float64(i)})
			_ = ListViewed(h)
			_ = ListTorrent()
		}(i)
	}
	wg.Wait()
}

func TestMigrationHelpers(t *testing.T) {
	src, dst := newMemDB(), newMemDB()
	if _, err := isByteArraysEqualJson(nil, nil); err != nil {
		t.Fatal(err)
	}
	eq, err := isByteArraysEqualJson([]byte(`{"b":1,"a":2}`), []byte(`{"a":2,"b":1}`))
	if err != nil || !eq {
		t.Fatalf("reorder %v %v", eq, err)
	}
	eq, err = isByteArraysEqualJson([]byte(`{"a":1}`), []byte(`{"a":2}`))
	if err != nil || eq {
		t.Fatalf("diff %v %v", eq, err)
	}
	if eq, _ = isByteArraysEqualJson(nil, []byte(`{}`)); eq {
		t.Fatal("empty vs json")
	}

	payload, _ := json.Marshal(map[string]int{"CacheSize": 3})
	src.Set("Settings", "BitTorr", payload)
	changed, err := MigrateSingle(src, dst, "Settings", "BitTorr")
	if err != nil || !changed || string(dst.Get("Settings", "BitTorr")) != string(payload) {
		t.Fatalf("migrate %v %v", changed, err)
	}
	changed, err = MigrateSingle(src, dst, "Settings", "BitTorr")
	if err != nil || changed {
		t.Fatalf("second %v %v", changed, err)
	}
	if _, err := MigrateSingle(src, dst, "Settings", "missing"); err != nil {
		t.Fatal(err)
	}
	n, skipped, err := MigrateAll(src, dst, "Settings")
	if err != nil || n+skipped < 1 {
		t.Fatalf("all %d %d %v", n, skipped, err)
	}

	src.Set("Viewed", "h1", []byte(`{"1":9}`))
	dst.Set("Viewed", "h2", []byte(`{"2":4}`))
	if err := syncViewedSimple(src, dst); err != nil {
		t.Fatal(err)
	}
	if src.Get("Viewed", "h2") == nil || dst.Get("Viewed", "h1") == nil {
		t.Fatal("viewed not synced")
	}
	if mergeViewedDataSimple(nil, nil) != nil {
		t.Fatal("nil merge")
	}
	if string(mergeViewedDataSimple(nil, []byte(`{"1":1}`))) != `{"1":1}` {
		t.Fatal("one side")
	}

	if err := SmartMigrate(src, dst, "nope"); err == nil {
		t.Fatal("unknown direction")
	}
	if err := SmartMigrate(src, dst, "settings_to_json"); err != nil {
		t.Fatal(err)
	}
	if err := SmartMigrate(src, dst, "sync_both"); err != nil {
		t.Fatal(err)
	}
	var eight [8]byte
	binary.BigEndian.PutUint64(eight[:], 7)
	if b2i(eight[:]) != 7 {
		t.Fatal(b2i(eight[:]))
	}
	if err := MigrateSettingsFromJson(src, dst); err != nil {
		t.Fatal(err)
	}
	if err := MigrateViewedToJson(src, dst); err != nil {
		t.Fatal(err)
	}
	if err := MigrateViewedFromJson(dst, src); err != nil {
		t.Fatal(err)
	}
}

func TestStorageSwitchesAndLoad(t *testing.T) {
	dir := t.TempDir()
	oldPath, oldSets, oldRO, oldWA := Path, BTsets, ReadOnly, SearchWA
	t.Cleanup(func() {
		CloseDB()
		Path, BTsets, ReadOnly, SearchWA = oldPath, oldSets, oldRO, oldWA
	})
	Path = dir
	ReadOnly = false
	if err := InitSets(false, false); err != nil {
		t.Fatal(err)
	}
	if BTsets.String() == "" {
		t.Fatal("string")
	}
	if GetStoragePreferences()["settings"] == nil {
		t.Fatal("prefs")
	}
	ReadOnly = true
	if err := SwitchSettingsStorage(true); err == nil {
		t.Fatal("readonly settings switch")
	}
	if err := SwitchViewedStorage(false); err == nil {
		t.Fatal("readonly viewed switch")
	}
	ReadOnly = false
	if err := SwitchSettingsStorage(false); err != nil {
		t.Fatal(err)
	}
	if err := SwitchSettingsStorage(true); err != nil {
		t.Fatal(err)
	}
	if err := SwitchViewedStorage(true); err != nil {
		t.Fatal(err)
	}
	if err := SwitchViewedStorage(false); err != nil {
		t.Fatal(err)
	}
	tdb.Set("Settings", "BitTorr", []byte(`{"CacheSize":1048576,"ReaderReadAHead":10}`))
	loadBTSets()
	if !BTsets.EnableBonjour || BTsets.DefaultTrackers == "" {
		t.Fatalf("upgraded %#v", BTsets)
	}
	if tdb != nil {
		tdb.Clear("Viewed")
	}
	js := NewJsonDB()
	js.CloseDB()
	MigrateTorrents()

	legacy, err := bolt.Open(filepath.Join(dir, "torrserver.db"), 0o666, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = legacy.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket(dbTorrentsName)
		if err != nil {
			return err
		}
		h, err := b.CreateBucket([]byte(strings.Repeat("ab", 20)))
		if err != nil {
			return err
		}
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], 10)
		if err := h.Put([]byte("Name"), []byte("Old")); err != nil {
			return err
		}
		if err := h.Put([]byte("Link"), []byte("magnet:?xt=urn:btih:"+strings.Repeat("ab", 20))); err != nil {
			return err
		}
		if err := h.Put([]byte("Size"), size[:]); err != nil {
			return err
		}
		return h.Put([]byte("Timestamp"), size[:])
	})
	_ = legacy.Close()
	if err != nil {
		t.Fatal(err)
	}
	MigrateTorrents()
	if len(ListTorrent()) != 1 {
		t.Fatalf("migrated %#v", ListTorrent())
	}
}
