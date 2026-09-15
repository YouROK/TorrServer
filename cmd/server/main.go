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
	// 1. Парсим аргументы командной строки
	configPath := flag.String("c", "config.yaml", "path to YAML configuration file")
	flag.Parse()

	// 2. Загружаем YAML-конфиг
	cfg, err := config.Load(*configPath)
	if err != nil {
		// Пока логгер не готов, используем fmt для вывода в stderr
		fmt.Fprintln(os.Stderr, "[Config] failed to load config:", err)
		os.Exit(1)
	}

	// 3. Инициализируем наш логгер с настройками из YAML
	if err := log.Init(cfg.Logging.Level, cfg.Logging.File); err != nil {
		fmt.Fprintln(os.Stderr, "[Logger] failed to initialize:", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Infof("[App] === Silo %s started ===", version.Version)
	log.Debugf("[Config] loaded successfully from %s", *configPath)

	// 4. Контекст завершения для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 5. Сборка всех модулей ядра воедино
	application, err := app.New(cfg)
	if err != nil {
		log.Errorf("[App] failed to initialize: %v", err)
		os.Exit(1)
	}

	application.SetShutdownFunc(cancel)

	// 6. Запуск приложения в фоне (блокирует до сигнала ctx.Done)
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
