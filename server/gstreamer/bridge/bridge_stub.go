//go:build !gst

package bridge

import "github.com/gin-gonic/gin"

func SetupRoute(_ gin.IRouter) {
}

func BuiltIn() bool {
	return false
}

func Stop() {
}

func Remove(_ string) bool {
	return false
}
