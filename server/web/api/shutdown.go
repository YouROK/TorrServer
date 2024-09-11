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
		time.Sleep(1000)
		torr.Shutdown()
	}()
}
