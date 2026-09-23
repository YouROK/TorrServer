package web

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"silo/internal/config"
	"silo/internal/log"
	"silo/internal/plugin"
	"silo/internal/torrent"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg           *config.Config
	userSvc       *user.Service
	torrentMgr    *torrent.Manager
	pluginMgr     *plugin.Manager
	router        *gin.Engine
	pluginsRouter *PluginsRouter
	httpSrv       *http.Server
	startedAt     time.Time
	shutdownFn    func()
}

func NewServer(
	cfg *config.Config,
	userSvc *user.Service,
	torrentMgr *torrent.Manager,
) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	plugRouter := NewPluginsRouter()

	s := &Server{
		cfg:           cfg,
		userSvc:       userSvc,
		torrentMgr:    torrentMgr,
		router:        engine,
		pluginsRouter: plugRouter,
		startedAt:     time.Now(),
	}

	s.registerRoutes()
	return s
}

// SetPluginManager связывает веб-сервер с менеджером плагинов после инициализации
func (s *Server) SetPluginManager(pm *plugin.Manager) {
	s.pluginMgr = pm

	// Загружаем оба словаря ядра: en и ru
	enData, enErr := fs.ReadFile(staticFS, "static/lang/en.json")
	ruData, ruErr := fs.ReadFile(staticFS, "static/lang/ru.json")

	if enErr == nil && ruErr == nil {
		pm.I18n().SeedCore(enData, ruData)
		log.Info("[Web] Core i18n seeded: en + ru")
	} else if enErr == nil {
		pm.I18n().SeedCore(enData, nil)
		log.Warnf("[Web] Core i18n seeded: en only (ru error: %v)", ruErr)
	} else {
		log.Errorf("[Web] failed to read core lang files: en=%v, ru=%v", enErr, ruErr)
	}

	log.Info("[Web] Plugin manager linked successfully")
}

func (s *Server) registerRoutes() {
	// Публичные роуты
	s.router.GET("/i18n.js", s.handleI18nJS)

	RegisterStaticRoutes(s.router)

	s.router.POST("/api/auth/login", s.handleLogin)
	s.router.POST("/api/auth/logout", s.handleLogout)

	// Защищенные роуты API
	api := s.router.Group("/api", AuthMiddleware(s.userSvc))
	{
		api.GET("/auth/me", s.handleGetMe)

		// Системные эндпоинты
		api.GET("/system/ping", s.handlePing)
		api.GET("/system/version", s.handleGetVersion)
		api.GET("/system/logs", s.handleGetLogs)
		api.POST("/system/blocklist", s.handleSetBlocklist)
		api.POST("/system/trackers", s.handleSetTrackers)

		// Пользователи (Admin+)
		api.GET("/users", s.handleListUsers)
		api.POST("/users", s.handleCreateUser)
		api.PUT("/users/:id", s.handleUpdateUser)
		api.DELETE("/users/:id", s.handleDeleteUser)
		api.POST("/users/:id/regenerate-token", s.handleRegenerateToken)
		api.POST("/users/:id/rank", s.handleSetRank)
		api.GET("/users/:id/torrents", s.handleAdminListTorrents)

		// Торренты
		api.GET("/torrents", s.handleListTorrents)
		api.POST("/torrents", s.handleAddTorrent)
		api.GET("/torrents/:hash", s.handleGetTorrent)
		api.DELETE("/torrents/:hash", s.handleDropTorrent)
		api.GET("/torrents/:hash/cache", s.handleGetCache)
		api.POST("/torrents/:hash/files/:idx/viewed", s.handleSetFileViewed)
		api.POST("/torrents/:hash/files/:idx/preload", s.handlePreloadTorrent)
		api.POST("/torrents/:hash/wake", s.handleWakeTorrent)

		// Стриминг видео (поддерживает Range-запросы и ?token=...)
		api.GET("/stream/:hash/:fileIdx", s.handleStream)

		// Живая статистика для дашборда админки
		api.GET("/system/stats", s.handleGetStats)

		// Конфиг торрент-движка из БД (Owner+)
		api.GET("/system/torrent-config", s.handleGetTorrentConfig)
		api.POST("/system/torrent-config", s.handleSaveTorrentConfig)

		// Плагины (Owner+)
		api.GET("/plugins", s.handleListPlugins)
		api.POST("/plugins/upload", s.handleUploadPlugin)
		api.POST("/plugins/install-url", s.handleInstallPluginFromURL)
		api.DELETE("/plugins/:id", s.handleDeletePlugin)
		api.PUT("/plugins/order", s.handleSetPluginOrder)
		api.GET("/plugins/:id/info", s.handlePluginInfo)
		api.POST("/plugins/:id/enable", s.handlePluginEnable)
		api.GET("/plugins/catalog", s.handleCatalog)
		api.POST("/plugins/catalog/install", s.handleInstallFromCatalog)

		// Меню-расширения от плагинов
		api.GET("/ui/menu", s.handlePluginMenu)

		// Библиотека в формате torrs (отдельный префикс, чтобы не конфликтовать с /torrents/:hash)
		api.GET("/library/export", s.handleExportLibrary)
		api.POST("/library/import", s.handleImportLibrary)

		// Правка личной карточки раздачи
		api.PUT("/torrents/:hash/meta", s.handleUpdateTorrentMeta)

		// Завершение работы процесса (Owner+)
		api.POST("/system/shutdown", s.handleShutdown)

		// SSE-поток обновлений библиотеки для главного экрана темы
		api.GET("/events", s.handleEventsStream)
	}

	// Главная страница и плагины
	s.router.GET("/", s.pluginsRouter.HandleRoot)
	s.router.Any("/plugins/:id/*path", s.pluginsRouter.Dispatch)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.httpSrv = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       5 * time.Minute,
	}

	log.Infof("[Web] Server listening on http://%s", addr)
	if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("[Web] Stopping HTTP server...")
	if s.httpSrv != nil {
		return s.httpSrv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) GetPluginsRouter() *PluginsRouter {
	return s.pluginsRouter
}

func (s *Server) SetShutdownFunc(fn func()) {
	s.shutdownFn = fn
}
