package pages

import (
	"github.com/gin-gonic/gin"
)

func mainPage(c *gin.Context) {
	c.HTML(200, "mainPage", nil)
}

func apijsPage(c *gin.Context) {
	c.HTML(200, "apijsPage", nil)
}

func mainjsPage(c *gin.Context) {
	c.HTML(200, "mainjsPage", nil)
}
