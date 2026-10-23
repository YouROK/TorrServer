//go:build !windows && !ios

package fuse

import (
	"errors"
	"io/fs"
	"syscall"
	"testing"
	"time"

	gofuse "github.com/hanwen/go-fuse/v2/fuse"
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

func TestNewFuseFSDisabled(t *testing.T) {
	ffs := NewFuseFS()
	if ffs.GetMountPath() != "" {
		t.Fatal(ffs.GetMountPath())
	}
	if err := ffs.Unmount(); err == nil {
		t.Fatal("unmount of a fresh filesystem should fail")
	}
}
