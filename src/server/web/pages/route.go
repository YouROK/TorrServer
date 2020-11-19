package pages

import (
	"server/web/pages/template"

	"github.com/gin-gonic/gin"
)

var temp *template.Template

func SetupRoute(route *gin.Engine) {
	temp = template.InitTemplate(route)

	route.GET("/", mainPage)
	// route.GET("/api.js", apijsPage)
	// route.GET("/main.js", mainjsPage)
}
