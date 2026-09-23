package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"silo/internal/log"
)

const (
	plugincatalogURL     = "https://releases.yourok.ru/torr/ts_plugins.json"
	plugincatalogTTL     = 15 * time.Minute
	plugincatalogRetry   = 1 * time.Minute
	plugincatalogTimeout = 15 * time.Second
	catalogMaxZipSize    = 50 * 1024 * 1024
)

type catalogEntry struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	Author         string `json:"author,omitempty"`
	Description    string `json:"description,omitempty"`
	Icon           string `json:"icon,omitempty"`
	DownloadURL    string `json:"download_url"`
	Size           int64  `json:"size,omitempty"`
	SHA256         string `json:"sha256"`
	Homepage       string `json:"homepage,omitempty"`
	MinSiloVersion string `json:"min_silo_version,omitempty"`
}

type catalogFile struct {
	Version   int             `json:"version"`
	UpdatedAt string          `json:"updated_at"`
	Plugins   []*catalogEntry `json:"plugins"`
}

type catalogResponse struct {
	Stale     bool            `json:"stale"`
	UpdatedAt string          `json:"updated_at"`
	Plugins   []*catalogEntry `json:"plugins"`
}

var sha256Re = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
var siloVerRe = regexp.MustCompile(`(?i)^silo\.(\d+)$`)

type PluginCatalog struct {
	mu         sync.Mutex
	cached     *catalogFile
	cachedAt   time.Time
	lastFailed time.Time
}

var globalCatalog = &PluginCatalog{}

func GetPluginCatalog() *PluginCatalog { return globalCatalog }

// Fetch возвращает каталог с TTL-кэшем.
// Если CDN недоступен, но есть устаревший кэш — отдаём его с флагом stale.
// Повторные попытки после ошибки — не чаще, чем раз в plugincatalogRetry.
func (c *PluginCatalog) Fetch() (*catalogResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cached != nil && time.Since(c.cachedAt) < plugincatalogTTL {
		return c.response(false), nil
	}

	if !c.lastFailed.IsZero() && time.Since(c.lastFailed) < plugincatalogRetry {
		if c.cached != nil {
			return c.response(true), nil
		}
		wait := (plugincatalogRetry - time.Since(c.lastFailed)).Round(time.Second)
		return nil, fmt.Errorf("catalog unavailable, retry in %s", wait)
	}

	cf, err := c.fetchRemote()
	if err != nil {
		c.lastFailed = time.Now()
		if c.cached != nil {
			log.Warnf("[Catalog] fetch failed, serving stale: %v", err)
			return c.response(true), nil
		}
		return nil, err
	}

	c.cached = cf
	c.cachedAt = time.Now()
	c.lastFailed = time.Time{}
	return c.response(false), nil
}

func (c *PluginCatalog) fetchRemote() (*catalogFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), plugincatalogTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, plugincatalogURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Silo/Catalog")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("catalog fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog status %d", resp.StatusCode)
	}

	var cf catalogFile
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4*1024*1024)).Decode(&cf); err != nil {
		return nil, fmt.Errorf("catalog parse: %w", err)
	}

	valid := cf.Plugins[:0]
	for _, e := range cf.Plugins {
		if e == nil || e.ID == "" || e.Version == "" || e.DownloadURL == "" {
			log.Warnf("[Catalog] entry skipped: missing required fields")
			continue
		}
		if !sha256Re.MatchString(e.SHA256) {
			log.Warnf("[Catalog] entry '%s' skipped: invalid sha256", e.ID)
			continue
		}
		if e.MinSiloVersion != "" {
			if _, ok := parseSiloVersion(e.MinSiloVersion); !ok {
				log.Warnf("[Catalog] entry '%s' skipped: invalid min_silo_version '%s'", e.ID, e.MinSiloVersion)
				continue
			}
		}
		valid = append(valid, e)
	}
	cf.Plugins = valid

	log.Infof("[Catalog] loaded %d plugins", len(cf.Plugins))
	return &cf, nil
}

func (c *PluginCatalog) response(stale bool) *catalogResponse {
	return &catalogResponse{
		Stale:     stale,
		UpdatedAt: c.cached.UpdatedAt,
		Plugins:   c.cached.Plugins,
	}
}

func (c *PluginCatalog) Get(id string) (*catalogEntry, error) {
	resp, err := c.Fetch()
	if err != nil {
		return nil, err
	}
	for _, e := range resp.Plugins {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, fmt.Errorf("plugin '%s' not found in catalog", id)
}

// Download скачивает ZIP, проверяет sha256 и возвращает путь к temp-файлу.
func (c *PluginCatalog) Download(entry *catalogEntry) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, entry.DownloadURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Silo/Catalog")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download status %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "silo-catalog-*.zip")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()

	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(resp.Body, catalogMaxZipSize+1))
	tmp.Close()
	if err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	if n > catalogMaxZipSize {
		os.Remove(tmpPath)
		return "", fmt.Errorf("zip too large (%d bytes)", n)
	}

	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, entry.SHA256) {
		os.Remove(tmpPath)
		return "", fmt.Errorf("sha256 mismatch")
	}

	return tmpPath, nil
}

// parseSiloVersion извлекает X из "silo.X".
func parseSiloVersion(s string) (int, bool) {
	m := siloVerRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return n, true
}

// IsCompatible проверяет, совместима ли запись каталога с текущей версией Silo.
func IsCompatible(entry *catalogEntry, currentVersion string) bool {
	if entry.MinSiloVersion == "" {
		return true
	}
	need, ok := parseSiloVersion(entry.MinSiloVersion)
	if !ok {
		return false
	}
	have, ok := parseSiloVersion(currentVersion)
	if !ok {
		return true
	}
	return have >= need
}
