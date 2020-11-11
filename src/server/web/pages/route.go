package pages

import (
	"github.com/gin-gonic/gin"
	"server/web/pages/template"
)

var temp *template.Template

func SetupRoute(route *gin.Engine) {
	temp = template.InitTemplate(route)

	route.GET("/", mainPage)

}
