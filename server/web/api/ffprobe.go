package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"server/utils"

	"github.com/gin-gonic/gin"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func ffp(c *gin.Context) {
	hash := c.Param("hash")
	indexStr := c.Param("id")

	if hash == "" || indexStr == "" {
		c.AbortWithError(http.StatusNotFound, errors.New("link should not be empty"))
		return
	}

	host := utils.GetScheme(c) + "://" + c.Request.Host + "/stream?link=" + hash + "&index=" + indexStr + "&play"
	// log.Println("ffprobe", host)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// path lookup
	path, err := exec.LookPath("ffprobe")
	if err == nil {
		// log.Println("ffprobe found in", path)
		ffprobe.SetFFProbeBinPath(path)
	} else {
		// log.Println("ffprobe not found in $PATH")
		// working dir
		if _, err := os.Stat("ffprobe"); os.IsNotExist(err) {
			ffprobe.SetFFProbeBinPath(filepath.Dir(os.Args[0]) + "/ffprobe")
		}
	}

	data, err := ffprobe.ProbeURL(ctx, host)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("error getting data: %v", err))
		return
	}

	c.JSON(200, data)
}
