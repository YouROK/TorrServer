//go:build !windows
// +build !windows

package fusefs

import (
	"context"
	"io"
	"syscall"

	"server/log"
	"server/settings"
	"server/torr"
	"server/torr/storage/torrstor"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type TorrentDir struct {
	fs.Inode
	isTorrentRoot bool
	torrent       *torr.Torrent
}

type TorrentFile struct {
	fs.Inode
	torrent *torr.Torrent
	file    *torrent.File
	reader  *torrstor.Reader
	mu      sync.Mutex
}

// Readdir lists the files in a torrent directory
func (td *TorrentDir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	if td.torrent == nil {
		return nil, syscall.ENOENT
	}

	if !td.torrent.GotInfo() {
		for i := 0; i < 10; i++ {
			tor := torr.GetTorrent(td.torrent.Hash().String())
			if tor.GotInfo() {
				td.torrent = tor
				break
			}
			time.Sleep(time.Second)
		}
	}

	files := td.torrent.Files()
	if files == nil {
		return nil, syscall.ENOENT
	}

	fullPath := getCurrentDirPath(&td.Inode)
	parts := strings.Split(fullPath, "/")

	relPrefix := ""
	if len(parts) > 2 {
		relPrefix = strings.Join(parts[2:], "/")
	}

	entriesMap := make(map[string]fuse.DirEntry)

	for _, file := range files {
		rel := file.DisplayPath()
		if relPrefix != "" {
			if !strings.HasPrefix(rel, relPrefix+"/") && rel != relPrefix {
				continue
			}
			rel = strings.TrimPrefix(rel, relPrefix+"/")
		}

		elems := strings.Split(rel, "/")
		if len(elems) == 1 {
			// файл текущего уровня
			n := sanitizeName(elems[0])
			if _, ok := entriesMap[n]; !ok {
				entriesMap[n] = fuse.DirEntry{
					Name: n,
					Ino:  inodeFromString(relPrefix + "/" + elems[0]),
					Mode: fuse.S_IFREG,
				}
			}
		} else {
			// подкаталог
			orig := elems[0]
			n := sanitizeName(orig)
			if _, ok := entriesMap[n]; !ok {
				entriesMap[n] = fuse.DirEntry{
					Name: n,
					Ino:  inodeFromString(relPrefix + "/" + orig),
					Mode: fuse.S_IFDIR,
				}
			}
		}
	}

	entries := make([]fuse.DirEntry, 0, len(entriesMap))
	for _, e := range entriesMap {
		entries = append(entries, e)
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

	fullPath := getCurrentDirPath(&td.Inode)
	parts := strings.Split(fullPath, "/")

	relPrefix := ""
	if len(parts) > 2 {
		relPrefix = strings.Join(parts[2:], "/")
	}

	baseInode := getTorrentInode(td.torrent)

	for i, file := range files {
		rel := file.DisplayPath()
		if relPrefix != "" {
			if !strings.HasPrefix(rel, relPrefix+"/") && rel != relPrefix {
				continue
			}
			rel = strings.TrimPrefix(rel, relPrefix+"/")
		}

		elems := strings.Split(rel, "/")
		if len(elems) == 1 {
			if sanitizeName(elems[0]) != name {
				continue
			}

			tf := &TorrentFile{
				torrent: td.torrent,
				file:    file,
			}

			out.Attr.Mode = fuse.S_IFREG | 0o644
			out.Attr.Size = uint64(file.Length())
			out.Attr.Ino = baseInode + uint64(i+1)

			return td.Inode.NewPersistentInode(ctx, tf, fs.StableAttr{
				Mode: fuse.S_IFREG,
				Ino:  out.Attr.Ino,
			}), 0
		}
	}

	for _, file := range files {
		rel := file.DisplayPath()
		if relPrefix != "" {
			if !strings.HasPrefix(rel, relPrefix+"/") && rel != relPrefix {
				continue
			}
			rel = strings.TrimPrefix(rel, relPrefix+"/")
		}

		elems := strings.Split(rel, "/")
		if len(elems) > 1 && sanitizeName(elems[0]) == name {
			childDir := &TorrentDir{
				torrent: td.torrent,
			}
			origDirName := elems[0]
			ino := inodeFromString(relPrefix + "/" + origDirName)

			out.Attr.Mode = fuse.S_IFDIR | 0o755
			out.Attr.Ino = ino

			return td.Inode.NewPersistentInode(ctx, childDir, fs.StableAttr{
				Mode: fuse.S_IFDIR,
				Ino:  ino,
			}), 0
		}
	}

	return nil, syscall.ENOENT
}

func (td *TorrentDir) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	if td.isTorrentRoot {
		out.Mode = fuse.S_IFDIR | 0o777
	} else {
		out.Mode = fuse.S_IFDIR | 0o755
	}
	out.Ino = td.Inode.StableAttr().Ino
	return 0
}

// Open torrent file for reading
func (tf *TorrentFile) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if settings.BTsets.EnableDebug {
		log.TLogln("TorrentFile.Open called")
		log.TLogln("File path:", tf.file.DisplayPath())
		log.TLogln("File size:", tf.file.Length())
		log.TLogln("Torrent title:", tf.torrent.Title)
	}

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
