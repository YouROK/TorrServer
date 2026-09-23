package plugin

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"silo/internal/torrfs"
	"strings"

	"silo/internal/log"

	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
)

// maxRequestBody — лимит чтения тела запроса (10 МБ)
const maxRequestBody = 10 * 1024 * 1024

// webResponse — накопитель ответа, который JS-обработчик заполняет через res.*
type webResponse struct {
	used        bool
	status      int
	contentType string
	headers     map[string]string
	bodyBytes   []byte
	empty       bool

	streamHandle *torrfs.Handle
	streamName   string
}

// createWebModule — объект ts.web: собственные HTTP-роуты и статика плагина
func (rt *JSRuntime) createWebModule(manifest *Manifest, vfs fs.FS, registrar WebRegistrar) *goja.Object {
	obj := rt.vm.NewObject()

	// Основные HTTP-методы: ts.web.get(route, handler), ts.web.post(...) и т.д.
	obj.Set("get", rt.webRegistrar(registrar, manifest, http.MethodGet))
	obj.Set("post", rt.webRegistrar(registrar, manifest, http.MethodPost))
	obj.Set("put", rt.webRegistrar(registrar, manifest, http.MethodPut))
	obj.Set("patch", rt.webRegistrar(registrar, manifest, http.MethodPatch))
	obj.Set("delete", rt.webRegistrar(registrar, manifest, http.MethodDelete))
	obj.Set("head", rt.webRegistrar(registrar, manifest, http.MethodHead))
	obj.Set("options", rt.webRegistrar(registrar, manifest, http.MethodOptions))

	// ts.web.any(route, handler) — обработчик сразу на все основные методы
	obj.Set("any", func(call goja.FunctionCall) goja.Value {
		route := call.Argument(0).String()
		handler, ok := rt.exportHandler(call.Argument(1))
		if !ok {
			return goja.Undefined()
		}
		methods := []string{
			http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions,
		}
		for _, m := range methods {
			rt.registerRoute(registrar, manifest, m, route, handler)
		}
		return goja.Undefined()
	})

	// ts.web.route(method, route, handler) — произвольный HTTP-метод
	obj.Set("route", func(call goja.FunctionCall) goja.Value {
		method := strings.ToUpper(call.Argument(0).String())
		route := call.Argument(1).String()
		handler, ok := rt.exportHandler(call.Argument(2))
		if !ok {
			return goja.Undefined()
		}
		rt.registerRoute(registrar, manifest, method, route, handler)
		return goja.Undefined()
	})

	// ts.web.staticFile(route, vfsPath) — отдать один файл из своей VFS
	obj.Set("staticFile", func(call goja.FunctionCall) goja.Value {
		route := call.Argument(0).String()
		vfsPath := call.Argument(1).String()
		if !routeAllowed(manifest, route) {
			log.Warnf("[Plugin:%s] route '%s' is not in manifest whitelist, static file skipped", rt.pluginID, route)
			return goja.Undefined()
		}
		registrar.AddStaticFile(rt.pluginID, route, vfsPath, vfs)
		return goja.Undefined()
	})

	// ts.web.staticDir(route, vfsDir) — отдать директорию из своей VFS
	obj.Set("staticDir", func(call goja.FunctionCall) goja.Value {
		route := call.Argument(0).String()
		vfsDir := call.Argument(1).String()

		// Автоматически добавляем /* если его нет
		if !strings.HasSuffix(route, "/*") {
			route = strings.TrimSuffix(route, "/") + "/*"
		}

		if !routeAllowed(manifest, route) {
			log.Warnf("[Plugin:%s] route '%s' is not in manifest whitelist, static dir skipped", rt.pluginID, route)
			return goja.Undefined()
		}
		registrar.AddStaticDir(rt.pluginID, route, vfsDir, vfs)
		log.Debugf("[Plugin:%s] registered static dir %s -> %s/", rt.pluginID, route, vfsDir)
		return goja.Undefined()
	})

	return obj
}

// webRegistrar возвращает функцию регистрации обработчика конкретного метода
func (rt *JSRuntime) webRegistrar(registrar WebRegistrar, manifest *Manifest, method string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		route := call.Argument(0).String()
		handler, ok := rt.exportHandler(call.Argument(1))
		if !ok {
			return goja.Undefined()
		}
		rt.registerRoute(registrar, manifest, method, route, handler)
		return goja.Undefined()
	}
}

