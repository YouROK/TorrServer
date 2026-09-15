package web

import (
	"errors"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"

	"silo/internal/log"

	"github.com/gin-gonic/gin"
)

type RouteEntry struct {
	Route   string
	VFS     fs.FS
	VFSPath string
	Handler gin.HandlerFunc
}

type PluginRouter struct {
	exactRoutes  map[string]RouteEntry
	prefixRoutes []RouteEntry
}

type PluginsRouter struct {
	mu            sync.RWMutex
	activeThemeID string
	routers       map[string]*PluginRouter
}

func NewPluginsRouter() *PluginsRouter {
	return &PluginsRouter{
		routers: make(map[string]*PluginRouter),
	}
}

func (pr *PluginsRouter) Reset() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.activeThemeID = ""
	pr.routers = make(map[string]*PluginRouter)
	log.Debug("[Web] Plugins router reset")
}

func (pr *PluginsRouter) SetActiveTheme(pluginID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.activeThemeID = pluginID
	log.Infof("[Web] Active theme UI set to plugin '%s'", pluginID)
}

func (pr *PluginsRouter) getOrCreatePluginRouter(pluginID string) *PluginRouter {
	if r, exists := pr.routers[pluginID]; exists {
		return r
	}
	r := &PluginRouter{
		exactRoutes:  make(map[string]RouteEntry),
		prefixRoutes: make([]RouteEntry, 0),
	}
	pr.routers[pluginID] = r
	return r
}

func (pr *PluginsRouter) AddJSHandler(pluginID, method, route string, handler gin.HandlerFunc) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	r := pr.getOrCreatePluginRouter(pluginID)
	key := method + ":" + path.Clean(route)
	r.exactRoutes[key] = RouteEntry{Route: route, Handler: handler}
	log.Debugf("[Web] Plugin '%s' registered JS handler for %s %s", pluginID, method, route)
}

func (pr *PluginsRouter) AddStaticFile(pluginID, route, vfsPath string, vfs fs.FS) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	r := pr.getOrCreatePluginRouter(pluginID)
	key := "GET:" + path.Clean(route)
	r.exactRoutes[key] = RouteEntry{Route: route, VFS: vfs, VFSPath: vfsPath}
	log.Debugf("[Web] Plugin '%s' registered static file %s -> %s", pluginID, route, vfsPath)
}

func (pr *PluginsRouter) AddStaticDir(pluginID, route, vfsDir string, vfs fs.FS) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	r := pr.getOrCreatePluginRouter(pluginID)
	cleanRoute := strings.TrimSuffix(route, "/*")
	r.prefixRoutes = append(r.prefixRoutes, RouteEntry{Route: cleanRoute, VFS: vfs, VFSPath: vfsDir})
	log.Debugf("[Web] Plugin '%s' registered static dir %s/* -> %s/", pluginID, cleanRoute, vfsDir)
}

func (pr *PluginsRouter) HandleRoot(c *gin.Context) {
	pr.mu.RLock()
	themeID := pr.activeThemeID
	pr.mu.RUnlock()

	if themeID != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/plugins/"+themeID+"/index.html")
		return
	}
	c.String(http.StatusOK, "Silo Core API is running. No theme UI plugin installed.")
}

func (pr *PluginsRouter) Dispatch(c *gin.Context) {
	pluginID := c.Param("id")
	reqPath := path.Clean(c.Param("path"))
	if reqPath == "" {
		reqPath = "/"
	}
	key := c.Request.Method + ":" + reqPath

	pr.mu.RLock()
	pluginR, exists := pr.routers[pluginID]
	pr.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found or has no routes"})
		return
	}

	// 1. Точное совпадение
	if entry, found := pluginR.exactRoutes[key]; found {
		pr.serveEntry(c, entry, entry.VFSPath)
		return
	}

	// 2. Префиксные папки
	if c.Request.Method == http.MethodGet {
		var bestMatch *RouteEntry
		for i := range pluginR.prefixRoutes {
			entry := &pluginR.prefixRoutes[i]
			if entry.Route == "/" || reqPath == entry.Route || strings.HasPrefix(reqPath, entry.Route+"/") {
				if bestMatch == nil || len(entry.Route) > len(bestMatch.Route) {
					bestMatch = entry
				}
			}
		}

		if bestMatch != nil {
			relPath := strings.TrimPrefix(reqPath, bestMatch.Route)
			relPath = strings.TrimPrefix(relPath, "/")
			targetFile := path.Join(bestMatch.VFSPath, relPath)
			pr.serveEntry(c, *bestMatch, targetFile)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "route not found in plugin"})
}

func (pr *PluginsRouter) serveEntry(c *gin.Context, entry RouteEntry, targetFile string) {
	if entry.Handler != nil {
		entry.Handler(c)
		return
	}

	cleanFile := strings.TrimPrefix(targetFile, "/")
	data, err := fs.ReadFile(entry.VFS, cleanFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found in VFS"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		}
		return
	}

	contentType := detectContentType(cleanFile, data)
	c.Data(http.StatusOK, contentType, data)
}

// RemovePlugin полностью удаляет все зарегистрированные маршруты конкретного плагина
func (pr *PluginsRouter) RemovePlugin(pluginID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	delete(pr.routers, pluginID)
	if pr.activeThemeID == pluginID {
		pr.activeThemeID = ""
	}
	log.Debugf("[Web] Removed all routes for plugin '%s'", pluginID)
}

// ActiveThemeID возвращает ID плагина, установленного темой главной страницы
func (pr *PluginsRouter) ActiveThemeID() string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	return pr.activeThemeID
}

func detectContentType(filePath string, data []byte) string {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webmanifest":
		return "application/manifest+json"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".html":
		return "text/html; charset=utf-8"
	case ".json":
		return "application/json"
	}

	ct := mime.TypeByExtension(ext)
	if ct == "" {
		ct = http.DetectContentType(data)
	}
	return ct
}
