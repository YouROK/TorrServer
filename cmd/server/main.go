package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"silo/internal/app"
	"silo/internal/config"
	"silo/internal/log"
	"silo/internal/version"
)

func main() {
	configPath := flag.String("c", "config.yaml", "path to YAML configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[Config] failed to load config:", err)
		os.Exit(1)
	}

	if err := log.Init(cfg.Logging.Level, cfg.Logging.File); err != nil {
		fmt.Fprintln(os.Stderr, "[Logger] failed to initialize:", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Infof("[App] === TorrServer %s, engine: %s started ===", version.Version, version.GetTorrentVersion())
	log.Debugf("[Config] loaded successfully from %s", *configPath)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	application, err := app.New(cfg)
	if err != nil {
		log.Errorf("[App] failed to initialize: %v", err)
		os.Exit(1)
	}

	application.SetShutdownFunc(cancel)

	go func() {
		if err := application.Start(ctx); err != nil {
			log.Errorf("[App] runtime error: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Warn("[App] shutdown signal received")
	application.Stop()
}
