//go:build !windows
// +build !windows

package fusefs

import (
	"context"
	"errors"
	"os"
	"server/settings"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"server/log"
)

// FUSE filesystem for TorrServer
type FuseFS struct {
	fs.Inode
	mountPath string
	server    *fuse.Server
	mu        sync.RWMutex
	enabled   bool
}

var (
	globalFuseFS *FuseFS
	fuseMutex    sync.Mutex
)

// create a new FUSE filesystem instance
func NewFuseFS() *FuseFS {
	return &FuseFS{
		enabled: false,
	}
}

// Returns the global FUSE filesystem instance
func GetFuseFS() *FuseFS {
	fuseMutex.Lock()
	defer fuseMutex.Unlock()

	if globalFuseFS == nil {
		globalFuseFS = NewFuseFS()
	}
	return globalFuseFS
}

func (ffs *FuseFS) Mount(mountPath string) error {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()

	if ffs.enabled {
		return errors.New("FUSE filesystem is already mounted")
	}

	// Ensure mount directory exists
	err := os.MkdirAll(mountPath, 0o755)
	if err != nil {
		log.TLogln("Error create FUSE mount point dir:", err)
		return err
	}

	ffs.mountPath = mountPath

	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			AllowOther: true,
			Name:       "torrserver",
			FsName:     "torrserver-fuse",
			Debug:      settings.BTsets.EnableDebug,
		},
		UID: uint32(os.Getuid()),
		GID: uint32(os.Getgid()),
	}

	server, err := fs.Mount(mountPath, ffs, opts)
	if err != nil {
		log.TLogln("Error mount FUSE filesystem:", err)
		os.Exit(1)
	}

	ffs.server = server
	ffs.enabled = true

	log.TLogln("FUSE filesystem mounted at", mountPath)

	// Start serving in background
	go ffs.server.Wait()

	return nil
}

func (ffs *FuseFS) Unmount() error {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()

	if !ffs.enabled {
		return errors.New("FUSE filesystem is not mounted")
	}

	err := ffs.server.Unmount()
	if err != nil {
		return err
	}

	ffs.enabled = false
	ffs.server = nil
	ffs.mountPath = ""

	log.TLogln("FUSE filesystem unmounted")
	return nil
}

// whether the FUSE filesystem is currently mounted
func (ffs *FuseFS) IsEnabled() bool {
	ffs.mu.RLock()
	defer ffs.mu.RUnlock()
	return ffs.enabled
}

// current mount path
func (ffs *FuseFS) GetMountPath() string {
	ffs.mu.RLock()
	defer ffs.mu.RUnlock()
	return ffs.mountPath
}

// OnAdd is called when the node is added to the in-memory tree
func (ffs *FuseFS) OnAdd(ctx context.Context) {
	// Initialize root directory with current torrents
	updateTorrents(ffs, ctx)
}

// contents of the root directory (all torrents)
func (ffs *FuseFS) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	updateTorrents(ffs, ctx)

	entries := []fuse.DirEntry{}
	for name, child := range ffs.Inode.Children() {
		entries = append(entries, fuse.DirEntry{
			Name: name,
			Ino:  child.StableAttr().Ino,
			Mode: child.StableAttr().Mode,
		})
	}

	return fs.NewListDirStream(entries), 0
}

// find a child node by name
func (ffs *FuseFS) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	updateTorrents(ffs, ctx)

	child := ffs.Inode.GetChild(name)
	if child == nil {
		return nil, syscall.ENOENT
	}

	return child, 0
}

func (ffs *FuseFS) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = fuse.S_IFDIR | 0o755
	out.Ino = ffs.Inode.StableAttr().Ino
	return 0
}

func (ffs *FuseFS) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	uploadNode := &UploadNode{name: name}
	ch := ffs.Inode.NewPersistentInode(ctx, uploadNode, fs.StableAttr{
		Mode: fuse.S_IFREG | 0o666,
	})

	handle := &UploadHandle{node: uploadNode, category: "other"}
	return ch, handle, 0, 0
}
