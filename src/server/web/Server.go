package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"server/settings"
	"server/torr"
	"server/version"
	"server/web/mods"
	"server/web/templates"

	"github.com/anacrolix/sync"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	server *echo.Echo
	bts    *torr.BTServer

	mutex   sync.Mutex
	fnMutex sync.Mutex
	err     error
)

func Start(port string) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("Start web server, version:", version.Version)

	bts = torr.NewBTS()
	err := bts.Connect()
	if err != nil {
		fmt.Println("Error start torrent client:", err)
		return
	}

	mutex.Lock()
	server = echo.New()
	server.HideBanner = true
	server.HidePort = true
	server.HTTPErrorHandler = HTTPErrorHandler

	//server.Use(middleware.Logger())
	server.Use(middleware.Recover())

	templates.InitTemplate(server)
	initTorrent(server)
	initSettings(server)
	initInfo(server)
	initAbout(server)
	mods.InitMods(server)

	server.GET("/", mainPage)
	server.GET("/echo", echoPage)
	server.POST("/shutdown", shutdownPage)
	server.GET("/js/api.js", templates.Api_JS)

	go func() {
		defer mutex.Unlock()

		server.Listener, err = net.Listen("tcp", "0.0.0.0:"+port)
		if err == nil {
			err = server.Start("0.0.0.0:" + port)
		}
		server = nil
		if err != nil {
			fmt.Println("Error start web server:", err)
		}
	}()
}

func Stop() {
	fnMutex.Lock()
	defer fnMutex.Unlock()
	if server != nil {
		fmt.Println("Stop web server")
		server.Close()
		server = nil
		if bts != nil {
			bts.Disconnect()
			bts = nil
		}
	}
}

func Wait() error {
	mutex.Lock()
	mutex.Unlock()
	return err
}

func mainPage(c echo.Context) error {
	return c.Render(http.StatusOK, "mainPage", nil)
}

func echoPage(c echo.Context) error {
	return c.String(http.StatusOK, version.Version)
}

func shutdownPage(c echo.Context) error {
	go func() {
		Stop()
		settings.CloseDB()
		time.Sleep(time.Second * 2)
		os.Exit(5)
	}()
	return c.NoContent(http.StatusOK)
}

func HTTPErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
		if he.Internal != nil {
			msg = fmt.Sprintf("%v, %v", err, he.Internal)
		}
	} else {
		msg = http.StatusText(code)
	}
	if _, ok := msg.(string); ok {
		msg = echo.Map{"message": msg}
	}

	if code != 404 && c.Request().URL.Path != "/torrent/stat" {
		log.Println("Web server error:", err, c.Request().URL)
	}

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD { // Issue #608
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, msg)
		}
		if err != nil {
			c.Logger().Error(err)
		}
	}
}
