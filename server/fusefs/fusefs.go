//go:build !windows
// +build !windows

package fusefs

import (
	"context"
	"errors"
	"os"
	"sync"
	"syscall"

	"github.com/anacrolix/torrent"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"server/log"
	"server/torr"
	"server/torr/storage/torrstor"
)

// FUSE filesystem for TorrServer
type FuseFS struct {
	fs.Inode
	mountPath string
	server    *fuse.Server
	mu        sync.RWMutex
	enabled   bool
}

// directory containing torrent files
type TorrentDir struct {
	fs.Inode
	torrent *torr.Torrent
}

// file within a torrent
type TorrentFile struct {
	fs.Inode
	torrent *torr.Torrent
	file    *torrent.File
	reader  *torrstor.Reader
	mu      sync.Mutex
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
		return err
	}

	ffs.mountPath = mountPath

	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			AllowOther: true,
			Name:       "torrserver",
			FsName:     "torrserver-fuse",
		},
	}

	server, err := fs.Mount(mountPath, ffs, opts)
	if err != nil {
		return err
	}

	ffs.server = server
	ffs.enabled = true

	log.TLogln("FUSE filesystem mounted at:", mountPath)

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
	ffs.updateTorrents(ctx)
}

// update the filesystem with current torrents
func (ffs *FuseFS) updateTorrents(ctx context.Context) {
	torrents := torr.ListTorrent()

	for _, t := range torrents {
		if t != nil && t.GotInfo() {
			// Get torrent name safely
			var dirName string
			if t.Torrent != nil && t.Torrent.Info() != nil {
				dirName = sanitizeName(t.Torrent.Name())
			} else if t.Title != "" {
				dirName = sanitizeName(t.Title)
			} else {
				// Skip this torrent if we can't get a name
				continue
			}

			if child := ffs.Inode.GetChild(dirName); child == nil {
				// Create new torrent directory
				torrentDir := &TorrentDir{torrent: t}
				ffs.Inode.AddChild(dirName, ffs.Inode.NewPersistentInode(ctx, torrentDir, fs.StableAttr{Mode: fuse.S_IFDIR}), false)
			}
		}
	}
}

// contents of the root directory (all torrents)
func (ffs *FuseFS) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	ffs.updateTorrents(ctx)

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
	ffs.updateTorrents(ctx)

	child := ffs.Inode.GetChild(name)
	if child == nil {
		return nil, syscall.ENOENT
	}

	return child, 0
}
