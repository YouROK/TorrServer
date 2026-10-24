//go:build !windows && !ios

package fuse

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"syscall"
	"testing"
	"time"

	gofuse "github.com/hanwen/go-fuse/v2/fuse"

	"server/settings"
)

type fakeInfo struct {
	name string
	size int64
	mode fs.FileMode
	dir  bool
	mt   time.Time
}

func (f fakeInfo) Name() string       { return f.name }
func (f fakeInfo) Size() int64        { return f.size }
func (f fakeInfo) Mode() fs.FileMode  { return f.mode }
func (f fakeInfo) ModTime() time.Time { return f.mt }
func (f fakeInfo) IsDir() bool        { return f.dir }
func (f fakeInfo) Sys() any           { return nil }

func TestFuseAttrHelpers(t *testing.T) {
	mt := time.Unix(100, 50)
	dir := fakeInfo{name: "d", size: 1, mode: 0o755, dir: true, mt: mt}
	mode := fuseModeFromInfo(dir)
	if mode&gofuse.S_IFDIR == 0 || mode&0o755 != 0o755 {
		t.Fatalf("dir mode %o", mode)
	}
	file := fakeInfo{name: "f", size: 9, mode: 0o644, mt: mt}
	var attr gofuse.Attr
	fillAttr(&attr, file)
	if attr.Size != 9 || attr.Mode&gofuse.S_IFREG == 0 || attr.Mtime != 100 || attr.Mtimensec != 50 {
		t.Fatalf("%#v", attr)
	}
	if attr.Atime != attr.Mtime || attr.Ctime != attr.Mtime {
		t.Fatalf("times %#v", attr)
	}
	var zero gofuse.Attr
	fillAttr(&zero, fakeInfo{mode: 0o644})
	if zero.Mtime == 0 {
		t.Fatal("zero mtime should be replaced")
	}
	var d gofuse.Attr
	fillAttr(&d, dir)
	if d.Size != 4096 {
		t.Fatalf("dir size %d", d.Size)
	}
}

func TestErrno(t *testing.T) {
	if errno(nil) != 0 {
		t.Fatal("nil")
	}
	if errno(fs.ErrNotExist) != syscall.ENOENT {
		t.Fatal("not exist")
	}
	if errno(fs.ErrPermission) != syscall.EPERM || errno(fs.ErrInvalid) != syscall.EINVAL {
		t.Fatal("perm/invalid")
	}
	if errno(&fs.PathError{Err: fs.ErrNotExist}) != syscall.ENOENT {
		t.Fatal("path error")
	}
	if errno(errors.New("other")) != syscall.EIO {
		t.Fatal("other")
	}
}

func TestFuseAutoMountNoPath(t *testing.T) {
	old := settings.Args
	settings.Args = &settings.ExecArgs{}
	t.Cleanup(func() { settings.Args = old })
	FuseAutoMount()
	FuseCleanup()
	if GetFuseFS() == nil || GetFuseFS().GetMountPath() != "" {
		t.Fatal("fuse fs")
	}
}

func TestNewFuseFSDisabled(t *testing.T) {
	ffs := NewFuseFS()
	if ffs.GetMountPath() != "" {
		t.Fatal(ffs.GetMountPath())
	}
	if err := ffs.Unmount(); err == nil {
		t.Fatal("unmount of a fresh filesystem should fail")
	}
}

