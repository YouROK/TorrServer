//go:build windows
// +build windows

package api

import (
	"github.com/gin-gonic/gin"
)

func setupFuseRoutes(authorized gin.IRouter) {
	// Empty implementation for Windows - no FUSE routes
}

func FuseAutoMount() {
	// Empty implementation for Windows
}

func FuseCleanup() {
	// Empty implementation for Windows
}
