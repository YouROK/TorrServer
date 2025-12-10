//go:build !windows
// +build !windows

package fusefs

import (
	"hash/fnv"
	"path/filepath"
	"regexp"
	"server/torr"
	"strings"

	"github.com/hanwen/go-fuse/v2/fs"
)

// Получить текущий путь директории через итерацию родителей
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

// Для TorrentDir получить уровень вложенности
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