func TestNodeOpsInMemory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/hello.txt", []byte("abcdef"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir+"/sub", 0o755); err != nil {
		t.Fatal(err)
	}
	tree := os.DirFS(dir)
	ctx := context.Background()

	root := &FuseFS{tfs: tree, p: "."}
	root.OnAdd(ctx)
	var attr gofuse.AttrOut
	if errno := root.Getattr(ctx, nil, &attr); errno != 0 || attr.Mode&gofuse.S_IFDIR == 0 {
		t.Fatalf("root getattr %v", errno)
	}
	stream, errno := root.Readdir(ctx)
	if errno != 0 || stream == nil {
		t.Fatalf("root readdir %v", errno)
	}
	var names []string
	for stream.HasNext() {
		de, errn := stream.Next()
		if errn != 0 {
			t.Fatal(errn)
		}
		names = append(names, de.Name)
	}
	stream.Close()
	if len(names) != 2 {
		t.Fatalf("names %v", names)
	}
	if _, errno := root.Lookup(ctx, "missing", &gofuse.EntryOut{}); errno == 0 {
		t.Fatal("missing lookup")
	}

	node := &tfsNode{tfs: tree, p: "hello.txt"}
	if errno := node.Getattr(ctx, nil, &attr); errno != 0 || attr.Size != 6 {
		t.Fatalf("file getattr size %d err %v", attr.Size, errno)
	}
	if _, errno := node.Readdir(ctx); errno == 0 {
		t.Fatal("readdir on a file")
	}
	if _, errno := node.Lookup(ctx, "nope", &gofuse.EntryOut{}); errno == 0 {
		t.Fatal("child lookup")
	}
	if _, _, errno := node.Open(ctx, gofuse.O_ANYWRITE); errno != syscall.EROFS {
		t.Fatalf("write open %v", errno)
	}
	fh, flags, errno := node.Open(ctx, 0)
	if errno != 0 || flags&gofuse.FOPEN_DIRECT_IO == 0 {
		t.Fatalf("open %v %o", errno, flags)
	}
	h := fh.(*tfsHandle)
	dest := make([]byte, 4)
	res, errno := h.Read(ctx, dest, 1)
	if errno != 0 {
		t.Fatal(errno)
	}
	buf, status := res.Bytes(nil)
	if status != 0 || string(buf) != "bcde" {
		t.Fatalf("read %q %v", buf, status)
	}
	if errno := h.Release(ctx); errno != 0 {
		t.Fatal(errno)
	}

	dirNode := &tfsNode{tfs: tree, p: "."}
	ds, errno := dirNode.Readdir(ctx)
	if errno != 0 {
		t.Fatal(errno)
	}
	n := 0
	for ds.HasNext() {
		if _, errn := ds.Next(); errn != 0 {
			t.Fatal(errn)
		}
		n++
	}
	ds.Close()
	if n != 2 {
		t.Fatalf("dir entries %d", n)
	}
	if node.full("x") == "" {
		t.Fatal("full")
	}
	_ = io.EOF
}

func TestMountWithoutKernel(t *testing.T) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = prev })
	ffs := NewFuseFS()
	ffs.enabled = true
	if err := ffs.Mount(t.TempDir()); err == nil {
		t.Fatal("second mount")
	}
	ffs.enabled = false
	err := ffs.Mount(t.TempDir())
	if err == nil {
		_ = ffs.Unmount()
		t.Fatal("kernel mount succeeded; expected a sandbox failure")
	}
}

type memFile struct {
	b        []byte
	off      int
	seekFail bool
}

func (f *memFile) Read(p []byte) (int, error) {
	if f.off >= len(f.b) {
		return 0, io.EOF
	}
	n := copy(p, f.b[f.off:])
	f.off += n
	return n, nil
}
func (f *memFile) Seek(int64, int) (int64, error) {
	if f.seekFail {
		return 0, errors.New("seek")
	}
	return 0, nil
}
func (f *memFile) Close() error { return nil }
func (f *memFile) Stat() (fs.FileInfo, error) {
	return fakeInfo{name: "a", size: int64(len(f.b)), mode: 0o644}, nil
}

type noSeekFile struct{}

