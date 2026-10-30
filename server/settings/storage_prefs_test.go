package settings

import (
	"strings"
	"testing"
)

func initStorageTest(t *testing.T) {
	t.Helper()
	oldPath, oldSets, oldRO, oldWA := Path, BTsets, ReadOnly, SearchWA
	t.Cleanup(func() {
		CloseDB()
		Path, BTsets, ReadOnly, SearchWA = oldPath, oldSets, oldRO, oldWA
	})
	Path = t.TempDir()
	ReadOnly = false
	if err := InitSets(false, false); err != nil {
		t.Fatal(err)
	}
}

// Switching the backend through /storage/settings must carry the data across.
func TestSetStoragePreferencesKeepsViewedAcrossBackends(t *testing.T) {
	initStorageTest(t)
	hash := strings.Repeat("cd", 20)
	BTsets.TrackTimecode = true
	SetViewed(&Viewed{Hash: hash, FileIndex: 3, TimeCode: 42})

	for _, pref := range []string{"json", "bbolt", "json"} {
		if err := SetStoragePreferences(map[string]interface{}{"viewed": pref}); err != nil {
			t.Fatalf("switch viewed to %s: %v", pref, err)
		}
		if BTsets.StoreViewedInJson != (pref == "json") {
			t.Fatalf("after %s: StoreViewedInJson=%v", pref, BTsets.StoreViewedInJson)
		}
		got := ListViewed(hash)
		if len(got) != 1 || got[0].FileIndex != 3 || got[0].TimeCode != 42 {
			t.Fatalf("after switch to %s viewed = %+v", pref, got)
		}
		if p := GetStoragePreferences(); p["viewed"] != pref {
			t.Fatalf("GetStoragePreferences viewed = %v, want %s", p["viewed"], pref)
		}
	}
}

func TestSetStoragePreferencesKeepsSettingsAcrossBackends(t *testing.T) {
	initStorageTest(t)
	BTsets.ConnectionsLimit = 77
	BTsets.RetrackersMode = 3
	SetBTSets(BTsets)

	for _, pref := range []string{"bbolt", "json"} {
		if err := SetStoragePreferences(map[string]interface{}{"settings": pref}); err != nil {
			t.Fatalf("switch settings to %s: %v", pref, err)
		}
		if BTsets.StoreSettingsInJson != (pref == "json") {
			t.Fatalf("after %s: StoreSettingsInJson=%v", pref, BTsets.StoreSettingsInJson)
		}
		// Re-read from the store the switch just selected.
		loadBTSets()
		if BTsets.ConnectionsLimit != 77 || BTsets.RetrackersMode != 3 {
			t.Fatalf("after %s settings lost: conns=%d retrackers=%d", pref, BTsets.ConnectionsLimit, BTsets.RetrackersMode)
		}
	}
}

func TestSetStoragePreferencesIgnoresNoopAndJunk(t *testing.T) {
	initStorageTest(t)
	wasSettings, wasViewed := BTsets.StoreSettingsInJson, BTsets.StoreViewedInJson

	for _, prefs := range []map[string]interface{}{
		nil,
		{},
		{"settings": "", "viewed": ""},
		{"settings": 1, "viewed": true}, // wrong types are skipped, not an error
		{"unknown": "json"},
	} {
		if err := SetStoragePreferences(prefs); err != nil {
			t.Errorf("%v: %v", prefs, err)
		}
		if BTsets.StoreSettingsInJson != wasSettings || BTsets.StoreViewedInJson != wasViewed {
			t.Errorf("%v changed storage: settings=%v viewed=%v", prefs, BTsets.StoreSettingsInJson, BTsets.StoreViewedInJson)
		}
	}
	// Asking for the backend already in use is a no-op.
	cur := map[string]interface{}{"settings": "bbolt", "viewed": "bbolt"}
	if wasSettings {
		cur["settings"] = "json"
	}
	if wasViewed {
		cur["viewed"] = "json"
	}
	if err := SetStoragePreferences(cur); err != nil {
		t.Error(err)
	}
}

func TestSetStoragePreferencesRefusals(t *testing.T) {
	initStorageTest(t)
	want := BTsets.StoreViewedInJson

	ReadOnly = true
	if err := SetStoragePreferences(map[string]interface{}{"viewed": "json"}); err == nil {
		t.Error("read-only mode accepted a switch")
	}
	if BTsets.StoreViewedInJson != want {
		t.Error("read-only refusal still changed the setting")
	}

	ReadOnly = false
	saved := BTsets
	BTsets = nil
	if err := SetStoragePreferences(map[string]interface{}{"viewed": "json"}); err == nil {
		t.Error("nil BTsets accepted a switch")
	}
	// GetStoragePreferences falls back to the defaults without settings.
	p := GetStoragePreferences()
	if p["settings"] != "json" || p["viewed"] != "bbolt" {
		t.Errorf("defaults = %v", p)
	}
	BTsets = saved
}
