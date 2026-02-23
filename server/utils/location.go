package utils

import (
	"github.com/gin-contrib/location/v2"
	"github.com/gin-gonic/gin"
)

func GetScheme(c *gin.Context) string {
	url := location.Get(c)
	if url == nil {
		return "http"
	}
	return url.Scheme
}

func GetHost(c *gin.Context) string {
	url := location.Get(c)
	if url == nil {
		return c.Request.Host
	}
	return url.Host
}
