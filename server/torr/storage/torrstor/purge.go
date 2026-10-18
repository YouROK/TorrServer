package torrstor

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"server/log"
	"server/settings"
)

// Turning the long-term disk cache off deletes everything it saved. The purge is
// recorded in a marker in the config dir before anything is deleted and the marker
// is removed only when every torrent's cache directory is gone, so a purge cut short
// by a crash or restart is finished by ResumeDiskCachePurge on the next start.

const purgeMarkerName = "cache_purge.json"

type purgeMarker struct {
	Path string `json:"path"`
}

func purgeMarkerPath() string {
	return filepath.Join(settings.Path, purgeMarkerName)
}

// PurgeDiskCache deletes the cache directories of all torrents under path.
func PurgeDiskCache(path string) error {
	abs, err := purgeTarget(path)
	if err != nil {
		return err
	}
	data, err := json.Marshal(purgeMarker{Path: abs})
	if err != nil {
		return err
	}
	if err := os.WriteFile(purgeMarkerPath(), data, 0o666); err != nil {
		return fmt.Errorf("save cache purge marker: %w", err)
	}
	log.TLogln("Long-term disk cache disabled, remove saved cache in", abs)
	return purgeDiskCache(abs)
}

// ResumeDiskCachePurge finishes a purge that was interrupted. It must run before
// the torrent client starts writing to the cache.
func ResumeDiskCachePurge() {
	data, err := os.ReadFile(purgeMarkerPath())
	if err != nil {
		return
	}
	var marker purgeMarker
	if err := json.Unmarshal(data, &marker); err != nil || marker.Path == "" {
		log.TLogln("Drop unreadable cache purge marker:", err)
		_ = os.Remove(purgeMarkerPath())
		return
	}
	log.TLogln("Resume interrupted cache purge in", marker.Path)
	if err := purgeDiskCache(marker.Path); err != nil {
		log.TLogln("Cache purge is not finished, retry on next start:", err)
	}
}

func purgeTarget(path string) (string, error) {
	if path == "" || path == "/" {
		return "", errors.New("cache path is not set")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if abs == filepath.VolumeName(abs)+string(filepath.Separator) {
		return "", errors.New("refusing to purge a filesystem root")
	}
	return abs, nil
}

// purgeDiskCache removes the torrent cache directories (named by info hash) and
// leaves anything else in path alone. The marker goes only after all are removed.
func purgeDiskCache(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	var failed int
	var lastErr error
	removed := 0
	for _, entry := range entries {
		if !entry.IsDir() || !isInfoHashName(entry.Name()) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
			failed++
			lastErr = err
			continue
		}
		removed++
	}

	// the in-memory usage no longer matches the disk; rescan on next use
	diskUsed.Store(0)
	diskUsedSync.Store(0)

	if failed > 0 {
		return fmt.Errorf("%d cache directories not removed: %w", failed, lastErr)
	}
	if err := os.Remove(purgeMarkerPath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	log.TLogln("Saved cache removed:", removed, "torrents")
	return nil
}

func isInfoHashName(name string) bool {
	if len(name) != 40 {
		return false
	}
	_, err := hex.DecodeString(name)
	return err == nil
}
