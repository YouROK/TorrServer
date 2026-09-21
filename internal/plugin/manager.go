package plugin

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"silo/internal/torrent"
	"silo/internal/user"
	"sort"
	"strings"
	"sync"
	"time"

	"silo/internal/bus"
	"silo/internal/database"
	"silo/internal/log"

	bolt "go.etcd.io/bbolt"
	"gopkg.in/yaml.v3"
)

// TorrentAPI — интерфейс для управления движком из плагинов
type TorrentAPI interface {
	SetTrackerPolicy(mode string, trackers []string)
	SetBlocklistText(text string) error
}

type Manager struct {
	mu           sync.RWMutex
	pluginsDir   string
	db           *database.DB
	registrar    WebRegistrar
	torrMgr      *torrent.Manager
	userSvc      *user.Service
	torrentAPI   TorrentAPI
	defaultUIFS  fs.FS
	defaultUIMan *Manifest
	runtimes     map[string]*JSRuntime
	manifests    map[string]*Manifest
	i18n         *I18nRegistry
}

func NewManager(pluginsDir string, db *database.DB, registrar WebRegistrar, torrMgr *torrent.Manager, userSvc *user.Service) (*Manager, error) {
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugins dir: %w", err)
	}

	defaultFS, err := GetDefaultUIFS()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded default UI: %w", err)
	}

	defaultMan, err := LoadManifestFromFS(defaultFS)
	if err != nil {
		return nil, fmt.Errorf("failed to load default UI manifest: %w", err)
	}

	m := &Manager{
		pluginsDir:   pluginsDir,
		db:           db,
		registrar:    registrar,
		torrMgr:      torrMgr,
		userSvc:      userSvc,
		defaultUIFS:  defaultFS,
		defaultUIMan: defaultMan,
		runtimes:     make(map[string]*JSRuntime),
		manifests:    make(map[string]*Manifest),
		i18n:         NewI18nRegistry(),
	}

	return m, nil
}

func (m *Manager) ReloadQueue() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.registrar.Reset()
	m.stopAllRuntimes()

	// 1. СНАЧАЛА загружаем манифесты ВСЕХ плагинов (активных и отключённых)
	//    чтобы ListPlugins мог показать их все
	m.manifests = make(map[string]*Manifest)
	m.i18n.Reset()
	m.manifests[m.defaultUIMan.ID] = m.defaultUIMan

	states, err := m.loadStates()
	if err != nil {
		return err
	}

	// Убеждаемся что встроенный плагин есть в списке состояний
	hasDefault := false
	for i := range states {
		if states[i].ID == m.defaultUIMan.ID {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		states = append(states, PluginState{ID: m.defaultUIMan.ID, Order: 0, Enabled: true})
	}

	sort.Slice(states, func(i, j int) bool {
		return states[i].Order < states[j].Order
	})

	// Загружаем манифесты для всех плагинов в кэш (независимо от флага Enabled)
	for _, state := range states {
		if state.ID == m.defaultUIMan.ID {
			continue // уже добавлен выше
		}

		_, manifest, loadErr := m.loadPluginVFS(state.ID)
		if loadErr != nil {
			log.Errorf("[Plugin] Failed to load manifest for '%s': %v", state.ID, loadErr)
			continue
		}
		m.manifests[state.ID] = manifest
	}

	// 2. ТЕПЕРЬ запускаем только те плагины, у которых Enabled=true
	loadedIDs := make(map[string]bool)
	themeSet := false

	for _, state := range states {
		if !state.Enabled {
			continue
		}

		manifest, ok := m.manifests[state.ID]
		if !ok {
			continue // манифест не удалось загрузить
		}

		if loadedIDs[manifest.ID] {
			log.Warnf("[Plugin] Duplicate plugin ID '%s' skipped", manifest.ID)
			continue
		}
		loadedIDs[manifest.ID] = true

		var pluginVFS fs.FS
		if state.ID == m.defaultUIMan.ID {
			pluginVFS = m.defaultUIFS
		} else {
			vfs, _, loadErr := m.loadPluginVFS(state.ID)
			if loadErr != nil {
				continue
			}
			pluginVFS = vfs
		}

		if manifest.ThemeUI && !themeSet {
			m.registrar.SetActiveTheme(manifest.ID)
			themeSet = true
			log.Infof("[Plugin] Theme UI set to '%s'", manifest.ID)
		}

		m.loadStaticI18n(manifest.ID, pluginVFS)
		m.startPlugin(manifest.ID, manifest, pluginVFS)
		log.Infof("[Plugin] Mounted plugin '%s' (ID: %s, order: %d)", manifest.Name, manifest.ID, state.Order)
	}

	return nil
}

