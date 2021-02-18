package server

import (
	"server/settings"
	"server/web"
)

func Start(port string, roSets bool) {
	settings.InitSets(roSets)
	if port == "" {
		port = "8090"
	}
	web.Start(port)
}

func WaitServer() string {
	err := web.Wait()
	if err != nil {
		return err.Error()
	}
	return ""
}

func Stop() {
	web.Stop()
	settings.CloseDB()
}
