package api

import (
	"errors"
	"fmt"
	"net/http"
	"server/ffprobe"
	sets "server/settings"

	"github.com/gin-gonic/gin"
)

func ffp(c *gin.Context) {
	hash := c.Param("hash")
	indexStr := c.Param("id")

	if hash == "" || indexStr == "" {
		c.AbortWithError(http.StatusNotFound, errors.New("link should not be empty"))
		return
	}

	link := "http://127.0.0.1:" + sets.Port + "/play/" + hash + "/" + indexStr

	data, err := ffprobe.ProbeUrl(link)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("error getting data: %v", err))
		return
	}

	c.JSON(200, data)
}
