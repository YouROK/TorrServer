package api

import (
	"errors"
	"fmt"
	"net/http"
	"server/ffprobe"

	"server/utils"

	"github.com/gin-gonic/gin"
)

func ffp(c *gin.Context) {
	hash := c.Param("hash")
	indexStr := c.Param("id")

	if hash == "" || indexStr == "" {
		c.AbortWithError(http.StatusNotFound, errors.New("link should not be empty"))
		return
	}

	host := utils.GetScheme(c) + "://" + c.Request.Host + "/stream?link=" + hash + "&index=" + indexStr + "&play"
	// log.Println("ffprobe", host)

	data, err := ffprobe.ProbeUrl(host)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("error getting data: %v", err))
		return
	}

	c.JSON(200, data)
}
