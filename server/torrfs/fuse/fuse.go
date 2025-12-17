//go:build !windows
// +build !windows

package fuse

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"sync"
	"syscall"
	"time"

	"server/log"
	"server/settings"
	torrfs "server/torrfs"

	gofusefs "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type FuseStatus struct {
	Enabled   bool   `json:"enabled"`
	MountPath string `json:"mount_path"`
}

type FuseFS struct {
	gofusefs.Inode

	mountPath string
	server    *fuse.Server

	mu      sync.RWMutex
	enabled bool

	tfs fs.FS
	p   string // "."
}

var (
	globalFuseFS *FuseFS
	fuseMutex    sync.Mutex
)

func NewFuseFS() *FuseFS { return &FuseFS{enabled: false} }

func FuseAutoMount() {
	if settings.Args.FusePath != "" {
		ffs := GetFuseFS()
		if !ffs.enabled {
			log.TLogln("FUSE mount")
			err := ffs.Mount(settings.Args.FusePath)
			if err != nil {
				log.TLogln("Failed to auto-mount FUSE filesystem:", err)
				os.Exit(1)
			}
		}
	}
}

func FuseCleanup() {
	ffs := GetFuseFS()
	if ffs.enabled {
		_ = ffs.Unmount()
	}
}

func GetFuseFS() *FuseFS {
	fuseMutex.Lock()
	defer fuseMutex.Unlock()
	if globalFuseFS == nil {
		globalFuseFS = NewFuseFS()
	}
	return globalFuseFS
}

func (ffs *FuseFS) GetMountPath() string {
	ffs.mu.RLock()
	defer ffs.mu.RUnlock()
	return ffs.mountPath
}

func (ffs *FuseFS) Mount(mountPath string) error {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()

	if ffs.enabled {
		return errors.New("FUSE filesystem is already mounted")
	}
	if err := os.MkdirAll(mountPath, 0o755); err != nil {
		log.TLogln("Error create FUSE mount point dir:", err)
		return err
	}

	ffs.mountPath = mountPath
	ffs.tfs = torrfs.AsFS(torrfs.New())
	ffs.p = "."

	entryTimeout := time.Second
	attrTimeout := time.Second

	opts := &gofusefs.Options{
		MountOptions: fuse.MountOptions{
			AllowOther: true,
			Name:       "torrserver",
			FsName:     "torrserver-fuse",
			Debug:      settings.BTsets.EnableDebug,
		},
		EntryTimeout: &entryTimeout,
		AttrTimeout:  &attrTimeout,
		UID:          uint32(os.Getuid()),
		GID:          uint32(os.Getgid()),
	}

	srv, err := gofusefs.Mount(mountPath, ffs, opts)
	if err != nil {
		log.TLogln("Error mount FUSE filesystem:", err)
		return err
	}

	ffs.server = srv
	ffs.enabled = true
	log.TLogln("FUSE filesystem mounted at", mountPath)
	go ffs.server.Wait()
	return nil
}

func (ffs *FuseFS) Unmount() error {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()

	if !ffs.enabled {
		return errors.New("FUSE filesystem is not mounted")
	}
	if err := ffs.server.Unmount(); err != nil {
		return err
	}

	ffs.enabled = false
	ffs.server = nil
	ffs.mountPath = ""
	ffs.tfs = nil
	ffs.p = ""

	log.TLogln("FUSE filesystem unmounted")
	return nil
}

// ----- go-fuse integration -----

var (
	_ = (gofusefs.InodeEmbedder)((*FuseFS)(nil))
	_ = (gofusefs.NodeOnAdder)((*FuseFS)(nil))
	_ = (gofusefs.NodeGetattrer)((*FuseFS)(nil))
	_ = (gofusefs.NodeReaddirer)((*FuseFS)(nil))
	_ = (gofusefs.NodeLookuper)((*FuseFS)(nil))
)

func (ffs *FuseFS) EmbeddedInode() *gofusefs.Inode { return &ffs.Inode }

func (ffs *FuseFS) OnAdd(ctx context.Context) {
	if ffs.p == "" {
		ffs.p = "."
	}
}

// ----- Root ops -----

func (ffs *FuseFS) Getattr(ctx context.Context, fh gofusefs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	fi, err := fs.Stat(ffs.tfs, ".")
	if err != nil {
		return errno(err)
	}
	fillAttr(&out.Attr, fi)
	return 0
}

func (ffs *FuseFS) Readdir(ctx context.Context) (gofusefs.DirStream, syscall.Errno) {
	des, err := fs.ReadDir(ffs.tfs, ".")
	if err != nil {
		log.TLogln("FUSE root Readdir error:", err)
		return nil, errno(err)
	}

	out := make([]fuse.DirEntry, 0, len(des))
	for _, de := range des {
		fi, err := de.Info()
		if err != nil {
			continue
		}
		out = append(out, fuse.DirEntry{
			Name: de.Name(),
			Mode: fuseModeFromInfo(fi),
		})
	}
	return gofusefs.NewListDirStream(out), 0
}

