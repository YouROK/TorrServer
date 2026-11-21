package settings

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testJsonDB(t *testing.T) (*JsonDB, string) {
	dir := t.TempDir()
	return &JsonDB{
		Path:              dir,
		filenameDelimiter: ".",
		filenameExtension: ".json",
		fileMode:          0o666,
		xPathDelimeter:    "/",
	}, filepath.Join(dir, "settings.json")
}

func TestJsonDBSetKeepsOtherEntries(t *testing.T) {
	db, path := testJsonDB(t)
	db.Set("Settings", "BitTorr", []byte(`{"CacheSize":67108864}`))
	db.Set("Settings", "Other", []byte(`{"a":1}`))

	if got := db.Get("Settings", "BitTorr"); !strings.Contains(string(got), "67108864") {
		t.Fatalf("BitTorr = %s after writing another entry", got)
	}
	if _, err := os.Stat(path + ".tmp"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("temporary file left behind: %v", err)
	}
}

// A truncated settings.json (killed mid-write, full disk) used to be read as an empty map,
// and the next Set wrote that map with one key over it. Now the damaged file is kept.
func TestJsonDBSetMovesDamagedFileAside(t *testing.T) {
	db, path := testJsonDB(t)
	damaged := []byte(`{"BitTorr": {"CacheSize": 1073741824}, "waf": {"Allow`)
	if err := os.WriteFile(path, damaged, 0o644); err != nil {
		t.Fatal(err)
	}

	if got := db.Get("Settings", "BitTorr"); got != nil {
		t.Fatalf("Get on a damaged file returned %s", got)
	}
	if kept, err := os.ReadFile(path); err != nil || !bytes.Equal(kept, damaged) {
		t.Fatalf("Get changed the damaged file: %v", err)
	}

	db.Set("Settings", "BitTorr", []byte(`{"CacheSize":67108864}`))

	if kept, err := os.ReadFile(path + ".bad"); err != nil || !bytes.Equal(kept, damaged) {
		t.Fatalf("damaged content not kept in settings.json.bad: %v", err)
	}
	if got := db.Get("Settings", "BitTorr"); !strings.Contains(string(got), "67108864") {
		t.Fatalf("BitTorr = %s after Set", got)
	}
}
