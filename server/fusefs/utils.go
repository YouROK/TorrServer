//go:build !windows
// +build !windows

package fusefs

import (
	"context"
	"hash/fnv"
	"path/filepath"
	"regexp"
	"strings"

	"server/torr"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func getCurrentDirPath(dir *fs.Inode) string {
	path := ""
	curr := dir

	for _, i := curr.Parent(); i != nil; {
		name, parent := curr.Parent()
		if name == "" {
			break
		}
		path = name + "/" + path
		curr = parent
	}

	return strings.Trim(path, "/")
}

func getDirLevel(dir *fs.Inode) int {
	level := 0
	curr := dir
	for _, i := curr.Parent(); i != nil; {
		_, parent := curr.Parent()
		if parent == nil {
			break
		}
		level++
		curr = parent
	}
	return level
}

func inodeFromString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// Hash to inode conversion
func hashToInode(hash [20]byte) uint64 {
	var inode uint64
	for i := 0; i < 8; i++ {
		inode = (inode << 8) | uint64(hash[i])
	}
	// Ensure it's not zero
	if inode == 0 {
		inode = 1
	}
	return inode
}

// Get torrent inode base
func getTorrentInode(t *torr.Torrent) uint64 {
	return hashToInode(t.Hash())
}

// sanitizeName cleans up file/directory names for filesystem compatibility
func sanitizeName(name string) string {
	// Remove or replace invalid characters
	invalidChars := regexp.MustCompile(`[<>:"|?*]`)
	name = invalidChars.ReplaceAllString(name, "_")

	// Replace forward slashes with underscores to avoid path confusion
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	// Trim whitespace and dots
	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")

	// Ensure the name is not empty
	if name == "" {
		name = "unnamed"
	}

	// Limit name length to avoid filesystem issues
	if len(name) > 255 {
		ext := filepath.Ext(name)
		name = name[:255-len(ext)] + ext
	}

	return name
}

// update the filesystem with current torrents
func updateTorrents(ffs *FuseFS, ctx context.Context) {
	torrents := torr.ListTorrent()

	type catState struct {
		inode    *fs.Inode
		torrents map[string]struct{}
	}
	categories := make(map[string]*catState)

	for name, child := range ffs.Inode.Children() {
		if _, ok := child.Operations().(*CategoryDir); ok {
			categories[name] = &catState{
				inode:    child,
				torrents: make(map[string]struct{}),
			}
		}
	}

	for _, t := range torrents {
		if t == nil {
			continue
		}

		catName := t.Category
		if strings.TrimSpace(catName) == "" {
			catName = "other"
		}
		catName = sanitizeName(catName)

		cs, ok := categories[catName]
		if !ok {
			catDir := &CategoryDir{category: catName}
			catInode := ffs.Inode.NewPersistentInode(ctx, catDir, fs.StableAttr{
				Mode: fuse.S_IFDIR,
				Ino:  inodeFromString("cat/" + catName),
			})
			ffs.Inode.AddChild(catName, catInode, false)

			cs = &catState{
				inode:    catInode,
				torrents: make(map[string]struct{}),
			}
			categories[catName] = cs
		}

		var dirName string
		if t.Title != "" {
			dirName = sanitizeName(t.Title)
		} else if t.Torrent != nil && t.Torrent.Info() != nil {
			dirName = sanitizeName(t.Torrent.Name())
		} else if len(t.Hash().HexString()) > 0 {
			dirName = t.Hash().HexString()
		} else {
			continue
		}

		catInode := cs.inode
		if child := catInode.GetChild(dirName); child == nil {
			torrentDir := &TorrentDir{
				torrent:       t,
				isTorrentRoot: true,
				Inode:         fs.Inode{},
			}

			ch := catInode.NewPersistentInode(ctx, torrentDir, fs.StableAttr{
				Mode: fuse.S_IFDIR,
				Ino:  getTorrentInode(t),
			})
			catInode.AddChild(dirName, ch, false)
		} else if dir, ok := child.Operations().(*TorrentDir); ok {
			dir.torrent = t
		}

		cs.torrents[dirName] = struct{}{}
	}

	for catName, cs := range categories {
		for name, child := range cs.inode.Children() {
			if _, ok := child.Operations().(*TorrentDir); !ok {
				continue
			}
			if _, alive := cs.torrents[name]; !alive {
				cs.inode.RmChild(name)
			}
		}

		if len(cs.inode.Children()) == 0 {
			ffs.Inode.RmChild(catName)
		}
	}
}

// GetStatus returns the current status of the FUSE filesystem
type FuseStatus struct {
	Enabled   bool   `json:"enabled"`
	MountPath string `json:"mount_path,omitempty"`
	Error     string `json:"error,omitempty"`
}

// GetStatus returns the current FUSE filesystem status
func GetStatus() FuseStatus {
	ffs := GetFuseFS()

	status := FuseStatus{
		Enabled:   ffs.IsEnabled(),
		MountPath: ffs.GetMountPath(),
	}

	return status
}
