package web

import (
	"github.com/gin-gonic/gin"
	"server/torr"
	"server/web/api"
)

var (
	BTS      = torr.NewBTS()
	waitChan = make(chan error)
)

func Start(port string) {

	BTS.Connect()

	route := gin.New()
	route.Use(gin.Logger(), gin.Recovery())

	api.SetupRouteApi(route, BTS)

	waitChan <- route.Run(":" + port)
}

func Wait() error {
	return <-waitChan
}

func Stop() {
	BTS.Disconnect()
	waitChan <- nil
}
