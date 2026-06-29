//go:build !gst

package bridge

import "github.com/gin-gonic/gin"

func SetupRoute(_ gin.IRouter) {
}

func Stop() {
}

func Remove(_ string) bool {
	return false
}