// registerRoute проверяет белый список манифеста и регистрирует обработчик
func (rt *JSRuntime) registerRoute(registrar WebRegistrar, manifest *Manifest, method, route string, handler goja.Callable) {
	if !routeAllowed(manifest, route) {
		log.Warnf("[Plugin:%s] route '%s' is not in manifest whitelist, handler skipped", rt.pluginID, route)
		return
	}
	registrar.AddJSHandler(rt.pluginID, method, route, rt.ginHandler(handler))
	log.Debugf("[Plugin:%s] registered %s %s", rt.pluginID, method, route)
}

// exportHandler приводит значение goja к вызываемой функции
func (rt *JSRuntime) exportHandler(v goja.Value) (goja.Callable, bool) {
	fn, ok := goja.AssertFunction(v)
	if !ok {
		log.Warnf("[Plugin:%s] web handler is not a function", rt.pluginID)
	}
	return fn, ok
}

// ginHandler оборачивает JS-обработчик в gin.HandlerFunc
func (rt *JSRuntime) ginHandler(handler goja.Callable) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, _ := io.ReadAll(io.LimitReader(c.Request.Body, maxRequestBody))

		rt.mu.Lock()
		rt.pendingHandles = nil
		res := rt.runWebHandler(handler, c, raw)
		opened := rt.pendingHandles
		rt.pendingHandles = nil
		rt.mu.Unlock()

		for k, v := range res.headers {
			c.Header(k, v)
		}

		if res.empty {
			c.Status(res.status)
			return
		}

		if res.streamHandle != nil {
			defer res.streamHandle.Close()
			for _, h := range opened {
				if h != res.streamHandle {
					_ = h.Close()
				}
			}

			name := res.streamName
			if name == "" {
				name = res.streamHandle.Name()
			}
			http.ServeContent(c.Writer, c.Request, name, res.streamHandle.ModTime(), res.streamHandle)
			return
		}

		for _, h := range opened {
			_ = h.Close()
		}

		ct := res.contentType
		if ct == "" {
			ct = "text/plain; charset=utf-8"
		}
		c.Data(res.status, ct, res.bodyBytes)
	}
}

// runWebHandler строит req/res, вызывает JS и возвращает готовый HTTP-ответ.
// Вызывается под заблокированным мьютексом рантайма.
func (rt *JSRuntime) runWebHandler(handler goja.Callable, c *gin.Context, raw []byte) *webResponse {
	vm := rt.vm
	res := &webResponse{status: http.StatusOK, headers: make(map[string]string)}

	req := vm.NewObject()
	req.Set("method", c.Request.Method)
	req.Set("path", c.Request.URL.Path)
	req.Set("raw", string(raw))

	// Query-параметры
	query := vm.NewObject()
	for k, vals := range c.Request.URL.Query() {
		if len(vals) > 0 {
			query.Set(k, vals[0])
		}
	}
	req.Set("query", query)

	// Заголовки запроса (ключи в нижнем регистре)
	headers := vm.NewObject()
	for k, vals := range c.Request.Header {
		if len(vals) > 0 {
			headers.Set(strings.ToLower(k), vals[0])
		}
	}
	req.Set("headers", headers)

	// Тело: JSON -> объект, urlencoded-форма -> объект, иначе сырая строка
	req.Set("body", rt.parseBody(c, raw))

	// Пользователь: опциональная аутентификация по токену из cookie/query/заголовка
	req.Set("user", rt.currentUserObject(c))

	// res в духе Express: res.status(200).json({...}), цепочки поддерживаются
	var resObj *goja.Object
	resObj = vm.NewObject()

	resObj.Set("status", func(call goja.FunctionCall) goja.Value {
		res.used = true
		res.status = int(call.Argument(0).ToInteger())
		return resObj
	})
	resObj.Set("header", func(call goja.FunctionCall) goja.Value {
		res.used = true
		res.headers[call.Argument(0).String()] = call.Argument(1).String()
		return resObj
	})
	resObj.Set("json", func(call goja.FunctionCall) goja.Value {
		res.used = true
		res.contentType = "application/json; charset=utf-8"
		data, err := json.Marshal(call.Argument(0).Export())
		if err != nil {
			data = []byte(`{"error":"failed to marshal response"}`)
		}
		res.bodyBytes = data
		return resObj
	})
	resObj.Set("text", func(call goja.FunctionCall) goja.Value {
		res.used = true
		res.contentType = "text/plain; charset=utf-8"
		res.bodyBytes = []byte(call.Argument(0).String())
		return resObj
	})
	resObj.Set("html", func(call goja.FunctionCall) goja.Value {
		res.used = true
		res.contentType = "text/html; charset=utf-8"
		res.bodyBytes = []byte(call.Argument(0).String())
		return resObj
	})
	resObj.Set("stream", func(call goja.FunctionCall) goja.Value {
		h, ok := call.Argument(0).Export().(*torrfs.Handle)
		if !ok || h == nil {
			panic(rt.vm.ToValue("res.stream: first argument must be a torrfs handle"))
		}
		name := call.Argument(1).String()
		res.used = true
		res.streamHandle = h
		res.streamName = name
		return resObj
	})
	resObj.Set("end", func(call goja.FunctionCall) goja.Value {
		res.used = true
		res.empty = true
		return resObj
	})

	result, err := handler(goja.Undefined(), req, resObj)
	if err != nil {
		log.Errorf("[Plugin:%s] web handler error: %v", rt.pluginID, err)
		return &webResponse{
			used:        true,
			status:      http.StatusInternalServerError,
			contentType: "application/json; charset=utf-8",
			headers:     make(map[string]string),
			bodyBytes:   []byte(`{"error":"plugin handler error"}`),
		}
	}

	// Обработчик ничего не записал в res: используем возвращенное значение.
	// undefined -> пустой ответ, строка начинающаяся с '<' -> HTML,
	// прочая строка -> text/plain, объект/массив -> JSON
	if !res.used {
		switch {
		case result == nil || goja.IsUndefined(result) || goja.IsNull(result):
			res.bodyBytes = nil
		case strings.HasPrefix(strings.TrimSpace(result.String()), "<"):
			res.contentType = "text/html; charset=utf-8"
			res.bodyBytes = []byte(result.String())
		default:
			exported := result.Export()
			if _, isString := exported.(string); isString {
				res.contentType = "text/plain; charset=utf-8"
				res.bodyBytes = []byte(result.String())
			} else {
				res.contentType = "application/json; charset=utf-8"
				data, mErr := json.Marshal(exported)
				if mErr != nil {
					data = []byte(`{"error":"failed to marshal response"}`)
				}
				res.bodyBytes = data
			}
		}
	}

	return res
}

