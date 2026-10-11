package plugin

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"silo/internal/log"

	"github.com/dop251/goja"
)

// createVFSModule - объект ts.vfs: чтение собственных файлов плагина
// через изолированную VFS (fs.FS, привязанную к папке или ZIP плагина).
// Обратные слэши отвергаются явно: fs.ValidPath считает их обычными
// символами и пропускает `..\..` на Windows.
//
// API:
//
//	ts.vfs.read(path)        -> string
//	ts.vfs.exists(path)      -> bool
//	ts.vfs.stat(path)        -> {name, size, is_dir, mode, mod_time}
//	ts.vfs.list([path="."])  -> [{name, size, is_dir}, ...]
func (rt *JSRuntime) createVFSModule(vfs fs.FS) *goja.Object {
	obj := rt.vm.NewObject()

	// checkPath нормализует путь и отвергает backslash (см. комментарий выше).
	// Возвращает путь, пригодный для передачи в fs.FS.
	checkPath := func(raw string) (string, error) {
		if strings.ContainsRune(raw, '\\') {
			return "", fmt.Errorf("backslashes are not allowed in VFS paths, use forward slashes")
		}
		return path.Clean(raw), nil
	}

	// Читает файл как строку:
	// ts.vfs.read(path) -> string
	obj.Set("read", func(call goja.FunctionCall) goja.Value {
		p, err := checkPath(call.Argument(0).String())
		if err != nil {
			panic(rt.vm.ToValue(err.Error()))
		}

		data, err := fs.ReadFile(vfs, p)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("vfs.read %q: %v", p, err)))
		}
		return rt.vm.ToValue(string(data))
	})

	// Проверяет наличие файла:
	// ts.vfs.exists(path) -> bool
	// Boolean-запрос: «ошибка пути» и «файла нет» для плагина равнозначны false.
	obj.Set("exists", func(call goja.FunctionCall) goja.Value {
		p, err := checkPath(call.Argument(0).String())
		if err != nil {
			return rt.vm.ToValue(false)
		}
		_, err = fs.Stat(vfs, p)
		return rt.vm.ToValue(err == nil)
	})

	// Сведения о файле:
	// ts.vfs.stat(path) -> {name, size, is_dir, mode, mod_time}
	obj.Set("stat", func(call goja.FunctionCall) goja.Value {
		p, err := checkPath(call.Argument(0).String())
		if err != nil {
			panic(rt.vm.ToValue(err.Error()))
		}

		info, err := fs.Stat(vfs, p)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("vfs.stat %q: %v", p, err)))
		}

		modTime := info.ModTime()
		if modTime.IsZero() {
			// *zip.Reader часто отдаёт нулевое время - нормализуем в epoch,
			// чтобы плагин не получал отрицательный timestamp.
			modTime = time.Unix(0, 0).UTC()
		}

		return rt.vm.ToValue(map[string]any{
			"name":     info.Name(),
			"size":     info.Size(),
			"is_dir":   info.IsDir(),
			"mode":     info.Mode().String(),
			"mod_time": modTime.Unix(),
		})
	})

	// Содержимое каталога:
	// ts.vfs.list([path="."]) -> [{name, size, is_dir}, ...]
	obj.Set("list", func(call goja.FunctionCall) goja.Value {
		p := "."
		if len(call.Arguments) > 0 &&
			!goja.IsUndefined(call.Argument(0)) &&
			!goja.IsNull(call.Argument(0)) {
			cp, err := checkPath(call.Argument(0).String())
			if err != nil {
				panic(rt.vm.ToValue(err.Error()))
			}
			p = cp
		}

		entries, err := fs.ReadDir(vfs, p)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("vfs.list %q: %v", p, err)))
		}

		result := make([]map[string]any, 0, len(entries))
		for _, e := range entries {
			item := map[string]any{
				"name":   e.Name(),
				"is_dir": e.IsDir(),
				"size":   int64(0),
			}
			if info, err := e.Info(); err == nil {
				item["size"] = info.Size()
			}
			result = append(result, item)
		}
		return rt.vm.ToValue(result)
	})

	log.Debugf("[Plugin:%s] VFS module initialized", rt.pluginID)
	return obj
}
