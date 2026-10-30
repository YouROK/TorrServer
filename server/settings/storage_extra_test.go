package settings

import (
	"encoding/json"
	"testing"
)

func TestSetBTSetsAppliesFailsafes(t *testing.T) {
	initStorageTest(t)
	in := &BTSets{ReaderReadAHead: 1, PreloadCache: -5, UseDisk: true} // zero cache, limit, timeout; no save path
	SetBTSets(in)
	if BTsets != in {
		t.Fatal("BTsets not replaced")
	}
	if in.CacheSize != 64<<20 || in.ConnectionsLimit != 25 || in.TorrentDisconnectTimeout != 30 {
		t.Fatalf("defaults: cache=%d conns=%d timeout=%d", in.CacheSize, in.ConnectionsLimit, in.TorrentDisconnectTimeout)
	}
	if in.ReaderReadAHead != 5 || in.PreloadCache != 0 {
		t.Fatalf("clamp low: ra=%d preload=%d", in.ReaderReadAHead, in.PreloadCache)
	}
	if in.UseDisk {
		t.Fatal("UseDisk must be off without TorrentsSavePath")
	}

	hi := &BTSets{CacheSize: 1 << 20, ConnectionsLimit: 1, TorrentDisconnectTimeout: 1, ReaderReadAHead: 500, PreloadCache: 900}
	SetBTSets(hi)
	if hi.ReaderReadAHead != 100 || hi.PreloadCache != 100 || hi.CacheSize != 1<<20 {
		t.Fatalf("clamp high / keep: ra=%d preload=%d cache=%d", hi.ReaderReadAHead, hi.PreloadCache, hi.CacheSize)
	}

	// The row written to the DB is the clamped value.
	var stored BTSets
	if err := json.Unmarshal(tdb.Get("Settings", "BitTorr"), &stored); err != nil {
		t.Fatal(err)
	}
	if stored.ReaderReadAHead != 100 || stored.PreloadCache != 100 {
		t.Fatalf("stored %+v", stored)
	}
}

func TestSetBTSetsNilAndReadOnlyAreNoops(t *testing.T) {
	initStorageTest(t)
	before := BTsets
	SetBTSets(nil)
	if BTsets != before {
		t.Fatal("nil replaced BTsets")
	}
	ReadOnly = true
	SetBTSets(&BTSets{CacheSize: 1})
	if BTsets != before {
		t.Fatal("read-only replaced BTsets")
	}
}

func TestDetermineStoragePreferences(t *testing.T) {
	bbolt, js := newMemDB(), newMemDB()

	t.Run("defaults", func(t *testing.T) {
		t.Setenv("TS_SETTINGS_STORAGE", "")
		t.Setenv("TS_VIEWED_STORAGE", "")
		s, v := determineStoragePreferences(bbolt, js)
		if !s || v {
			t.Fatalf("settings json=%v viewed json=%v, want true/false", s, v)
		}
	})
	t.Run("env applies when no row exists", func(t *testing.T) {
		t.Setenv("TS_SETTINGS_STORAGE", "bbolt")
		t.Setenv("TS_VIEWED_STORAGE", "json")
		s, v := determineStoragePreferences(bbolt, js)
		if s || !v {
			t.Fatalf("settings json=%v viewed json=%v, want false/true", s, v)
		}
	})
	t.Run("stored row wins over env", func(t *testing.T) {
		t.Setenv("TS_SETTINGS_STORAGE", "bbolt")
		buf, _ := json.Marshal(&BTSets{StoreSettingsInJson: true, StoreViewedInJson: true})
		js.Set("Settings", "BitTorr", buf)
		s, v := determineStoragePreferences(bbolt, js)
		if !s || !v {
			t.Fatalf("settings json=%v viewed json=%v, want true/true", s, v)
		}
	})
}

func TestLoadExistingSettings(t *testing.T) {
	bbolt, js := newMemDB(), newMemDB()
	if loadExistingSettings(bbolt, js) != nil {
		t.Fatal("settings from nothing")
	}

	buf, _ := json.Marshal(&BTSets{ConnectionsLimit: 11})
	bbolt.Set("Settings", "BitTorr", buf)
	if got := loadExistingSettings(bbolt, js); got == nil || got.ConnectionsLimit != 11 {
		t.Fatalf("bbolt fallback: %+v", got)
	}

	jbuf, _ := json.Marshal(&BTSets{ConnectionsLimit: 22})
	js.Set("Settings", "BitTorr", jbuf)
	if got := loadExistingSettings(bbolt, js); got == nil || got.ConnectionsLimit != 22 {
		t.Fatalf("json first: %+v", got)
	}

	js.Set("Settings", "BitTorr", []byte(`{broken`))
	if got := loadExistingSettings(bbolt, js); got == nil || got.ConnectionsLimit != 11 {
		t.Fatalf("corrupt json falls back to bbolt: %+v", got)
	}
	bbolt.Set("Settings", "BitTorr", []byte(`{broken`))
	if loadExistingSettings(bbolt, js) != nil {
		t.Fatal("both corrupt must give nil")
	}
}

