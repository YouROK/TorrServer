//go:build !windows
// +build !windows

package fusefs

import (
	"hash/fnv"
	"path/filepath"
	"regexp"
	"strings"

	"server/torr"
)

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

// Generate inode from string (useful for consistent inodes for same paths)
func inodeFromString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// Get torrent inode base
func getTorrentInode(t *torr.Torrent) uint64 {
	return hashToInode(t.Hash())
}

// For torrent directories - combine torrent hash with path
func getTorrentDirIno(torrentHash, dirName string) uint64 {
	return inodeFromString(torrentHash + ":" + dirName)
}

// For torrent files - combine torrent hash with file path
func getTorrentFileIno(torrentHash, filePath string, index int) uint64 {
	// Use both string hash and index for extra uniqueness
	return inodeFromString(torrentHash+":"+filePath) + uint64(index)
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

// joinPath safely joins path components
func joinPath(base, name string) string {
	return filepath.Join(base, sanitizeName(name))
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
