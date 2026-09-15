package plugin

import (
	"runtime"

	"silo/internal/version"

	"github.com/dop251/goja"
)

func (rt *JSRuntime) createInfoModule() *goja.Object {
	info := rt.vm.NewObject()
	info.Set("version", version.Version)
	info.Set("torrent_version", version.GetTorrentVersion())
	info.Set("os", runtime.GOOS)
	info.Set("arch", runtime.GOARCH)
	info.Set("num_cpu", runtime.NumCPU())
	info.Set("plugin_id", rt.pluginID)
	return info
}
