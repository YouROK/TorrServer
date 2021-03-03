package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"server/log"
	"server/torr"
	"server/version"
	"server/web/api"
	"server/web/auth"
	"server/web/pages"
)

var (
	BTS      = torr.NewBTS()
	waitChan = make(chan error)
)

func Start(port string) {
	log.TLogln("Start TorrServer", version.Version)
	err := BTS.Connect()
	if err != nil {
		waitChan <- err
		return
	}
	gin.SetMode(gin.ReleaseMode)

	route := gin.New()
	route.Use(gin.Recovery(), cors.Default())

	route.GET("/echo", echo)

	routeAuth := auth.SetupAuth(route)
	if routeAuth != nil {
		api.SetupRoute(routeAuth)
		pages.SetupRoute(routeAuth)
	} else {
		api.SetupRoute(&route.RouterGroup)
		pages.SetupRoute(&route.RouterGroup)
	}
	log.TLogln("Start web", port)
	waitChan <- route.Run(":" + port)
}

func Wait() error {
	return <-waitChan
}

func Stop() {
	BTS.Disconnect()
	waitChan <- nil
}

func echo(c *gin.Context) {
	c.String(200, "%v", version.Version)
}
