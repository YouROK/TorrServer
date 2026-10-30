package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestClearStr(t *testing.T) {
	cases := map[string]string{
		"":                        "",
		"The Matrix (1999)":       "thematrix1999",
		"Ёжик в Тумане":           "ёжиквтумане",
		"Spider-Man: No Way Home": "spidermannowayhome",
		"日本語 abc":                 "abc",
		"!!!":                     "",
	}
	for in, want := range cases {
		if got := ClearStr(in); got != want {
			t.Errorf("ClearStr(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMD5FileMissingIsEmpty(t *testing.T) {
	if got := MD5File(filepath.Join(t.TempDir(), "nope")); got != "" {
		t.Fatalf("missing file hash = %q, want empty", got)
	}
}

// updateDB compares MD5File of the old and new index, so a missing original
// must never equal a real download.
func TestMD5FileContent(t *testing.T) {
	// larger than the 1MiB read buffer to cover the multi-read loop
	data := make([]byte, 1<<20+123)
	for i := range data {
		data[i] = byte(i * 7)
	}
	p := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	if got, want := MD5File(p), hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("hash = %s, want %s", got, want)
	}

	empty := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if MD5File(empty) == "" {
		t.Fatal("empty file must hash to a non-empty digest, distinct from a missing file")
	}
}
