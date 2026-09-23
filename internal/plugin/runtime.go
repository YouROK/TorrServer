package plugin

import (
	"io/fs"
	"silo/internal/torrfs"
	"sync"

	"silo/internal/database"
	"silo/internal/torrent"
	"silo/internal/user"

	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
)

type WebRegistrar interface {
	Reset()
	SetActiveTheme(pluginID string)
	ActiveThemeID() string
	RemovePlugin(pluginID string)
	AddJSHandler(pluginID, method, route string, handler gin.HandlerFunc)
	AddStaticFile(pluginID, route, vfsFile string, vfs fs.FS)
	AddStaticDir(pluginID, route, vfsDir string, vfs fs.FS)
}

type JSRuntime struct {
	pluginID string
	i18n     *I18nRegistry
	vm       *goja.Runtime
	userSvc  *user.Service
	mu       sync.Mutex

	pendingHandles []*torrfs.Handle
}

func NewJSRuntime(
	pluginID string,
	manifest *Manifest,
	vfs fs.FS,
	registrar WebRegistrar,
	db *database.DB,
	torrMgr *torrent.Manager,
	userSvc *user.Service,
	i18nReg *I18nRegistry,
	torrFS *torrfs.TorrFS,
) *JSRuntime {
	vm := goja.New()
	rt := &JSRuntime{
		pluginID: pluginID,
		vm:       vm,
		i18n:     i18nReg,
		userSvc:  userSvc,
	}

	// 1. Системная консоль (глобально)
	rt.bindConsole()

	// 2. Единое пространство имен "ts" и "vault"
	tsObj := vm.NewObject()
	tsObj.Set("info", rt.createInfoModule())
	tsObj.Set("bus", rt.createBusModule(manifest))
	tsObj.Set("web", rt.createWebModule(manifest, vfs, registrar))
	tsObj.Set("http", rt.createHTTPModule())
	tsObj.Set("storage", rt.createStorageModule(db))
	tsObj.Set("torrent", rt.createTorrentModule(torrMgr, userSvc))
	tsObj.Set("users", rt.createUsersModule(userSvc))
	tsObj.Set("i18n", rt.createI18nModule())
	tsObj.Set("vfs", rt.createVFSModule(vfs))
	tsObj.Set("crypto", rt.createCryptoModule())
	tsObj.Set("torrfs", rt.createTorrFSModule(torrFS, userSvc))

	vm.Set("ts", tsObj)

	return rt
}

func (rt *JSRuntime) Execute(code string) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	_, err := rt.vm.RunString(code)
	return err
}

func (rt *JSRuntime) Stop() {
	rt.vm.Interrupt("plugin unloaded")
}

// removePendingHandle убирает handle из списка текущего запроса.
func (rt *JSRuntime) removePendingHandle(h *torrfs.Handle) {
	for i, x := range rt.pendingHandles {
		if x == h {
			rt.pendingHandles = append(rt.pendingHandles[:i], rt.pendingHandles[i+1:]...)
			return
		}
	}
}
