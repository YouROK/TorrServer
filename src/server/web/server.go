package web

import (
	"github.com/gin-gonic/gin"
	"server/torr"
)

var (
	BTS      = torr.NewBTS()
	waitChan = make(chan error)
)

func Start(port string) {

	BTS.Connect()

	route := gin.New()
	route.Use(gin.Logger(), gin.Recovery())

	route.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	waitChan <- route.Run(":" + port)
}

func Wait() error {
	return <-waitChan
}

func Stop() {
	BTS.Disconnect()
	waitChan <- nil
}
