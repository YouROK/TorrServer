package app

import (
	"context"
	"fmt"
	"os"
	"silo/internal/torrfs"
	"time"

	"silo/internal/bus"
	"silo/internal/config"
	"silo/internal/database"
	"silo/internal/log"
	"silo/internal/plugin"
	"silo/internal/torrent"
	"silo/internal/user"
	"silo/internal/web"
)

type torrentAPIAdapter struct {
	mgr    *torrent.Manager
	engine *torrent.Engine
}

func (a *torrentAPIAdapter) SetTrackerPolicy(mode string, trackers []string) {
	a.mgr.SetTrackerPolicy(torrent.TrackerMode(mode), trackers)
}

func (a *torrentAPIAdapter) SetBlocklistText(text string) error {
	return a.engine.SetBlocklistText(text)
}

type App struct {
	cfg       *config.Config
	db        *database.DB
	bus       *bus.Client
	userSvc   *user.Service
	engine    *torrent.Engine
	pluginMgr *plugin.Manager
	webSrv    *web.Server
}

func New(cfg *config.Config) (*App, error) {
	if err := os.MkdirAll(cfg.Storage.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}
	if err := os.MkdirAll(cfg.Plugins.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugins dir: %w", err)
	}

	db, err := database.Open(cfg.DBPath())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	appBus := bus.Get("app_core")

	userStore := user.NewStore(db)
	userSvc := user.NewService(userStore, cfg)

	torrentStore := torrent.NewStore(db)

	torrentCfg, err := torrentStore.GetConfig()
	if err != nil {
		log.Warnf("[App] Failed to load torrent config from DB, using defaults: %v", err)
		torrentCfg = torrent.DefaultConfig()
	}

	engine, err := torrent.NewEngine(torrentCfg)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize torrent engine: %w", err)
	}

	torrentMgr := torrent.NewManager(engine, torrentStore, userSvc)

	webServer := web.NewServer(cfg, userSvc, torrentMgr)

	torrFS := torrfs.New(userSvc, torrentStore, torrentMgr)

	pluginMgr, err := plugin.NewManager(cfg.Plugins.Dir, db, webServer.GetPluginsRouter(), torrentMgr, userSvc, torrFS)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize plugin manager: %w", err)
	}

	webServer.SetPluginManager(pluginMgr)

	if err := pluginMgr.ReloadQueue(); err != nil {
		log.Errorf("[Plugin] Error loading initial plugin queue: %v", err)
	}

	return &App{
		cfg:       cfg,
		db:        db,
		bus:       appBus,
		userSvc:   userSvc,
		engine:    engine,
		pluginMgr: pluginMgr,
		webSrv:    webServer,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	log.Info("[App] Silo core started successfully")
	log.Infof("[App] Database: %s", a.cfg.DBPath())
	log.Infof("[App] Plugins dir: %s", a.cfg.Plugins.Dir)

	if a.userSvc.IsAuthRequired() {
		log.Info("[Auth] Access mode: PASSWORD REQUIRED")
	} else {
		log.Warn("[Auth] Access mode: OPEN (Running as Owner without password)")
	}

	go func() {
		if err := a.webSrv.Start(); err != nil {
			log.Errorf("[Web] Server runtime error: %v", err)
		}
	}()

	a.bus.Emit("system:boot", map[string]any{
		"host": a.cfg.Server.Host,
		"port": a.cfg.Server.Port,
	})

	<-ctx.Done()
	return nil
}

func (a *App) Stop() {
	log.Info("[App] Shutting down services...")

	a.bus.Emit("system:shutdown", map[string]any{
		"reason": "os_signal",
	})

	time.Sleep(500 * time.Millisecond)

	a.pluginMgr.UnloadAll()
	a.bus.UnsubscribeAll()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.webSrv.Stop(shutdownCtx); err != nil {
		log.Errorf("[Web] Error stopping web server: %v", err)
	}

	if err := a.engine.Close(); err != nil {
		log.Errorf("[App] Error closing torrent engine: %v", err)
	}

	if err := a.db.Close(); err != nil {
		log.Errorf("[App] Error closing database: %v", err)
	} else {
		log.Info("[App] Database closed successfully")
	}

	log.Info("[App] Silo server stopped cleanly")
}

func (a *App) SetShutdownFunc(fn func()) {
	a.webSrv.SetShutdownFunc(fn)
}