func (m *Manager) startPlugin(pluginID string, manifest *Manifest, vfs fs.FS) {
	if manifest.Entry == "" {
		return
	}

	codeBytes, err := fs.ReadFile(vfs, manifest.Entry)
	if err != nil {
		log.Errorf("[Plugin] Failed to read entry file '%s' for '%s': %v", manifest.Entry, pluginID, err)
		return
	}

	rt := NewJSRuntime(pluginID, manifest, vfs, m.registrar, m.db, m.torrMgr, m.userSvc, m.i18n)
	m.runtimes[pluginID] = rt

	go func(id string, code string, runtime *JSRuntime) {
		log.Debugf("[Plugin:%s] Starting background script...", id)
		if err := runtime.Execute(code); err != nil {
			log.Errorf("[Plugin:%s] Execution error: %v. Purging from memory...", id, err)

			m.mu.Lock()
			runtime.Stop()
			delete(m.runtimes, id)
			m.registrar.RemovePlugin(id)
			m.mu.Unlock()

			bus.Clear(id)

			log.Warnf("[Plugin:%s] Plugin completely purged due to startup error", id)
		}
	}(pluginID, string(codeBytes), rt)
}

func (m *Manager) stopAllRuntimes() {
	for id, rt := range m.runtimes {
		rt.Stop()
		bus.Clear(id)
	}
	m.runtimes = make(map[string]*JSRuntime)
}

func (m *Manager) UnloadAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Info("[Plugin] Unloading all active plugins...")
	m.registrar.Reset()
	m.stopAllRuntimes()
	m.manifests = make(map[string]*Manifest)
	log.Info("[Plugin] All plugins successfully unloaded")
}

func (m *Manager) loadPluginVFS(pluginID string) (fs.FS, *Manifest, error) {
	vfs, err := m.getFS(pluginID)
	if err != nil {
		return nil, nil, err
	}

	manifest, err := LoadManifestFromFS(vfs)
	if err != nil {
		return nil, nil, err
	}

	return vfs, manifest, nil
}

func (m *Manager) getFS(pluginID string) (fs.FS, error) {
	dirPath := filepath.Join(m.pluginsDir, pluginID)
	if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
		return os.DirFS(dirPath), nil
	}

	zipPath := filepath.Join(m.pluginsDir, pluginID+".zip")
	if info, err := os.Stat(zipPath); err == nil && !info.IsDir() {
		zr, err := zip.OpenReader(zipPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open plugin zip: %w", err)
		}
		return zr, nil
	}

	return nil, fmt.Errorf("plugin files not found on disk")
}

func (m *Manager) loadStates() ([]PluginState, error) {
	var list []PluginState
	err := m.db.GetRawConn().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketPlugins)
		return b.ForEach(func(k, v []byte) error {
			var s PluginState
			if err := json.Unmarshal(v, &s); err == nil {
				list = append(list, s)
			}
			return nil
		})
	})
	return list, err
}

func (m *Manager) SetPluginOrder(orderedIDs []string) error {
	err := m.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketPlugins)
		for index, id := range orderedIDs {
			var state PluginState

			// Сохраняем текущее состояние плагина, меняем только порядок
			if data := b.Get([]byte(id)); data != nil {
				_ = json.Unmarshal(data, &state)
			} else {
				state.Enabled = true
				state.CreatedAt = time.Now()
			}

			state.ID = id
			state.Order = index + 1

			data, err := json.Marshal(state)
			if err != nil {
				return err
			}
			if err := b.Put([]byte(id), data); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	return m.ReloadQueue()
}

// ============================================================================
// REST API Методы (Используются веб-слоем)
// ============================================================================

