package pages

import (
	"github.com/gin-gonic/gin"
	"server/web/pages/template"
)

func mainPage(c *gin.Context) {
	c.Data(200, "text/html; charset=utf-8", []byte(template.MainPage))
}
