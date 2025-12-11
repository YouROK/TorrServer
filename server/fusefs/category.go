//go:build !windows
// +build !windows

package fusefs

import (
	"context"
	"server/torr"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type CategoryDir struct {
	fs.Inode
	category string
}

func (cd *CategoryDir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entries := []fuse.DirEntry{}
	for name, child := range cd.Inode.Children() {
		entries = append(entries, fuse.DirEntry{
			Name: name,
			Ino:  child.StableAttr().Ino,
			Mode: child.StableAttr().Mode,
		})
	}
	return fs.NewListDirStream(entries), 0
}

func (cd *CategoryDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	child := cd.Inode.GetChild(name)
	if child == nil {
		return nil, syscall.ENOENT
	}
	return child, 0
}

func (cd *CategoryDir) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = fuse.S_IFDIR | 0o777
	out.Ino = cd.Inode.StableAttr().Ino
	return 0
}

func (cd *CategoryDir) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	uploadNode := &UploadNode{name: name}
	ch := cd.Inode.NewPersistentInode(ctx, uploadNode, fs.StableAttr{
		Mode: fuse.S_IFREG | 0o666,
	})

	handle := &UploadHandle{node: uploadNode, category: cd.category}
	return ch, handle, 0, 0
}

func (cd *CategoryDir) Rmdir(ctx context.Context, name string) syscall.Errno {
	child := cd.Inode.GetChild(name)
	if child == nil {
		return syscall.ENOENT
	}

	td, ok := child.Operations().(*TorrentDir)
	if !ok || td.torrent == nil {
		return syscall.EPERM
	}

	hash := td.torrent.Hash().HexString()
	torr.DropTorrent(hash)
	torr.RemTorrent(hash)

	return 0
}
