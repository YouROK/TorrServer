package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	cfg "server/settings"
)

func dlnaFS(c *gin.Context) {
	if cfg.BTsets == nil || !cfg.BTsets.EnableDLNALocal {
		c.String(http.StatusForbidden, "DLNA local is disabled")
		return
	}

	root := cfg.BTsets.DLNALocalRoot
	if root == "" {
		// Fallback to "where the server is installed"
		exe, err := os.Executable()
		if err != nil {
			c.String(http.StatusBadRequest, "DLNA local root is not configured")
			return
		}
		root = filepath.Dir(exe)
	}

	// Gin wildcard includes leading slash: "/Movies/film.mkv"
	rel := strings.TrimPrefix(c.Param("path"), "/")

	full, err := dlnaSecureJoin(root, rel)
	if err != nil {
		c.String(http.StatusForbidden, "Forbidden")
		return
	}

	st, err := os.Stat(full)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if st.IsDir() {
		c.String(http.StatusBadRequest, "Not a file")
		return
	}

	// Supports Range requests (required by many DLNA clients).
	c.File(full)
}

func dlnaSecureJoin(root, rel string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	rel = filepath.FromSlash(rel)
	rel = filepath.Clean(rel)
	if rel == "." {
		rel = ""
	}

	full := filepath.Join(rootAbs, rel)
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}

	// Ensure fullAbs is inside rootAbs
	if fullAbs != rootAbs {
		prefix := rootAbs + string(os.PathSeparator)
		if !strings.HasPrefix(fullAbs, prefix) {
			return "", os.ErrPermission
		}
	}

	return fullAbs, nil
}
