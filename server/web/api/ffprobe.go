package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/vansante/go-ffprobe.v2"
	"net/http"
	"os"
	"path/filepath"
	"server/utils"
)

func ffp(c *gin.Context) {
	hash := c.Param("hash")
	indexStr := c.Param("id")

	if hash == "" || indexStr == "" {
		c.AbortWithError(http.StatusNotFound, errors.New("link should not be empty"))
		return
	}

	host := utils.GetScheme(c) + "://" + c.Request.Host + "/stream?link=" + hash + "&index=" + indexStr + "&play"
	fmt.Println(host)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	if _, err := os.Stat("ffprobe"); os.IsNotExist(err) {
		ffprobe.SetFFProbeBinPath(filepath.Dir(os.Args[0]) + "/ffprobe")
	}

	data, err := ffprobe.ProbeURL(ctx, host)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("error getting data: %v", err))
		return
	}

	c.JSON(200, data)
}
