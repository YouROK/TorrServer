//go:build !windows
// +build !windows

package fusefs

import (
	"context"
	"io"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Readdir lists the files in a torrent directory
func (td *TorrentDir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	if td.torrent == nil || !td.torrent.GotInfo() {
		return nil, syscall.ENOENT
	}

	files := td.torrent.Files()
	if files == nil {
		return nil, syscall.ENOENT
	}

	entries := []fuse.DirEntry{}

	for i, file := range files {
		fileName := sanitizeName(file.DisplayPath())
		entries = append(entries, fuse.DirEntry{
			Name: fileName,
			Ino:  uint64(i + 1),
			Mode: fuse.S_IFREG,
		})
	}

	return fs.NewListDirStream(entries), 0
}

// Lookup finds a file within the torrent directory
func (td *TorrentDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if td.torrent == nil || !td.torrent.GotInfo() {
		return nil, syscall.ENOENT
	}

	files := td.torrent.Files()
	if files == nil {
		return nil, syscall.ENOENT
	}
	for i, file := range files {
		if sanitizeName(file.DisplayPath()) == name {
			torrentFile := &TorrentFile{
				torrent: td.torrent,
				file:    file,
			}

			out.Attr.Mode = fuse.S_IFREG | 0o644
			out.Attr.Size = uint64(file.Length())
			out.Attr.Ino = uint64(i + 1)

			return td.Inode.NewPersistentInode(ctx, torrentFile, fs.StableAttr{
				Mode: fuse.S_IFREG,
				Ino:  out.Attr.Ino,
			}), 0
		}
	}

	return nil, syscall.ENOENT
}

func (td *TorrentDir) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = fuse.S_IFDIR | 0o755
	out.Ino = td.Inode.StableAttr().Ino
	return 0
}

// Open torrent file for reading
func (tf *TorrentFile) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if tf.reader == nil {
		reader := tf.torrent.NewReader(tf.file)
		if reader == nil {
			return nil, 0, syscall.EIO
		}
		tf.reader = reader
	}

	return &FuseFileHandle{
		torrentFile: tf,
	}, fuse.FOPEN_DIRECT_IO, 0
}

func (tf *TorrentFile) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = fuse.S_IFREG | 0o644
	out.Size = uint64(tf.file.Length())
	out.Ino = tf.Inode.StableAttr().Ino
	return 0
}

type FuseFileHandle struct {
	torrentFile *TorrentFile
}

// Read data from the torrent file
func (fh *FuseFileHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	fh.torrentFile.mu.Lock()
	defer fh.torrentFile.mu.Unlock()

	if fh.torrentFile.reader == nil {
		return nil, syscall.EIO
	}

	// Seek to the requested offset
	_, err := fh.torrentFile.reader.Seek(off, io.SeekStart)
	if err != nil {
		return nil, syscall.EIO
	}

	// Read the data
	n, err := fh.torrentFile.reader.Read(dest)
	if err != nil && err != io.EOF {
		return nil, syscall.EIO
	}

	return fuse.ReadResultData(dest[:n]), 0
}

// close the file handle
func (fh *FuseFileHandle) Release(ctx context.Context) syscall.Errno {
	fh.torrentFile.mu.Lock()
	defer fh.torrentFile.mu.Unlock()

	if fh.torrentFile.reader != nil {
		fh.torrentFile.torrent.CloseReader(fh.torrentFile.reader)
		fh.torrentFile.reader = nil
	}

	return 0
}

// Flush  file data (no-op for read-only FUSE)
func (fh *FuseFileHandle) Flush(ctx context.Context) syscall.Errno {
	return 0
}

// Fsync syncs file data (no-op for read-only FUSE)
func (fh *FuseFileHandle) Fsync(ctx context.Context, flags uint32) syscall.Errno {
	return 0
}
