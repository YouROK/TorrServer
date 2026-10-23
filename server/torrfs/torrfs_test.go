package torrfs

import (
	"errors"
	"io"
	"io/fs"
	"sync"
	"testing"

	"server/settings"
	"server/torr"
)

func TestSanitizeName(t *testing.T) {
	cases := map[string]string{
		"":            "",
		".":           "",
		"..":          "",
		"  ..  ":      "",
		"a/b\\c":      "a_b_c",
		"ok name":     "ok name",
		"has\x00null": "hasnull",
		"  trim  ":    "trim",
	}
	for in, want := range cases {
		if got := SanitizeName(in); got != want {
			t.Errorf("SanitizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCategoryAndTorrentName(t *testing.T) {
	if categoryName("") != "other" || categoryName("a/b") != "a_b" || categoryName("movie") != "movie" {
		t.Fatalf("category %q %q", categoryName(""), categoryName("a/b"))
	}
	if got := torrentName(&torr.Torrent{Title: "Show/S01"}); got != "Show_S01" {
		t.Fatalf("title name %q", got)
	}
	if got := torrentName(&torr.Torrent{}); got == "" {
		t.Fatal("empty title should fall back to hash")
	}
}

func TestFileInfo(t *testing.T) {
	d := NewRootDir()
	fi, err := d.Stat()
	if err != nil || !fi.IsDir() || fi.Name() != "/" || fi.Size() != 4096 || fi.Sys() != nil {
		t.Fatalf("%v %#v", err, fi)
	}
	n, err := d.Read(make([]byte, 4))
	if n != 0 || !errors.Is(err, fs.ErrInvalid) {
		t.Fatalf("read %d %v", n, err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	cat := NewCategoryDir("")
	if cat.Name() != "other" || !cat.IsDir() {
		t.Fatalf("category %#v", cat.Name())
	}
}

func TestRootEmptyReadDir(t *testing.T) {
	torr.InitApiHelper(torr.NewBTS())
	old := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = old })

	root := NewRootDir()
	entries, err := root.ReadDir(-1)
	if err != nil || len(entries) != 0 {
		t.Fatalf("entries %v %v", entries, err)
	}
	f, err := AsFS(root).Open(".")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	if _, err := f.Read(make([]byte, 1)); !errors.Is(err, fs.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := AsFS(root).Open("../escape"); !errors.Is(err, fs.ErrInvalid) && !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("escape %v", err)
	}
}

func TestRootReadDirConcurrent(t *testing.T) {
	torr.InitApiHelper(torr.NewBTS())
	old := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = old })

	root := NewRootDir()
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			entries, err := root.ReadDir(0)
			if err != nil {
				t.Error(err)
			}
			_ = entries
			_, _ = io.ReadAll(root)
		}()
	}
	wg.Wait()
}
