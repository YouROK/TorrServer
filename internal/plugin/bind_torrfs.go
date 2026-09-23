package plugin

import (
	"fmt"

	"silo/internal/log"
	"silo/internal/torrfs"
	"silo/internal/user"

	"github.com/dop251/goja"
)

func (rt *JSRuntime) createTorrFSModule(fs *torrfs.TorrFS, userSvc *user.Service) *goja.Object {
	obj := rt.vm.NewObject()

	resolveUser := func(id string) *user.User {
		if id == "" {
			panic(rt.vm.ToValue("user id is required"))
		}
		u, err := userSvc.GetUserByID(id)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("user not found: %s", id)))
		}
		return u
	}

	obj.Set("list", func(call goja.FunctionCall) goja.Value {
		u := resolveUser(call.Argument(0).String())
		path := call.Argument(1).String()

		nodes, err := fs.List(u, path)
		if err != nil {
			if err == torrfs.ErrNotFound || err == torrfs.ErrNotDir {
				return goja.Null()
			}
			panic(rt.vm.ToValue(fmt.Sprintf("torrfs.list %q: %v", path, err)))
		}

		result := make([]map[string]any, 0, len(nodes))
		for _, n := range nodes {
			result = append(result, nodeToMap(n))
		}
		return rt.vm.ToValue(result)
	})

	obj.Set("stat", func(call goja.FunctionCall) goja.Value {
		u := resolveUser(call.Argument(0).String())
		path := call.Argument(1).String()

		n, err := fs.Stat(u, path)
		if err != nil {
			if err == torrfs.ErrNotFound {
				return goja.Null()
			}
			panic(rt.vm.ToValue(fmt.Sprintf("torrfs.stat %q: %v", path, err)))
		}
		return rt.vm.ToValue(nodeToMap(n))
	})

	obj.Set("open", func(call goja.FunctionCall) goja.Value {
		u := resolveUser(call.Argument(0).String())
		path := call.Argument(1).String()

		h, err := fs.OpenFile(u, path)
		if err != nil {
			if err == torrfs.ErrNotFound {
				return goja.Null()
			}
			panic(rt.vm.ToValue(fmt.Sprintf("torrfs.open %q: %v", path, err)))
		}

		rt.pendingHandles = append(rt.pendingHandles, h)
		return rt.vm.ToValue(h)
	})

	obj.Set("close", func(call goja.FunctionCall) goja.Value {
		h, ok := call.Argument(0).Export().(*torrfs.Handle)
		if !ok || h == nil {
			panic(rt.vm.ToValue("torrfs.close: expected handle"))
		}
		_ = h.Close()
		rt.removePendingHandle(h)
		return goja.Undefined()
	})

	log.Debugf("[Plugin:%s] TorrFS module initialized", rt.pluginID)
	return obj
}

// nodeToMap превращает узел дерева в JS-объект.
func nodeToMap(n torrfs.Node) map[string]any {
	m := map[string]any{
		"name":   n.Name(),
		"path":   n.Path(),
		"is_dir": n.IsDir(),
		"size":   n.Size(),
		"mtime":  n.ModTime().Unix(),
	}
	if !n.IsDir() {
		m["mime"] = n.MimeType()
	}
	return m
}
