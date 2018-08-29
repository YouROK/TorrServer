package server

import (
	"fmt"

	"server/settings"
	"server/web"
)

func Start(settingsPath, port string) {
	settings.Path = settingsPath
	err := settings.ReadSettings()
	if err != nil {
		fmt.Println("Error read settings:", err)
	}
	if port == "" {
		port = "8090"
	}
	server.Start(port)
}

func WaitServer() string {
	err := server.Wait()
	if err != nil {
		return err.Error()
	}
	return ""
}

func Stop() {
	go server.Stop()
	settings.CloseDB()
}
