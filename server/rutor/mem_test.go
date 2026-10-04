package rutor

import (
	"compress/flate"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"server/rutor/models"
)

// openRutorLs opens the rutor.ls index next to the tests and returns a
// decoder positioned inside its top-level array. rutor.ls is downloaded at
// runtime and isn't in the repo, so the test is skipped when it's missing.
func openRutorLs(t *testing.T) *json.Decoder {
	t.Helper()
	path, _ := os.Getwd()
	ff, err := os.Open(filepath.Join(path, "rutor.ls"))
	if errors.Is(err, fs.ErrNotExist) {
		t.Skip("rutor.ls not present (downloaded at runtime)")
	}
	if err != nil {
		t.Fatal(err)
	}
	r := flate.NewReader(ff)
	t.Cleanup(func() {
		r.Close()
		ff.Close()
	})
	dec := json.NewDecoder(r)
	if _, err := dec.Token(); err != nil {
		t.Fatal(err)
	}
	return dec
}

func TestParseChannel(t *testing.T) {
	dec := openRutorLs(t)

	channel := make(chan *models.TorrentDetails)
	done := make(chan struct{})
	var ftors []*models.TorrentDetails
	go func() {
		defer close(done)
		for torr := range channel {
			ftors = append(ftors, torr)
		}
	}()

	for dec.More() {
		var torr *models.TorrentDetails
		if err := dec.Decode(&torr); err != nil {
			close(channel)
			t.Fatal(err)
		}
		channel <- torr
	}
	close(channel)
	<-done
	t.Logf("parsed %d torrents", len(ftors))
}

func TestParseArr(t *testing.T) {
	dec := openRutorLs(t)

	var ftors []*models.TorrentDetails
	for dec.More() {
		var torr *models.TorrentDetails
		if err := dec.Decode(&torr); err != nil {
			t.Fatal(err)
		}
		ftors = append(ftors, torr)
	}
	t.Logf("parsed %d torrents", len(ftors))
}
