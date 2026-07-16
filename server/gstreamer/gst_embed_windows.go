//go:build gst && windows && amd64 && embed_gstlib

package gstreamer

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"server/settings"
)

//go:embed embedded_gstlib_windows_amd64.zip
var embeddedGSTLibZip []byte

var (
	embeddedGSTOnce  sync.Once
	embeddedGSTRoot  string
	embeddedGSTError string
)

func embeddedGSTRuntimeRoot() string {
	embeddedGSTOnce.Do(func() {
		embeddedGSTRoot = extractEmbeddedGSTRuntime()
	})
	return embeddedGSTRoot
}

func extractEmbeddedGSTRuntime() string {
	if len(embeddedGSTLibZip) == 0 {
		embeddedGSTError = "embedded GStreamer archive is empty"
		return ""
	}

	sum := sha256.Sum256(embeddedGSTLibZip)
	fullHash := hex.EncodeToString(sum[:])
	shortHash := fullHash[:16]

	var failures []string
	for _, cacheRoot := range embeddedGSTCacheRoots() {
		root := filepath.Join(cacheRoot, "gst-lib-"+shortHash)
		if embeddedGSTRuntimeReady(root, fullHash) {
			embeddedGSTError = ""
			return root
		}
		if err := extractEmbeddedGSTRuntimeAt(root, fullHash); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", root, err))
			continue
		}
		embeddedGSTError = ""
		return root
	}
	embeddedGSTError = strings.Join(failures, "; ")
	return ""
}

func embeddedGSTCacheRoots() []string {
	var roots []string
	if cacheRoot, err := os.UserCacheDir(); err == nil && cacheRoot != "" {
		roots = appendUniquePath(roots, filepath.Join(cacheRoot, "TorrServer"))
	}
	if settings.Path != "" {
		if workingRoot, err := filepath.Abs(settings.Path); err == nil {
			roots = appendUniquePath(roots, filepath.Join(workingRoot, ".cache", "TorrServer"))
		}
	}
	if tempRoot := os.TempDir(); tempRoot != "" {
		roots = appendUniquePath(roots, filepath.Join(tempRoot, "TorrServer"))
	}
	return roots
}

func extractEmbeddedGSTRuntimeAt(root string, fullHash string) error {
	if err := os.RemoveAll(root); err != nil {
		return fmt.Errorf("remove stale runtime: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	if err := unzipEmbeddedGSTRuntime(root); err != nil {
		_ = os.RemoveAll(root)
		return fmt.Errorf("extract runtime: %w", err)
	}
	if err := os.WriteFile(embeddedGSTMarkerPath(root), []byte(fullHash), 0o644); err != nil {
		_ = os.RemoveAll(root)
		return fmt.Errorf("write marker: %w", err)
	}
	if !embeddedGSTRuntimeReady(root, fullHash) {
		_ = os.RemoveAll(root)
		return errors.New("runtime failed integrity validation")
	}
	return nil
}

func embeddedGSTRuntimeStatus() componentStatus {
	status := componentStatus{Found: len(embeddedGSTLibZip) > 0}
	root := embeddedGSTRuntimeRoot()
	status.Available = root != ""
	status.Works = status.Available
	status.Error = embeddedGSTError
	return status
}

func embeddedGSTRuntimeReady(root string, hash string) bool {
	marker, err := os.ReadFile(embeddedGSTMarkerPath(root))
	if err != nil || strings.TrimSpace(string(marker)) != hash {
		return false
	}

	required := []string{
		filepath.Join(root, "bin", "libgstreamer-1.0-0.dll"),
		filepath.Join(root, "bin", "libgstapp-1.0-0.dll"),
		filepath.Join(root, "bin", "gst-discoverer-1.0.exe"),
		filepath.Join(root, "lib", "gstreamer-1.0"),
		filepath.Join(root, "libexec", "gstreamer-1.0", "gst-plugin-scanner.exe"),
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

func embeddedGSTMarkerPath(root string) string {
	return filepath.Join(root, ".torrserver-gstlib.sha256")
}

func unzipEmbeddedGSTRuntime(root string) error {
	reader, err := zip.NewReader(bytes.NewReader(embeddedGSTLibZip), int64(len(embeddedGSTLibZip)))
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		target, ok := embeddedGSTZipTarget(root, file.Name)
		if !ok {
			continue
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}
		err = writeEmbeddedGSTFile(target, src, file.Mode())
		closeErr := src.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func embeddedGSTZipTarget(root string, name string) (string, bool) {
	cleanName := filepath.Clean(strings.ReplaceAll(name, "\\", "/"))
	if cleanName == "." || filepath.IsAbs(cleanName) || strings.HasPrefix(cleanName, "..") {
		return "", false
	}

	target := filepath.Join(root, cleanName)
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", false
	}
	return target, true
}

func writeEmbeddedGSTFile(path string, src io.Reader, mode os.FileMode) error {
	dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
