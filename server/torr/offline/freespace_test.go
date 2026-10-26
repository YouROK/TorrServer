package offline

import (
	"path/filepath"
	"testing"
)

func TestFreeSpaceOnExistingDir(t *testing.T) {
	dir := t.TempDir()
	free, err := FreeSpace(dir)
	if err != nil {
		t.Fatalf("FreeSpace(%q): %v", dir, err)
	}
	if free <= 0 {
		t.Fatalf("FreeSpace(%q) = %d, want > 0", dir, free)
	}
}

func TestFreeSpaceOnMissingDirFails(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	if free, err := FreeSpace(dir); err == nil {
		t.Fatalf("FreeSpace(%q) = %d, want an error for a missing directory", dir, free)
	}
}
