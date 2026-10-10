package ffmpeg

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultMinVersion     = "6.0"
	DefaultMaxSessions    = 3
	DefaultProbeTimeout   = 30 * time.Second
	DefaultStartTimeout   = 15 * time.Second
	DefaultStopTimeout    = 5 * time.Second
	DefaultProgressPeriod = time.Second
	DefaultCacheDirName   = "transcode"
	// DefaultCacheSegments - сколько сегментов HLS остаётся на диске
	DefaultCacheSegments = 6
	DefaultKeepAhead     = 4
	// DefaultHWDevice - стандартный узел рендеринга VAAPI
	DefaultHWDevice = "/dev/dri/renderD128"
)

// Config описывает настройки транскодинга из файла конфигурации.
type Config struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	CacheDir      string `yaml:"cache_dir" json:"cache_dir"`       // Пусто - используется <data_dir>/transcode
	MaxSessions   int    `yaml:"max_sessions" json:"max_sessions"` // 0 - без ограничения
	ProbeTimeout  int    `yaml:"probe_timeout_sec" json:"probe_timeout_sec"`
	Threads       int    `yaml:"threads" json:"threads"`               // 0 - авто
	HWAccel       string `yaml:"hwaccel" json:"hwaccel"`               // none, vaapi, nvenc
	KeepStderr    bool   `yaml:"keep_stderr" json:"keep_stderr"`       // Писать полный stderr ffmpeg в лог
	CacheSegments int    `yaml:"cache_segments" json:"cache_segments"` // Сколько сегментов HLS держать на диске
	HWDevice      string `yaml:"hw_device" json:"hw_device"`           // Путь к устройству ускорения, пусто - значение по умолчанию
	KeepAhead     int    `yaml:"keep_ahead" json:"keep_ahead"`         // Сколько сегментов кодировать наперёд
	Throttle      *bool  `yaml:"throttle" json:"throttle"`             // Придерживать ffmpeg при отставании плеера
}

// HWDevicePath возвращает путь к устройству аппаратного ускорения.
func (c *Config) HWDevicePath() string {
	if path := strings.TrimSpace(c.HWDevice); path != "" {
		return path
	}
	return DefaultHWDevice
}

// ThrottleEnabled сообщает, нужно ли придерживать кодирование.
func (c *Config) ThrottleEnabled() bool {
	return c.Throttle == nil || *c.Throttle
}

// DefaultConfig возвращает настройки транскодинга по умолчанию.
func DefaultConfig() Config {
	return Config{
		Enabled:      true,
		MaxSessions:  DefaultMaxSessions,
		ProbeTimeout: int(DefaultProbeTimeout / time.Second),
		HWAccel:      "none",
		KeepAhead:    DefaultKeepAhead,
	}
}

// Normalize приводит настройки к допустимым значениям.
func (c *Config) Normalize() {
	if c.MaxSessions < 0 {
		c.MaxSessions = 0
	}
	if c.ProbeTimeout <= 0 {
		c.ProbeTimeout = int(DefaultProbeTimeout / time.Second)
	}
	if c.Threads < 0 {
		c.Threads = 0
	}
	switch strings.ToLower(strings.TrimSpace(c.HWAccel)) {
	case "vaapi", "nvenc", "qsv":
		c.HWAccel = strings.ToLower(strings.TrimSpace(c.HWAccel))
	case "auto":
		// Ускорение подбирается по возможностям сборки и наличию устройств
		c.HWAccel = "auto"
	default:
		c.HWAccel = "none"
	}
	if c.CacheSegments < 0 {
		c.CacheSegments = 0
	}
	if c.KeepAhead <= 0 {
		c.KeepAhead = DefaultKeepAhead
	}
	c.CacheDir = strings.TrimSpace(c.CacheDir)
}

// ResolveCacheDir возвращает каталог для файлов транскодирования.
func (c *Config) ResolveCacheDir(dataDir string) string {
	if c.CacheDir != "" {
		return c.CacheDir
	}
	return filepath.Join(dataDir, DefaultCacheDirName)
}

func (c *Config) probeTimeout() time.Duration {
	return time.Duration(c.ProbeTimeout) * time.Second
}

// validate проверяет настройки перед запуском.
func (c *Config) validate() error {
	if c.MaxSessions < 0 {
		return fmt.Errorf("max_sessions cannot be negative")
	}
	if c.ProbeTimeout < 1 {
		return fmt.Errorf("probe_timeout_sec must be positive")
	}
	return nil
}
