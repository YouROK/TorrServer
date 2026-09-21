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

// createVFSModule — объект ts.vfs: чтение собственных файлов плагина
// через изолированную VFS (fs.FS, привязанную к папке или ZIP плагина).
//
// Безопасность обеспечивает сама VFS. os.DirFS и *zip.Reader реализуют
// fs.FS и внутри вызывают fs.ValidPath, который отвергает:
//   - пустые сегменты и ведущий/замыкающий "/",
//   - сегменты "." и "..",
//   - невалидный UTF-8.
//
// Это отсекает классический path traversal на Linux/macOS. Единственное,
// что VFS не закрывает — обратные слэши на Windows: fs.ValidPath считает
// `a\..\b` валидным (backslash для него обычный символ), а
// filepath.FromSlash его не преобразует. В итоге filepath.Join разрешает
// `..\..` относительно реальной ФС и выходит за пределы VFS. Поэтому
// backslash отвергаем явно, до вызова vfs.Open.
//
// path.Clean применяется, чтобы авторы плагинов могли писать `./x`
// вместо `x`. Clean не создаёт новых сегментов — он только схлопывает
// `a/../b` в `b` и убирает `.`. Итог всё равно проверяется VFS.
//
// Ошибки передаются в JS через panic: goja на границе VM превращает
// panic(vm.ToValue(...)) в JS-исключение, которое плагин может поймать
// через try/catch. Это тот же стиль, что в bind_http.go, bind_torrent.go,
// bind_storage.go — консистентно с остальными модулями ts.*.
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

	// ts.vfs.exists(path) -> bool.
	// Boolean-запрос: «ошибка пути» и «файла нет» для плагина равнозначны false.
	obj.Set("exists", func(call goja.FunctionCall) goja.Value {
		p, err := checkPath(call.Argument(0).String())
		if err != nil {
			return rt.vm.ToValue(false)
		}
		_, err = fs.Stat(vfs, p)
		return rt.vm.ToValue(err == nil)
	})

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
			// *zip.Reader часто отдаёт нулевое время — нормализуем в epoch,
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