// ListPlugins возвращает ВСЕ плагины (активные и отключённые) в порядке загрузки
func (m *Manager) ListPlugins() []map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states, _ := m.loadStates()
	stateByID := make(map[string]PluginState, len(states))
	for _, st := range states {
		stateByID[st.ID] = st
	}

	// Текущая тема главной страницы
	activeTheme := m.registrar.ActiveThemeID()

	result := make([]map[string]any, 0, len(m.manifests))
	for id, man := range m.manifests {
		enabled := true
		order := 0
		if st, ok := stateByID[id]; ok {
			enabled = st.Enabled
			order = st.Order
		}

		iconURL := ""
		if man.Icon != "" {
			iconURL = "/plugins/" + id + man.Icon
		}

		result = append(result, map[string]any{
			"id":            id,
			"name":          man.Name,
			"version":       man.Version,
			"theme_ui":      man.ThemeUI,
			"current_theme": id == activeTheme, // этот плагин сейчас является темой
			"enabled":       enabled,
			"page":          hasRootRoute(man.Routes),
			"order":         order,
			"builtin":       id == m.defaultUIMan.ID,
			"icon":          iconURL,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i]["order"].(int) < result[j]["order"].(int)
	})
	return result
}

func hasRootRoute(routes []string) bool {
	for _, r := range routes {
		if r == "/" || r == "/index.html" {
			return true
		}
	}
	return false
}

// InstallPlugin устанавливает плагин из ZIP-файла на диск и регистрирует в БД
// InstallPlugin устанавливает плагины из ZIP-файла.
// Поддерживает: (1) плагин в корне, (2) плагин в подпапке, (3) несколько плагинов в разных подпапках.
// Возвращает список ID установленных плагинов.
func (m *Manager) InstallPlugin(zipPath string) ([]string, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer zr.Close()

	// Ищем все вхождения info.yaml на глубине 0 или 1 (корень или подпапка)
	type pluginRoot struct {
		name   string // "" для корня, имя подпапки для подпапки
		prefix string // префикс пути в ZIP для фильтрации файлов
	}

	var roots []pluginRoot

	// Проверка: есть ли info.yaml в корне?
	_, err = zr.Open("info.yaml")
	if err == nil {
		roots = append(roots, pluginRoot{name: "", prefix: ""})
	}

	// Сканируем подпапки верхнего уровня
	dirMap := make(map[string]bool)
	for _, f := range zr.File {
		parts := strings.SplitN(f.Name, "/", 2)
		if len(parts) == 2 && parts[1] != "" && !f.FileInfo().IsDir() {
			// Это файл в подпапке верхнего уровня
			dirMap[parts[0]] = true
		}
	}

	for dir := range dirMap {
		infoPath := dir + "/info.yaml"
		if f, err := zr.Open(infoPath); err == nil {
			f.Close()
			roots = append(roots, pluginRoot{name: dir, prefix: dir + "/"})
		}
	}

	if len(roots) == 0 {
		return nil, fmt.Errorf("manifest not found: neither root nor subdirectory contains info.yaml")
	}

	installed := make([]string, 0, len(roots))

	for _, root := range roots {
		// Читаем манифест из нужного места
		infoPath := "info.yaml"
		if root.prefix != "" {
			infoPath = root.prefix + "info.yaml"
		}

		infoFile, err := zr.Open(infoPath)
		if err != nil {
			return installed, fmt.Errorf("failed to open %s: %w", infoPath, err)
		}
		infoBytes, err := io.ReadAll(infoFile)
		infoFile.Close()
		if err != nil {
			return installed, fmt.Errorf("failed to read %s: %w", infoPath, err)
		}

		var manifest Manifest
		if err := yaml.Unmarshal(infoBytes, &manifest); err != nil {
			return installed, fmt.Errorf("failed to parse %s: %w", infoPath, err)
		}

		if manifest.ID == "" {
			return installed, fmt.Errorf("manifest in %s missing 'id' field", infoPath)
		}

		pluginDir := filepath.Join(m.pluginsDir, manifest.ID)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			return installed, fmt.Errorf("failed to create plugin dir: %w", err)
		}

		// Распаковываем только файлы из нужной подпапки (или все из корня)
		for _, f := range zr.File {
			// Фильтр: если есть префикс — берём только файлы из этой подпапки
			if root.prefix != "" && !strings.HasPrefix(f.Name, root.prefix) {
				continue
			}

			// Относительный путь в плагине
			relName := strings.TrimPrefix(f.Name, root.prefix)
			if relName == "" || relName == "/" {
				continue
			}

			fpath := filepath.Join(pluginDir, relName)

			// Защита от ZipSlip
			cleanPath := filepath.Clean(fpath)
			if !strings.HasPrefix(cleanPath, filepath.Clean(pluginDir)+string(os.PathSeparator)) && cleanPath != filepath.Clean(pluginDir) {
				return installed, fmt.Errorf("illegal file path: %s", fpath)
			}

			if f.FileInfo().IsDir() {
				os.MkdirAll(fpath, os.ModePerm)
				continue
			}

			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return installed, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return installed, err
			}

			rc, err := f.Open()
			if err != nil {
				outFile.Close()
				return installed, err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()

			if err != nil {
				return installed, err
			}
		}

		// Добавляем в базу с максимальным порядком
		err = m.db.GetRawConn().Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(database.BucketPlugins)

			maxOrder := 0
			_ = b.ForEach(func(k, v []byte) error {
				var s PluginState
				if err := json.Unmarshal(v, &s); err == nil {
					if s.Order > maxOrder {
						maxOrder = s.Order
					}
				}
				return nil
			})

			state := PluginState{
				ID:        manifest.ID,
				Order:     maxOrder + 1,
				Enabled:   true,
				CreatedAt: time.Now(),
			}
			data, _ := json.Marshal(state)
			return b.Put([]byte(manifest.ID), data)
		})

		if err != nil {
			return installed, fmt.Errorf("failed to save plugin state to db: %w", err)
		}

		log.Infof("[Plugin] Installed plugin '%s' from %s", manifest.ID, zipPath)
		installed = append(installed, manifest.ID)
	}

	if err := m.ReloadQueue(); err != nil {
		log.Errorf("[Plugin] Failed to reload queue after install: %v", err)
	}

	return installed, nil
}

