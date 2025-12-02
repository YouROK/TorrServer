//go:build !windows
// +build !windows

package api

import (
	"github.com/gin-gonic/gin"
)

func setupFuseRoutes(authorized gin.IRouter) {
	// FUSE filesystem routes
	authorized.GET("/fuse/status", fuseStatus)
	authorized.POST("/fuse/mount", fuseMount)
	authorized.POST("/fuse/unmount", fuseUnmount)
}