func TestMigrateMissingCopiesBothWays(t *testing.T) {
	a, b := newMemDB(), newMemDB()
	a.Set("Settings", "BitTorr", []byte(`{"a":1}`))
	if err := migrateMissing(a, b, "Settings", "BitTorr"); err != nil {
		t.Fatal(err)
	}
	if string(b.Get("Settings", "BitTorr")) != `{"a":1}` {
		t.Fatal("a -> b")
	}

	c, d := newMemDB(), newMemDB()
	d.Set("Settings", "BitTorr", []byte(`{"d":1}`))
	if err := migrateMissing(c, d, "Settings", "BitTorr"); err != nil {
		t.Fatal(err)
	}
	if string(c.Get("Settings", "BitTorr")) != `{"d":1}` {
		t.Fatal("d -> c")
	}

	// A value present on both sides is never overwritten.
	e, f := newMemDB(), newMemDB()
	e.Set("Settings", "BitTorr", []byte(`{"e":1}`))
	f.Set("Settings", "BitTorr", []byte(`{"f":1}`))
	_ = migrateMissing(e, f, "Settings", "BitTorr")
	if string(e.Get("Settings", "BitTorr")) != `{"e":1}` || string(f.Get("Settings", "BitTorr")) != `{"f":1}` {
		t.Fatal("existing values overwritten")
	}
}

func TestSafeMigrateClearsSourceOnlyWhenAsked(t *testing.T) {
	src, dst := newMemDB(), newMemDB()
	src.Set("Settings", "BitTorr", []byte(`{"a":1}`))
	safeMigrate(src, dst, "Settings", "BitTorr", "Target", false)
	if dst.Get("Settings", "BitTorr") == nil || src.Get("Settings", "BitTorr") == nil {
		t.Fatal("copy without clear")
	}

	src2, dst2 := newMemDB(), newMemDB()
	src2.Set("Settings", "BitTorr", []byte(`{"a":1}`))
	safeMigrate(src2, dst2, "Settings", "BitTorr", "Target", true)
	if dst2.Get("Settings", "BitTorr") == nil || src2.Get("Settings", "BitTorr") != nil {
		t.Fatal("move with clear")
	}

	safeMigrate(newMemDB(), newMemDB(), "Settings", "missing", "Target", true) // nothing to do, no panic
}

func TestSafeMigrateAllClearsSource(t *testing.T) {
	src, dst := newMemDB(), newMemDB()
	src.Set("Viewed", "h1", []byte(`{"1":1}`))
	src.Set("Viewed", "h2", []byte(`{"2":1}`))
	safeMigrateAll(src, dst, "Viewed", "Target", true)
	if len(dst.List("Viewed")) != 2 || len(src.List("Viewed")) != 0 {
		t.Fatalf("dst=%v src=%v", dst.List("Viewed"), src.List("Viewed"))
	}

	keep, dst2 := newMemDB(), newMemDB()
	keep.Set("Viewed", "h1", []byte(`{"1":1}`))
	safeMigrateAll(keep, dst2, "Viewed", "Target", false)
	if len(keep.List("Viewed")) != 1 || len(dst2.List("Viewed")) != 1 {
		t.Fatal("copy without clear")
	}
	safeMigrateAll(newMemDB(), newMemDB(), "Viewed", "Target", true) // empty source
}

func TestIsByteArraysEqualJsonOptimized(t *testing.T) {
	tests := []struct {
		name string
		a, b []byte
		want bool
	}{
		{"both nil", nil, nil, true},
		{"both empty", []byte{}, []byte{}, true},
		{"identical", []byte(`{"a":1}`), []byte(`{"a":1}`), true},
		{"different length", []byte(`{"a":1}`), []byte(`{"a":10}`), false},
		{"same length different bytes", []byte(`{"a":1}`), []byte(`{"a":2}`), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isByteArraysEqualJsonOptimized(tt.a, tt.b)
			if err != nil || got != tt.want {
				t.Fatalf("got %v, %v; want %v", got, err, tt.want)
			}
		})
	}
}

func TestInitSetsHonoursStorageEnvOnFreshDir(t *testing.T) {
	t.Setenv("TS_SETTINGS_STORAGE", "bbolt")
	t.Setenv("TS_VIEWED_STORAGE", "json")
	initStorageTest(t)
	if BTsets.StoreSettingsInJson || !BTsets.StoreViewedInJson {
		t.Fatalf("StoreSettingsInJson=%v StoreViewedInJson=%v, want false/true", BTsets.StoreSettingsInJson, BTsets.StoreViewedInJson)
	}
	if p := GetStoragePreferences(); p["settings"] != "bbolt" || p["viewed"] != "json" {
		t.Fatalf("prefs %v", p)
	}
}
