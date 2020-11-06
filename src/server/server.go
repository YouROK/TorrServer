package server

import (
	"server/settings"
)

func Start(settingsPath, port string, roSets bool) {
	settings.InitSets(settingsPath, roSets)
	if port == "" {
		port = "8090"
	}
	//server.Start(port)
}

func WaitServer() string {
	//err := server.Wait()
	// if err != nil {
	// 	return err.Error()
	// }
	return ""
}

func Stop() {
	// go server.Stop()
	settings.CloseDB()
}
