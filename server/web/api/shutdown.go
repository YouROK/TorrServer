package api

import (
	"net/http"
	sets "server/settings"
	"server/torr"
	"time"

	"github.com/gin-gonic/gin"
)

// shutdown godoc
// @Summary		Shuts down server
// @Description	Gracefully shuts down server.
//
// @Tags			API
//
// @Success		200
// @Router			/shutdown [get]
func shutdown(c *gin.Context) {
	if sets.ReadOnly {
		c.Status(http.StatusForbidden)
		return
	}
	c.Status(200)
	go func() {
		time.Sleep(1000)
		torr.Shutdown()
	}()
}
