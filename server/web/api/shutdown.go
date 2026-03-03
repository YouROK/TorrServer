package api

import (
	"net/http"
	"strings"
	"time"

	sets "server/settings"
	"server/torr"

	"github.com/gin-gonic/gin"
)

// shutdown godoc
// @Summary		Shuts down server
// @Description	Gracefully shuts down server after 1 second.
//
// @Tags			API
//
// @Success		200
// @Router			/shutdown [get]
func shutdown(c *gin.Context) {
	reasonStr := strings.ReplaceAll(c.Param("reason"), `/`, "")
	if sets.ReadOnly && reasonStr == "" {
		c.Status(http.StatusForbidden)
		return
	}
	c.Status(200)
	go func() {
		time.Sleep(time.Second)
		torr.Shutdown()
	}()
}

// restart godoc
// @Summary		Restarts server
// @Description	Gracefully restarts server after 1 second. Exits with non-zero code to trigger service manager restart (systemd, launchd, etc.).
//
// @Tags			API
//
// @Success		200
// @Router			/restart [get]
func restart(c *gin.Context) {
	reasonStr := strings.ReplaceAll(c.Param("reason"), `/`, "")
	if sets.ReadOnly && reasonStr == "" {
		c.Status(http.StatusForbidden)
		return
	}
	c.Status(200)
	go func() {
		time.Sleep(time.Second)
		torr.Restart()
	}()
}
