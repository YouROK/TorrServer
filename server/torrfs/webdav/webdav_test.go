package webdav

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestCleanWebDAVPath(t *testing.T) {
	cases := map[string]string{
		"":       ".",
		"/":      ".",
		"a":      "a",
		"/a/b":   "a/b",
		"a/../b": "b",
	}
	for in, want := range cases {
		if got := cleanWebDAVPath(in); got != want {
			t.Errorf("cleanWebDAVPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestReadOnlyFS(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	ro := &ReadOnlyFS{FS: os.DirFS(dir)}
	ctx := context.Background()
	if err := ro.Mkdir(ctx, "x", 0o755); !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}
	if err := ro.RemoveAll(ctx, "a.txt"); !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}
	if err := ro.Rename(ctx, "a.txt", "b.txt"); !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}
	if _, err := ro.OpenFile(ctx, "a.txt", os.O_WRONLY, 0); !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}
	if _, err := ro.OpenFile(ctx, "missing", os.O_RDONLY, 0); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}

	f, err := ro.OpenFile(ctx, "/a.txt", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 8)
	n, err := f.Read(buf)
	if n != 5 || (err != nil && err != io.EOF) || string(buf[:n]) != "hello" {
		t.Fatalf("read %d %v %q", n, err, buf[:n])
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("x")); !errors.Is(err, fs.ErrPermission) {
		t.Fatal(err)
	}
	fi, err := f.Stat()
	if err != nil || fi.Name() != "a.txt" || fi.IsDir() {
		t.Fatalf("stat %#v %v", fi, err)
	}
	if _, err := f.Readdir(-1); !errors.Is(err, fs.ErrInvalid) {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Read(buf); !errors.Is(err, fs.ErrClosed) {
		t.Fatal(err)
	}

	dirf, err := ro.OpenFile(ctx, ".", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	infos, err := dirf.Readdir(1)
	if err != nil || len(infos) != 1 {
		t.Fatalf("readdir %v %v", infos, err)
	}
	rest, err := dirf.Readdir(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos)+len(rest) != 2 {
		t.Fatalf("listed %d + %d", len(infos), len(rest))
	}
	if _, err := dirf.Readdir(1); err != io.EOF {
		t.Fatalf("second pass %v", err)
	}
	_ = dirf.Close()
}

func TestReadOnlyConcurrent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	ro := &ReadOnlyFS{FS: os.DirFS(dir)}
	f, err := ro.OpenFile(context.Background(), "a.txt", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 4)
			_, _ = f.Read(buf)
			_, _ = f.Stat()
		}()
	}
	wg.Wait()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