// UninstallPlugin удаляет плагин с диска и из базы данных
func (m *Manager) UninstallPlugin(pluginID string) error {
	if pluginID == "default_ui" {
		return fmt.Errorf("cannot uninstall built-in default UI")
	}
	if pluginID == "admin_ui" {
		return fmt.Errorf("cannot uninstall built-in admin UI")
	}

	err := m.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		pluginsBucket := tx.Bucket(database.BucketPlugins)
		if err := pluginsBucket.Delete([]byte(pluginID)); err != nil {
			return err
		}

		dataBucket := tx.Bucket(database.BucketPluginData)
		if dataBucket == nil {
			return nil
		}

		prefix := []byte(pluginID + ":")
		c := dataBucket.Cursor()

		var keysToDelete [][]byte
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			keyCopy := make([]byte, len(k))
			copy(keyCopy, k)
			keysToDelete = append(keysToDelete, keyCopy)
		}

		for _, key := range keysToDelete {
			if err := dataBucket.Delete(key); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete plugin from db: %w", err)
	}

	pluginDir := filepath.Join(m.pluginsDir, pluginID)
	if err := os.RemoveAll(pluginDir); err != nil {
		log.Warnf("[Plugin] Failed to remove plugin directory: %v", err)
	}

	zipPath := filepath.Join(m.pluginsDir, pluginID+".zip")
	_ = os.Remove(zipPath)

	log.Infof("[Plugin] Uninstalled plugin '%s'", pluginID)
	return m.ReloadQueue()
}

// SetPluginEnabled включает или выключает плагин с перезагрузкой очереди
func (m *Manager) SetPluginEnabled(pluginID string, enabled bool) error {
	err := m.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketPlugins)
		data := b.Get([]byte(pluginID))

		var state PluginState
		if data != nil {
			if err := json.Unmarshal(data, &state); err != nil {
				return err
			}
		} else {
			// Плагин еще не в базе (например, встроенный) — создаем запись
			state.Enabled = true
			state.CreatedAt = time.Now()
		}

		state.ID = pluginID
		state.Enabled = enabled

		out, err := json.Marshal(state)
		if err != nil {
			return err
		}
		return b.Put([]byte(pluginID), out)
	})

	if err != nil {
		return err
	}

	log.Infof("[Plugin] Plugin '%s' enabled set to %v", pluginID, enabled)
	return m.ReloadQueue()
}

// GetPluginInfo возвращает манифест плагина для карточки информации
func (m *Manager) GetPluginInfo(pluginID string) (*Manifest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	man, ok := m.manifests[pluginID]
	if !ok {
		return nil, fmt.Errorf("plugin not found")
	}
	return man, nil
}

