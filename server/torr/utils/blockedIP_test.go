package utils

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"server/settings"
)

func TestReadBlockedIP(t *testing.T) {
	old := settings.Path
	dir := t.TempDir()
	settings.Path = dir
	t.Cleanup(func() { settings.Path = old })

	if _, err := ReadBlockedIP(); err == nil {
		t.Fatal("missing file should fail")
	}
	body := "# comment\n\nExample:1.2.3.0-1.2.3.255\n"
	if err := os.WriteFile(filepath.Join(dir, "blocklist"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ranger, err := ReadBlockedIP()
	if err != nil {
		t.Fatal(err)
	}
	if ranger == nil || ranger.NumRanges() != 1 {
		t.Fatalf("ranges %v", ranger)
	}
	if _, ok := ranger.Lookup(net.ParseIP("1.2.3.10")); !ok {
		t.Fatal("expected 1.2.3.10 to be blocked")
	}
	if _, ok := ranger.Lookup(net.ParseIP("8.8.8.8")); ok {
		t.Fatal("8.8.8.8 should not match")
	}
}
