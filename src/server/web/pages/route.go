package pages

import (
	"github.com/gin-gonic/gin"
)

func SetupRoute(route *gin.Engine) {
	route.GET("/", mainPage)
}
