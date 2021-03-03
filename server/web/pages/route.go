package pages

import (
	"github.com/gin-gonic/gin"
	"server/torr"
	"server/web/pages/template"
)

func SetupRoute(route *gin.RouterGroup) {
	route.GET("/", mainPage)
	route.GET("/stat", statPage)
}

func mainPage(c *gin.Context) {
	c.Data(200, "text/html; charset=utf-8", template.IndexHtml)
}

func statPage(c *gin.Context) {
	torr.WriteStatus(c.Writer)
	c.Status(200)
}
