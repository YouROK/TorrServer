package plugin

import (
	"fmt"
	"strings"

	"silo/internal/bus"

	"github.com/dop251/goja"
)

func (rt *JSRuntime) createBusModule(manifest *Manifest) *goja.Object {
	pBus := bus.Get(rt.pluginID)
	busObj := rt.vm.NewObject()

	ownPrefix := "plugin:" + rt.pluginID + ":"

	checkEventPerm := func(topic string, isEmit bool) bool {
		if strings.HasPrefix(topic, ownPrefix) {
			return true
		}

		if isEmit {
			if !strings.HasPrefix(topic, "plugin:") {
				return false
			}
			for _, pattern := range manifest.Events {
				if matchTopicPattern(pattern, topic) {
					return true
				}
			}
			return false
		}

		for _, pattern := range manifest.Events {
			if matchTopicPattern(pattern, topic) {
				return true
			}
		}
		return false
	}

	busObj.Set("emit", func(call goja.FunctionCall) goja.Value {
		rawTopic := call.Argument(0).String()
		payload := call.Argument(1).Export()

		topic := rawTopic
		if !strings.Contains(topic, ":") {
			topic = ownPrefix + topic
		}

		if !checkEventPerm(topic, true) {
			panic(rt.vm.ToValue(fmt.Sprintf("permission denied: cannot emit to topic '%s'", topic)))
		}

		pBus.Emit(topic, payload)
		return goja.Undefined()
	})

	busObj.Set("on", func(call goja.FunctionCall) goja.Value {
		rawTopic := call.Argument(0).String()
		fn, ok := goja.AssertFunction(call.Argument(1))
		if !ok {
			panic(rt.vm.ToValue("second argument must be a function"))
		}

		topic := rawTopic
		if !strings.Contains(topic, ":") {
			topic = ownPrefix + topic
		}

		if !checkEventPerm(topic, false) {
			panic(rt.vm.ToValue(fmt.Sprintf("permission denied: not allowed to listen topic '%s' (declare it in info.yaml)", topic)))
		}

		unsub := pBus.On(topic, func(payload any) {
			rt.mu.Lock()
			defer rt.mu.Unlock()
			fn(goja.Undefined(), rt.vm.ToValue(payload))
		})

		return rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			unsub()
			return goja.Undefined()
		})
	})

	return busObj
}

func matchTopicPattern(pattern, topic string) bool {
	if pattern == topic {
		return true
	}
	if strings.HasSuffix(pattern, ":*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(topic, prefix)
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(topic, prefix)
	}
	return false
}
