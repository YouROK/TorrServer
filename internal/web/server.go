package web

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sync"
	"time"

	"silo/internal/config"
	"silo/internal/ffmpeg"
	"silo/internal/ffmpeg/profile"
	"silo/internal/ffmpeg/source"
	"silo/internal/log"
	"silo/internal/plugin"
	"silo/internal/torrent"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg            *config.Config
	userSvc        *user.Service
	torrentMgr     *torrent.Manager
	pluginMgr      *plugin.Manager
	transcoder     *ffmpeg.FFmpeg
	profiles       *profile.Service
	router         *gin.Engine
	pluginsRouter  *PluginsRouter
	streamTracker  *StreamTracker
	sourceRegistry *source.Registry
	hlsOnce        sync.Once
	hlsRegistry    *hlsRegistry
	httpSrv        *http.Server
	startedAt      time.Time
	shutdownFn     func()

	// Адрес, по которому сервер доступен с этой же машины
	sourceMu   sync.RWMutex
	sourceHost string
	sourcePort int

	// sourceOpen подменяет открытие файла раздачи в тестах
	sourceOpen sourceOpener
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
		cfg:            cfg,
		userSvc:        userSvc,
		torrentMgr:     torrentMgr,
		router:         engine,
		pluginsRouter:  plugRouter,
		streamTracker:  NewStreamTracker(),
		sourceRegistry: source.NewRegistry(0),
		hlsRegistry:    newHLSRegistry(),
		startedAt:      time.Now(),
	}

	s.registerRoutes()
	return s
}

// SetTranscoder связывает веб-сервер с модулем транскодирования.
func (s *Server) SetTranscoder(t *ffmpeg.FFmpeg) {
	s.transcoder = t
}

// SetProfileService связывает веб-сервер с сервисом профилей транскодирования.
func (s *Server) SetProfileService(svc *profile.Service) {
	s.profiles = svc
}

// stopTimeout возвращает время ожидания мягкой остановки процессов ffmpeg.
func (s *Server) stopTimeout() time.Duration {
	return ffmpeg.DefaultStopTimeout
}

// transcodeEnabled сообщает, доступно ли транскодирование.
func (s *Server) transcodeEnabled() bool {
	return s.transcoder != nil && s.transcoder.Enabled()
}

// transcodeReason возвращает причину недоступности транскодирования.
func (s *Server) transcodeReason() string {
	if s.transcoder == nil {
		return "transcoding module is not linked"
	}
	return s.transcoder.Reason()
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

	// Ping сервера
	s.router.GET("/api/system/ping", s.handlePing)

	// Внутренний источник для ffmpeg: доступ ограничен адресом отправителя
	s.router.GET("/api/internal/source/:id", s.handleInternalSource)

	// Защищенные роуты API
	api := s.router.Group("/api", AuthMiddleware(s.userSvc))
	{
		api.GET("/auth/me", s.handleGetMe)

		// Системные эндпоинты
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
		api.GET("/torrents/:hash/peers", s.handleGetPeers)
		api.POST("/torrents/:hash/files/:idx/viewed", s.handleSetFileViewed)
		api.POST("/torrents/:hash/files/:idx/preload", s.handlePreloadTorrent)
		api.POST("/torrents/:hash/wake", s.handleWakeTorrent)

		// Стриминг видео (поддерживает Range-запросы и ?token=...)
		api.GET("/stream/:hash/:fileIdx", s.handleStream)

		// Транскодирование: сведения о файле, поток, список и остановка сессий
		api.GET("/transcode/module", s.handleTranscodeModule)
		api.GET("/transcode/sessions", s.handleTranscodeSessions)
		api.DELETE("/transcode/sessions/:id", s.handleTranscodeStop)
		api.POST("/transcode/sessions/:id/pause", s.handleTranscodePause)
		api.POST("/transcode/sessions/:id/resume", s.handleTranscodeResume)
		api.GET("/transcode/info/:hash/:fileIdx", s.handleTranscodeInfo)
		api.GET("/transcode/hls/session/:sessionID/main.m3u8", s.handleHLSPlaylist)
		api.GET("/transcode/hls/session/:sessionID/segment/:name", s.handleHLSSegment)
		api.DELETE("/transcode/hls/session/:sessionID", s.handleHLSStop)
		api.GET("/transcode/hls/:hash/:fileIdx/master.m3u8", s.handleHLSMaster)
		api.GET("/transcode/:hash/:fileIdx", s.handleTranscodeStream)

		// Профили транскодирования
		api.GET("/transcode/profiles", s.handleListProfiles)
		api.POST("/transcode/profiles", s.handleCreateProfile)
		api.GET("/transcode/profiles/:id", s.handleGetProfile)
		api.PUT("/transcode/profiles/:id", s.handleUpdateProfile)
		api.DELETE("/transcode/profiles/:id", s.handleDeleteProfile)
		api.POST("/transcode/profiles/:id/default", s.handleSetDefaultProfile)
		api.POST("/transcode/preferred-profile", s.handleSetPreferredProfile)

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
	s.router.Handle("PROPFIND", "/plugins/:id/*path", s.pluginsRouter.Dispatch)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)

	// Слушаем явно, чтобы узнать фактический адрес при порте 0
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.setSourceAddr(ln.Addr())

	s.httpSrv = &http.Server{
		Handler:           s.router,
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       5 * time.Minute,
	}

	s.startStreamJanitor()
	s.startRegistryJanitor()
	s.startHLSJanitor()

	log.Infof("[Web] Server listening on http://%s", ln.Addr())
	if err := s.httpSrv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// startRegistryJanitor периодически убирает просроченные доступы к источникам.
func (s *Server) startRegistryJanitor() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if n := s.sourceRegistry.Cleanup(); n > 0 {
				log.Debugf("[Web] Removed %d expired source lease(s)", n)
			}
		}
	}()
}

// startStreamJanitor периодически убирает из трекера зависшие потоки,
// которые так и не отдали ни байта (соединение открылось, но чтение не пошло).
func (s *Server) startStreamJanitor() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if n := s.streamTracker.Idle(30 * time.Minute); n > 0 {
				log.Warnf("[Web] Dropped %d stale stream(s) from tracker", n)
			}
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("[Web] Stopping HTTP server...")
	s.StopAllHLS()

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
