package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/YouROK/tunsgo/opts"
	"github.com/YouROK/tunsgo/p2p"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func main() {
	flag.Parse()

	opts := opts.DefOptions()
	buf, err := os.ReadFile("tuns.conf")
	if err != nil {
		buf, _ = yaml.Marshal(&opts)
		os.WriteFile("tuns.conf", buf, 0644)
	} else {
		err = yaml.Unmarshal(buf, &opts)
		if err != nil {
			log.Fatal(err)
		}
	}

	server, err := p2p.NewP2PServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	gin.SetMode(gin.ReleaseMode)
	route := gin.New()

	route.Use(gin.Recovery())

	route.Any("/proxy/*url", server.GinHandler)
	route.GET("/status", func(c *gin.Context) {
		st := server.Status()
		c.JSON(http.StatusOK, st)
	})

	httpSrv := &http.Server{
		Addr:    ":" + opts.Server.Port,
		Handler: route,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %s\n", err)
		}
	}()
	fmt.Println("HTTP сервер запущен на :" + opts.Server.Port)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigc
	server.Stop()
}
