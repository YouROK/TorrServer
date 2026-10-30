package models

import "testing"

func TestGetNames(t *testing.T) {
	if got := (TorrentDetails{}).GetNames(); got != "" {
		t.Fatalf("no names = %q", got)
	}
	d := TorrentDetails{Names: []string{"Матрица", "The Matrix"}}
	if got := d.GetNames(); got != "Матрица The Matrix" {
		t.Fatalf("GetNames = %q", got)
	}
}