func (ffs *FuseFS) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*gofusefs.Inode, syscall.Errno) {
	childPath := path.Join(".", name)

	fi, err := fs.Stat(ffs.tfs, childPath)
	if err != nil {
		return nil, errno(err)
	}

	fillAttr(&out.Attr, fi)
	out.AttrValid = 1
	out.AttrValidNsec = 0
	out.EntryValid = 1
	out.EntryValidNsec = 0

	mode := fuseModeFromInfo(fi)
	ch := ffs.NewInode(ctx, &tfsNode{tfs: ffs.tfs, p: childPath}, gofusefs.StableAttr{Mode: mode})
	return ch, 0
}

// ----- Regular nodes -----

type tfsNode struct {
	gofusefs.Inode
	tfs fs.FS
	p   string
}

var (
	_ = (gofusefs.NodeGetattrer)((*tfsNode)(nil))
	_ = (gofusefs.NodeReaddirer)((*tfsNode)(nil))
	_ = (gofusefs.NodeLookuper)((*tfsNode)(nil))
	_ = (gofusefs.NodeOpener)((*tfsNode)(nil))
)

func (n *tfsNode) full(name string) string { return path.Join(n.p, name) }

func (n *tfsNode) Getattr(ctx context.Context, fh gofusefs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	fi, err := fs.Stat(n.tfs, n.p)
	if err != nil {
		return errno(err)
	}
	fillAttr(&out.Attr, fi)
	return 0
}

func (n *tfsNode) Readdir(ctx context.Context) (gofusefs.DirStream, syscall.Errno) {
	des, err := fs.ReadDir(n.tfs, n.p)
	if err != nil {
		return nil, errno(err)
	}

	out := make([]fuse.DirEntry, 0, len(des))
	for _, de := range des {
		fi, err := de.Info()
		if err != nil {
			continue
		}
		out = append(out, fuse.DirEntry{
			Name: de.Name(),
			Mode: fuseModeFromInfo(fi),
		})
	}
	return gofusefs.NewListDirStream(out), 0
}

func (n *tfsNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*gofusefs.Inode, syscall.Errno) {
	childPath := n.full(name)

	fi, err := fs.Stat(n.tfs, childPath)
	if err != nil {
		return nil, errno(err)
	}

	fillAttr(&out.Attr, fi)
	out.AttrValid = 1
	out.AttrValidNsec = 0
	out.EntryValid = 1
	out.EntryValidNsec = 0

	mode := fuseModeFromInfo(fi)
	ch := n.NewInode(ctx, &tfsNode{tfs: n.tfs, p: childPath}, gofusefs.StableAttr{Mode: mode})
	return ch, 0
}

func (n *tfsNode) Open(ctx context.Context, flags uint32) (gofusefs.FileHandle, uint32, syscall.Errno) {
	if flags&(fuse.O_ANYWRITE) != 0 {
		return nil, 0, syscall.EROFS
	}

	f, err := n.tfs.Open(n.p)
	if err != nil {
		return nil, 0, errno(err)
	}
	if _, ok := f.(io.ReadSeeker); !ok {
		_ = f.Close()
		return nil, 0, syscall.ENOSYS
	}

	return &tfsHandle{f: f}, fuse.FOPEN_DIRECT_IO, 0
}

// ----- File handle -----

type tfsHandle struct {
	f fs.File // must implement io.ReadSeeker
}

var (
	_ = (gofusefs.FileReader)((*tfsHandle)(nil))
	_ = (gofusefs.FileReleaser)((*tfsHandle)(nil))
)

func (h *tfsHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	rs := h.f.(io.ReadSeeker)

	if _, err := rs.Seek(off, io.SeekStart); err != nil {
		return nil, syscall.EIO
	}
	n, err := rs.Read(dest)
	if err != nil && err != io.EOF {
		return nil, syscall.EIO
	}
	return fuse.ReadResultData(dest[:n]), 0
}

func (h *tfsHandle) Release(ctx context.Context) syscall.Errno {
	_ = h.f.Close()
	return 0
}

// ----- Attribute helpers -----

func fuseModeFromInfo(fi fs.FileInfo) uint32 {
	if fi.IsDir() {
		return fuse.S_IFDIR | uint32(fi.Mode().Perm())
	}
	return fuse.S_IFREG | uint32(fi.Mode().Perm())
}

func fillAttr(a *fuse.Attr, fi fs.FileInfo) {
	a.Mode = fuseModeFromInfo(fi)

	if fi.IsDir() {
		a.Size = 4096
	} else {
		a.Size = uint64(fi.Size())
	}

	mt := fi.ModTime()
	if mt.IsZero() {
		mt = time.Now()
	}
	a.Mtime = uint64(mt.Unix())
	a.Mtimensec = uint32(mt.Nanosecond())

	a.Ctime = a.Mtime
	a.Ctimensec = a.Mtimensec

	a.Atime = a.Mtime
	a.Atimensec = a.Mtimensec
}

// ----- errno mapping -----

func errno(err error) syscall.Errno {
	if err == nil {
		return 0
	}
	if pe, ok := err.(*fs.PathError); ok {
		return errno(pe.Err)
	}
	switch err {
	case fs.ErrNotExist:
		return syscall.ENOENT
	case fs.ErrPermission:
		return syscall.EPERM
	case fs.ErrInvalid:
		return syscall.EINVAL
	default:
		return syscall.EIO
	}
}