// InstallFromURL скачивает плагин по URL и устанавливает его.
// Поддерживает любые прямые ссылки на .zip с автоматическим следованием редиректам.
// Защита: SSRF (запрет на приватные IP), ограничение размера 50MB, таймаут 60с.
func (m *Manager) InstallFromURL(rawURL string) ([]string, error) {
	// 1. Парсим и валидируем URL
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("only http and https schemes are allowed")
	}

	// 2. SSRF-защита: резолвим хост и проверяем, что IP не приватный
	host := parsed.Hostname()
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host %s: %w", host, err)
	}
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return nil, fmt.Errorf("private IP addresses are not allowed: %s", ip.String())
		}
	}

	// 3. Скачиваем с кастомным клиентом: таймаут + автоматические редиректы (до 10)
	client := &http.Client{
		Timeout: 60 * time.Second,
		// CheckRedirect оставляем дефолтным: http.Client сам следует до 10 редиректам
	}

	log.Infof("[Plugin] Downloading plugin from %s", rawURL)

	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// 4. Создаём временный файл
	tmpFile, err := os.CreateTemp("", "silo-plugin-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// 5. Скачиваем с лимитом 50 МБ
	const maxPluginSize = 50 * 1024 * 1024
	written, err := io.Copy(tmpFile, io.LimitReader(resp.Body, maxPluginSize+1))
	tmpFile.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to write plugin file: %w", err)
	}
	if written > maxPluginSize {
		return nil, fmt.Errorf("plugin file too large (max %d MB)", maxPluginSize/1024/1024)
	}

	log.Infof("[Plugin] Downloaded %d bytes, installing...", written)

	// 6. Делегируем установку существующему методу
	return m.InstallPlugin(tmpPath)
}

// isPrivateIP проверяет, принадлежит ли IP к приватным диапазонам
func isPrivateIP(ip net.IP) bool {
	privateRanges := []struct {
		network string
	}{
		{"10.0.0.0/8"},
		{"172.16.0.0/12"},
		{"192.168.0.0/16"},
		{"127.0.0.0/8"},
		{"169.254.0.0/16"},
		{"::1/128"},
		{"fc00::/7"},
		{"fe80::/10"},
	}

	for _, r := range privateRanges {
		_, network, err := net.ParseCIDR(r.network)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	// Link-local и unspecified
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}

	return false
}

// I18n возвращает реестр переводов для веб-слоя
func (m *Manager) I18n() *I18nRegistry {
	return m.i18n
}

// loadStaticI18n читает файлы плагина i18n/<lang>.json в его слой перевода.
// Позволяет пакам переводов существовать без service.js.
func (m *Manager) loadStaticI18n(pluginID string, vfs fs.FS) {
	entries, err := fs.ReadDir(vfs, "i18n")
	if err != nil {
		return // папки i18n нет — это нормально
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		lang := strings.TrimSuffix(e.Name(), ".json")
		data, err := fs.ReadFile(vfs, "i18n/"+e.Name())
		if err != nil {
			log.Warnf("[Plugin:%s] failed to read i18n file %s: %v", pluginID, e.Name(), err)
			continue
		}

		m.i18n.AddStatic(pluginID, lang, data)
	}
}

// MenuFor собирает пункты меню активных плагинов для указанного ранга
func (m *Manager) MenuFor(rank int) []map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states, _ := m.loadStates()
	order := make(map[string]int, len(states))
	for i, st := range states {
		order[st.ID] = i
	}

	// ID плагинов в порядке загрузки
	ids := make([]string, 0, len(m.manifests))
	for id := range m.manifests {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		oi, oki := order[ids[i]]
		oj, okj := order[ids[j]]
		if !oki {
			oi = -1
		}
		if !okj {
			oj = -1
		}
		return oi < oj
	})

	result := make([]map[string]any, 0)
	for _, id := range ids {
		man := m.manifests[id]

		// Меню могут давать только реально запущенные плагины
		if _, active := m.runtimes[id]; !active {
			continue
		}

		for _, e := range man.Menu {
			if e.Rank > rank {
				continue
			}

			entry := map[string]any{
				"plugin_id": id,
				"title":     e.Title,
				"title_key": e.TitleKey,
				"href":      "/plugins/" + id + e.Route,
				"rank":      e.Rank,
			}
			if e.Icon != "" {
				entry["icon"] = "/plugins/" + id + e.Icon
			}
			result = append(result, entry)
		}
	}

	return result
}
