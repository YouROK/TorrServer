//go:build !windows
// +build !windows

package fusefs

import (
	"path/filepath"
	"regexp"
	"strings"
)

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