func (noSeekFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (noSeekFile) Close() error               { return nil }
func (noSeekFile) Stat() (fs.FileInfo, error) { return fakeInfo{name: "a", mode: 0o644}, nil }

type tinyFS struct{ kind string }

func (t tinyFS) Open(string) (fs.File, error) {
	switch t.kind {
	case "noseek":
		return noSeekFile{}, nil
	case "seekerr":
		return &memFile{b: []byte("hi"), seekFail: true}, nil
	default:
		return nil, fs.ErrNotExist
	}
}

func TestHandleErrors(t *testing.T) {
	ctx := context.Background()
	node := &tfsNode{tfs: tinyFS{kind: "noseek"}, p: "a"}
	if _, _, errno := node.Open(ctx, 0); errno != syscall.ENOSYS {
		t.Fatalf("noseek %v", errno)
	}
	node.tfs = tinyFS{kind: "seekerr"}
	fh, _, errno := node.Open(ctx, 0)
	if errno != 0 {
		t.Fatal(errno)
	}
	if _, errno := fh.(*tfsHandle).Read(ctx, make([]byte, 2), 0); errno != syscall.EIO {
		t.Fatalf("seek %v", errno)
	}
	_ = fh.(*tfsHandle).Release(ctx)
	var attr gofuse.AttrOut
	if errno := (&tfsNode{tfs: tinyFS{}, p: "missing"}).Getattr(ctx, nil, &attr); errno == 0 {
		t.Fatal("missing getattr")
	}
	broken := &FuseFS{tfs: tinyFS{}, p: "."}
	if _, errno := broken.Readdir(ctx); errno == 0 {
		t.Fatal("root readdir")
	}
	if errno := broken.Getattr(ctx, nil, &attr); errno == 0 {
		t.Fatal("root getattr")
	}
	if _, _, errno := (&tfsNode{tfs: tinyFS{}, p: "missing"}).Open(ctx, 0); errno == 0 {
		t.Fatal("missing open")
	}
	skip := &FuseFS{tfs: infoErrFS{}, p: "."}
	if stream, errno := skip.Readdir(ctx); errno != 0 || stream == nil {
		t.Fatalf("skip readdir %v", errno)
	} else {
		for stream.HasNext() {
			_, _ = stream.Next()
		}
		stream.Close()
	}
	if _, errno := (&tfsNode{tfs: infoErrFS{}, p: "."}).Readdir(ctx); errno != 0 {
		t.Fatal(errno)
	}
}

type infoErrFS struct{}

func (infoErrFS) Open(name string) (fs.File, error) {
	if name == "." {
		return infoErrDir{}, nil
	}
	return nil, fs.ErrNotExist
}

type infoErrDir struct{}

func (infoErrDir) Read([]byte) (int, error) { return 0, io.EOF }
func (infoErrDir) Close() error             { return nil }
func (infoErrDir) Stat() (fs.FileInfo, error) {
	return fakeInfo{name: ".", mode: 0o755, dir: true}, nil
}
func (infoErrDir) ReadDir(int) ([]fs.DirEntry, error) { return []fs.DirEntry{infoErrEntry{}}, nil }

type infoErrEntry struct{}

func (infoErrEntry) Name() string               { return "x" }
func (infoErrEntry) IsDir() bool                { return false }
func (infoErrEntry) Type() fs.FileMode          { return 0 }
func (infoErrEntry) Info() (fs.FileInfo, error) { return nil, errors.New("info") }

type errReadFile struct{}

func (errReadFile) Read([]byte) (int, error)       { return 0, errors.New("boom") }
func (errReadFile) Seek(int64, int) (int64, error) { return 0, nil }
func (errReadFile) Close() error                   { return nil }
func (errReadFile) Stat() (fs.FileInfo, error)     { return fakeInfo{name: "a", mode: 0o644}, nil }

func TestReadError(t *testing.T) {
	node := &tfsNode{tfs: readErrFS{}, p: "a"}
	fh, _, errno := node.Open(context.Background(), 0)
	if errno != 0 {
		t.Fatal(errno)
	}
	if _, errno := fh.(*tfsHandle).Read(context.Background(), make([]byte, 2), 0); errno != syscall.EIO {
		t.Fatalf("read %v", errno)
	}
	if errno := fh.(*tfsHandle).Release(context.Background()); errno != 0 {
		t.Fatal(errno)
	}
}

type readErrFS struct{}

func (readErrFS) Open(string) (fs.File, error) { return errReadFile{}, nil }
