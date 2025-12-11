//go:build !windows
// +build !windows

package fusefs

import (
	"bytes"
	"context"
	"syscall"
	"time"

	"server/log"
	"server/torr"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type UploadNode struct {
	fs.Inode
	name string
}
type UploadHandle struct {
	node     *UploadNode
	category string
	buf      bytes.Buffer
}

func (n *UploadNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = fuse.S_IFREG | 0o666
	out.Ino = n.Inode.StableAttr().Ino
	return 0
}

func (h *UploadHandle) Write(ctx context.Context, data []byte, off int64) (uint32, syscall.Errno) {
	if off != int64(h.buf.Len()) {
		return 0, syscall.EINVAL
	}
	if _, err := h.buf.Write(data); err != nil {
		return 0, syscall.EIO
	}
	return uint32(len(data)), 0
}

// Release вызывается, когда файл закрывают, для добавления торрента и удаления самого файла
func (h *UploadHandle) Release(ctx context.Context) syscall.Errno {
	minfo, err := metainfo.Load(&h.buf)
	if err != nil {
		log.TLogln("Error read torrent file:", err)
		return 0
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		log.TLogln("Error parse torrent file:", err)
		return 0
	}

	// mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	mag := minfo.Magnet(nil, &info)
	tspec := &torrent.TorrentSpec{
		InfoBytes:   minfo.InfoBytes,
		Trackers:    [][]string{mag.Trackers},
		DisplayName: info.Name,
		InfoHash:    minfo.HashInfoBytes(),
	}

	tor, err := torr.AddTorrent(tspec, "", "", "", h.category)
	if err != nil {
		log.TLogln("Error add torrent from fuse fs:", err)
		return 0
	}
	torr.SaveTorrentToDB(tor)

	// удаление через секунду чтоб проводники файлов не ругались, так как они после записи проверяют сам файл
	go func() {
		time.Sleep(time.Second)

		_, parent := h.node.Parent()
		if parent == nil {
			return
		}

		child := parent.GetChild(h.node.name)
		if child == nil {
			return
		}

		if _, ok := child.Operations().(*UploadNode); ok {
			parent.RmChild(h.node.name)
		}
	}()

	return 0
}
