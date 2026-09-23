package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server" json:"server"`
	Storage StorageConfig `yaml:"storage" json:"storage"`
	Plugins PluginsConfig `yaml:"plugins" json:"plugins"`
	Torrent TorrentConfig `yaml:"torrent" json:"torrent"`
	Logging LoggingConfig `yaml:"logging" json:"logging"`
	Auth    AuthConfig    `yaml:"auth" json:"auth"`
}

type ServerConfig struct {
	Host string `yaml:"host" json:"host"` // 0.0.0.0
	Port int    `yaml:"port" json:"port"` // 8090
}

type StorageConfig struct {
	DataDir string `yaml:"data_dir" json:"data_dir"` // Базовая папка для данных программы (./data)
	DBName  string `yaml:"db_name" json:"db_name"`   // Имя файла БД (torrserver.db)
}

type PluginsConfig struct {
	Dir string `yaml:"dir" json:"dir"`
}

type TorrentConfig struct {
	CacheSizeMB    int64 `yaml:"cache_size_mb" json:"cache_size_mb"`       // Размер RAM-кэша
	PeersLimit     int   `yaml:"peers_limit" json:"peers_limit"`           // Макс. кол-во пиров
	DownloadRateKB int   `yaml:"download_rate_kb" json:"download_rate_kb"` // 0 - без ограничений
	UploadRateKB   int   `yaml:"upload_rate_kb" json:"upload_rate_kb"`     // 0 - без ограничений
}

type LoggingConfig struct {
	Level string `yaml:"level" json:"level"` // "debug", "info", "warn", "error"
	File  string `yaml:"file" json:"file"`   // Если пусто, пишем только в консоль
}

type AuthConfig struct {
	// Если пароль пустой ("") — авторизация отключена, сервер работает в открытом режиме!
	OwnerPassword string `yaml:"owner_password" json:"owner_password"`
}

// DBPath возвращает полный путь к файлу базы данных
func (c *Config) DBPath() string {
	return filepath.Join(c.Storage.DataDir, c.Storage.DBName)
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8090,
		},
		Storage: StorageConfig{
			DataDir: "data",
			DBName:  "torrserver.db",
		},
		Plugins: PluginsConfig{
			Dir: "plugins",
		},
		Torrent: TorrentConfig{
			CacheSizeMB:    200,
			PeersLimit:     200,
			DownloadRateKB: 0,
			UploadRateKB:   0,
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "",
		},
		Auth: AuthConfig{
			OwnerPassword: "",
		},
	}
}

// Load загружает конфиг из YAML или создает дефолтный, если файла нет
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(configPath)
	if errors.Is(err, os.ErrNotExist) {
		// Файла нет — создаем дефолтный с красивым форматированием
		if err := Save(configPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Парсим существующий файл
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return cfg, nil
}

// Save сохраняет конфиг в файл
func Save(configPath string, cfg *Config) error {
	dir := filepath.Dir(configPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
