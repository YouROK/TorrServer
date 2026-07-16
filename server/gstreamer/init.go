//go:build gst

package gstreamer

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	defaultServiceMu sync.Mutex
	defaultService   *Service
)

func SetupRoute(route gin.IRouter) {
	getDefaultService().SetupRoute(route)
}

func Stop() {
	defaultServiceMu.Lock()
	service := defaultService
	defaultService = nil
	defaultServiceMu.Unlock()

	if service != nil {
		service.Dispose()
	}
}

func Remove(id string) bool {
	defaultServiceMu.Lock()
	service := defaultService
	defaultServiceMu.Unlock()

	if service == nil {
		return false
	}
	return service.TryRemove(id)
}

func getDefaultService() *Service {
	defaultServiceMu.Lock()
	defer defaultServiceMu.Unlock()

	if defaultService == nil {
		defaultService = NewService(DefaultConfig())
	}

	return defaultService
}

func CurrentConfig() Config {
	conf := getDefaultService().currentConfig()
	checkGStreamer(conf)
	if version := effectiveGStreamerVersion(conf); version.valid() {
		conf.GSTVersion = version.configValue()
	}
	return conf
}

func UpdateConfig(conf Config) error {
	if err := SaveConfig(conf); err != nil {
		return err
	}
	getDefaultService().updateConfig(conf)
	return nil
}

func PlatformDefaults() Config {
	return defaultConfigWithoutSettings().normalized()
}

func ResetConfig() error {
	conf := PlatformDefaults()
	if err := SaveConfig(conf); err != nil {
		return err
	}
	getDefaultService().updateConfig(conf)
	return nil
}
