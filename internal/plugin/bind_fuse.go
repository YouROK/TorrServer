package plugin

import (
	"errors"
	"fmt"

	"silo/internal/fusefs"

	"github.com/dop251/goja"
)

// ownerUserID - владелец библиотеки, которую отдаёт файловая система.
const ownerUserID = "owner"

// createFuseModule собирает объект ts.fuse: монтирование библиотеки владельца.
//
// API:
//
//	ts.fuse.supported()          -> bool
//	ts.fuse.mount(path, opts)    -> {ok, path, mount_time} | {ok: false, code, error}
//	ts.fuse.unmount()            -> {ok} | {ok: false, code, error}
//	ts.fuse.status()             -> {supported, mounted, path, mount_time}
func (rt *JSRuntime) createFuseModule(mgr *fusefs.Manager) *goja.Object {
	obj := rt.vm.NewObject()

	obj.Set("supported", func(call goja.FunctionCall) goja.Value {
		return rt.vm.ToValue(fusefs.Supported())
	})

	obj.Set("mount", func(call goja.FunctionCall) goja.Value {
		if mgr == nil {
			return rt.fuseError(fusefs.ErrNoSource)
		}
		if !fusefs.Supported() {
			return rt.fuseError(fusefs.ErrUnsupported)
		}

		path := call.Argument(0).String()
		allowOther := false
		if opts := call.Argument(1); opts != nil && !goja.IsUndefined(opts) && !goja.IsNull(opts) {
			if obj, ok := opts.Export().(map[string]any); ok {
				allowOther, _ = obj["allow_other"].(bool)
			}
		}

		// Источник задаётся только когда он есть: без него менеджер
		// ответит no_source, проверив сначала сам путь.
		if rt.torrFS != nil && rt.userSvc != nil {
			u, err := rt.userSvc.GetUserByID(ownerUserID)
			if err != nil {
				return rt.fuseError(fmt.Errorf("owner not found: %w", err))
			}
			mgr.SetSource(&torrFSAdapter{fs: rt.torrFS, user: u})
		}

		status, err := mgr.Mount(path, fusefs.Options{AllowOther: allowOther})
		if err != nil {
			return rt.fuseError(err)
		}

		return rt.vm.ToValue(map[string]any{
			"ok":         true,
			"path":       status.Path,
			"mount_time": status.MountTime,
		})
	})

	obj.Set("unmount", func(call goja.FunctionCall) goja.Value {
		if mgr == nil {
			return rt.fuseError(fusefs.ErrNotMounted)
		}

		status, err := mgr.Unmount()
		if err != nil {
			return rt.fuseError(err)
		}

		return rt.vm.ToValue(map[string]any{
			"ok":   true,
			"path": status.Path,
		})
	})

	obj.Set("status", func(call goja.FunctionCall) goja.Value {
		result := map[string]any{
			"supported": fusefs.Supported(),
			"mounted":   false,
		}

		if mgr == nil {
			return rt.vm.ToValue(result)
		}
		if status, mounted := mgr.Status(); mounted {
			result["mounted"] = true
			result["path"] = status.Path
			result["mount_time"] = status.MountTime
		}
		return rt.vm.ToValue(result)
	})

	return obj
}

// fuseError превращает ошибку монтирования в ответ с кодом для перевода
// и подробностью для показа.
func (rt *JSRuntime) fuseError(err error) goja.Value {
	result := map[string]any{
		"ok":    false,
		"code":  fuseErrorCode(err),
		"error": err.Error(),
	}

	var fe *fusefs.Error
	if errors.As(err, &fe) && fe.Detail != "" {
		result["detail"] = fe.Detail
	}
	return rt.vm.ToValue(result)
}

func fuseErrorCode(err error) string {
	switch {
	case errors.Is(err, fusefs.ErrUnsupported):
		return "unsupported"
	case errors.Is(err, fusefs.ErrEmptyPath):
		return "empty_path"
	case errors.Is(err, fusefs.ErrNotAbsolute):
		return "not_absolute"
	case errors.Is(err, fusefs.ErrNotExist):
		return "not_exist"
	case errors.Is(err, fusefs.ErrPermission):
		return "permission"
	case errors.Is(err, fusefs.ErrNotDir):
		return "not_a_directory"
	case errors.Is(err, fusefs.ErrAlreadyMounted):
		return "already_mounted"
	case errors.Is(err, fusefs.ErrNotMounted):
		return "not_mounted"
	case errors.Is(err, fusefs.ErrNoSource):
		return "no_source"
	case errors.Is(err, fusefs.ErrAllowOther):
		return "allow_other"
	case errors.Is(err, fusefs.ErrNoDevice):
		return "no_device"
	default:
		return "mount_failed"
	}
}
