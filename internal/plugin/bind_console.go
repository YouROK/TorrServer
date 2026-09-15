package plugin

import (
	"silo/internal/log"

	"github.com/dop251/goja"
)

func (rt *JSRuntime) bindConsole() {
	console := rt.vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		log.Infof("[Plugin:%s] %s", rt.pluginID, call.Argument(0).String())
		return goja.Undefined()
	})
	console.Set("error", func(call goja.FunctionCall) goja.Value {
		log.Errorf("[Plugin:%s] %s", rt.pluginID, call.Argument(0).String())
		return goja.Undefined()
	})
	rt.vm.Set("console", console)
}
