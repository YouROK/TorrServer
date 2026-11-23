package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"silo/internal/app"
	"silo/internal/config"
	"silo/internal/log"
	"silo/internal/version"
)

func main() {
	configPath := flag.String("c", "config.yaml", "path to YAML configuration file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

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
	log.Debugf("[Config] Loaded successfully from %s", *configPath)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	application, err := app.New(cfg)
	if err != nil {
		log.Errorf("[App] Failed to initialize: %v", err)
		os.Exit(1)
	}

	application.SetShutdownFunc(cancel)

	go func() {
		if err := application.Start(ctx); err != nil {
			log.Errorf("[App] Runtime error: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Warn("[App] Shutdown signal received")
	application.Stop()
}

// printVersion выводит версию сборки, движка и платформу одной строкой.
func printVersion() {
	release := version.Version
	if release == "" {
		release = "dev"
	}
	engine := version.GetTorrentVersion()
	if engine == "" {
		engine = "unknown"
	}
	fmt.Printf("%s (torrent engine %s, %s %s/%s)\n",
		release, engine, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