// parseBody разбирает тело запроса по Content-Type
func (rt *JSRuntime) parseBody(c *gin.Context, raw []byte) goja.Value {
	ct := strings.ToLower(c.Request.Header.Get("Content-Type"))

	if strings.Contains(ct, "application/json") && len(raw) > 0 {
		var parsed any
		if err := json.Unmarshal(raw, &parsed); err == nil {
			return rt.vm.ToValue(parsed)
		}
	}

	if strings.Contains(ct, "application/x-www-form-urlencoded") && len(raw) > 0 {
		form := map[string]any{}
		if vals, err := url.ParseQuery(string(raw)); err == nil {
			for k, v := range vals {
				if len(v) > 0 {
					form[k] = v[0]
				}
			}
		}
		return rt.vm.ToValue(form)
	}

	return rt.vm.ToValue(string(raw))
}

// currentUserObject возвращает профиль пользователя по токену или null.
// Имя куки дублируем строкой, чтобы не тянуть циклический импорт пакета web.
func (rt *JSRuntime) currentUserObject(c *gin.Context) goja.Value {
	if rt.userSvc == nil {
		return goja.Null()
	}

	token := c.Query("token")
	if token == "" {
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		}
	}
	if token == "" {
		token, _ = c.Cookie("silo_token")
	}

	u, err := rt.userSvc.Authenticate(token)
	if err != nil || u == nil {
		return goja.Null()
	}

	obj := rt.vm.NewObject()
	obj.Set("id", u.ID)
	obj.Set("username", u.Username)
	obj.Set("rank", int(u.Rank))
	obj.Set("is_banned", u.IsBanned)
	return obj
}

// routeAllowed проверяет маршрут по белому списку manifest.Routes.
// Поддерживает точное совпадение и префиксы вида "/css/*"
func routeAllowed(manifest *Manifest, route string) bool {
	clean := "/" + strings.Trim(route, "/")
	for _, allowed := range manifest.Routes {
		allowed = "/" + strings.Trim(allowed, "/")
		if allowed == clean {
			return true
		}
		if strings.HasSuffix(allowed, "/*") && strings.HasPrefix(clean, strings.TrimSuffix(allowed, "*")) {
			return true
		}
	}
	return false
}
