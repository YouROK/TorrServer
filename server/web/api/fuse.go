//go:build !windows
// +build !windows

package api

import (
	"net/http"
	"path/filepath"

	"server/fusefs"
	"server/log"
	config "server/settings"

	"github.com/gin-gonic/gin"
)

// FuseMountRequest represents a request to mount FUSE filesystem
type FuseMountRequest struct {
	MountPath string `json:"mount_path" binding:"required"`
}

// FuseUnmountRequest represents a request to unmount FUSE filesystem
type FuseUnmountRequest struct{}

// fuseStatus godoc
//
//	@Summary		Get FUSE filesystem status
//	@Description	Returns the current status of the FUSE filesystem mount
//	@Tags			FUSE
//	@Produce		json
//	@Success		200	{object}	fusefs.FuseStatus
//	@Router			/fuse/status [get]
func fuseStatus(c *gin.Context) {
	status := fusefs.GetStatus()
	c.JSON(http.StatusOK, status)
}

// fuseMount godoc
//
//	@Summary		Mount FUSE filesystem
//	@Description	Mounts the FUSE filesystem at the specified path
//	@Tags			FUSE
//	@Accept			json
//	@Produce		json
//	@Param			request	body		FuseMountRequest	true	"Mount request"
//	@Success		200		{object}	fusefs.FuseStatus
//	@Router			/fuse/mount [post]
func fuseMount(c *gin.Context) {
	var req FuseMountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and clean the mount path
	mountPath := filepath.Clean(req.MountPath)
	if mountPath == "" || mountPath == "." {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mount path"})
		return
	}

	ffs := fusefs.GetFuseFS()

	// Check if already mounted
	if ffs.IsEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FUSE filesystem is already mounted"})
		return
	}

	// Attempt to mount
	err := ffs.Mount(mountPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update settings to remember the mount path
	if config.BTsets != nil {
		config.BTsets.EnableFUSE = true
		config.BTsets.FUSEPath = mountPath
		config.SetBTSets(config.BTsets)
	}

	status := fusefs.GetStatus()
	c.JSON(http.StatusOK, status)
}

// fuseUnmount godoc
//
//	@Summary		Unmount FUSE filesystem
//	@Description	Unmounts the FUSE filesystem
//	@Tags			FUSE
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	fusefs.FuseStatus
//	@Router			/fuse/unmount [post]
func fuseUnmount(c *gin.Context) {
	ffs := fusefs.GetFuseFS()

	// Check if mounted
	if !ffs.IsEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FUSE filesystem is not mounted"})
		return
	}

	// Attempt to unmount
	err := ffs.Unmount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update settings
	if config.BTsets != nil {
		config.BTsets.EnableFUSE = false
		config.SetBTSets(config.BTsets)
	}

	status := fusefs.GetStatus()
	c.JSON(http.StatusOK, status)
}

// FuseAutoMount attempts to auto-mount FUSE if enabled in settings
func FuseAutoMount() {
	if config.BTsets != nil && config.BTsets.EnableFUSE && config.BTsets.FUSEPath != "" {
		ffs := fusefs.GetFuseFS()
		if !ffs.IsEnabled() {
			err := ffs.Mount(config.BTsets.FUSEPath)
			if err != nil {
				// Log error but don't fail startup
				log.TLogln("Failed to auto-mount FUSE filesystem:", err)
			}
		}
	}
}

// FuseCleanup unmounts FUSE filesystem during shutdown
func FuseCleanup() {
	ffs := fusefs.GetFuseFS()
	if ffs.IsEnabled() {
		_ = ffs.Unmount()
	}
}
