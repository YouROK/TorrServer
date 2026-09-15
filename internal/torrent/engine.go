package torrent

import (
	"fmt"
	"strings"
	"sync"

	"silo/internal/log"
	"silo/internal/torrent/storage/torrstor"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/iplist"
	"github.com/anacrolix/torrent/metainfo"
)

type Engine struct {
	cfg       *Config
	client    *torrent.Client
	storage   *torrstor.Storage
	blocklist *DynamicBlocklist
	sessions  map[metainfo.Hash]*Session
	mu        sync.RWMutex
}

func NewEngine(cfg *Config) (*Engine, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 1. Создаем изолированное хранилище памяти torrstor
	stor := torrstor.NewStorage(cfg.Storage)

	dynBlocklist := NewDynamicBlocklist(nil)

	// 2. Настраиваем anacrolix клиент
	clientCfg := torrent.NewDefaultClientConfig()
	clientCfg.DefaultStorage = stor
	clientCfg.ListenPort = cfg.ListenPort
	clientCfg.EstablishedConnsPerTorrent = cfg.Storage.ConnectionsLimit
	clientCfg.TotalHalfOpenConns = 500
	clientCfg.IPBlocklist = dynBlocklist
	clientCfg.NoDHT = cfg.DisableDHT
	clientCfg.DisablePEX = cfg.DisablePEX
	clientCfg.NoDefaultPortForwarding = cfg.DisableUPNP
	clientCfg.DisableUTP = cfg.DisableUTP
	clientCfg.DisableTCP = cfg.DisableTCP
	clientCfg.DisableIPv6 = !cfg.EnableIPv6

	// Лимитеры скорости
	if cfg.DownloadRateKB > 0 {
		clientCfg.DownloadRateLimiter = NewRateLimiter(cfg.DownloadRateKB)
	}
	if cfg.UploadRateKB > 0 {
		clientCfg.UploadRateLimiter = NewRateLimiter(cfg.UploadRateKB)
	}

	// ID клиента
	peerID := GeneratePeerID("-qB4390-")
	clientCfg.PeerID = peerID
	clientCfg.Bep20 = "-qB4390-"
	clientCfg.HTTPUserAgent = "qBittorrent/4.3.9"
	clientCfg.ExtendedHandshakeClientVersion = "qBittorrent/4.3.9"

	client, err := torrent.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize bittorrent client: %w", err)
	}

	var listenAddrs []string
	for _, a := range client.ListenAddrs() {
		listenAddrs = append(listenAddrs, a.String())
	}
	log.Infof("[Torrent Engine] Initialized client listening on %v", listenAddrs)

	return &Engine{
		cfg:       cfg,
		client:    client,
		storage:   stor,
		blocklist: dynBlocklist,
		sessions:  make(map[metainfo.Hash]*Session),
	}, nil
}

// Restart полностью перезапускает anacrolix-клиент и хранилище с новой конфигурацией
func (e *Engine) Restart(cfg *Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	log.Info("[Torrent Engine] Restarting engine with new configuration...")

	// 1. Закрываем все текущие сессии в RAM
	for _, s := range e.sessions {
		s.Close()
	}
	e.sessions = make(map[metainfo.Hash]*Session)

	// 2. Закрываем хранилище и старый клиент
	if e.storage != nil {
		_ = e.storage.Close()
	}
	if e.client != nil {
		e.client.Close()
	}

	// Принудительно возвращаем память операционной системе
	torrstor.FreeOSMemGC()

	// 3. Создаем новое хранилище torrstor с новым размером кэша
	if cfg == nil {
		cfg = DefaultConfig()
	}
	e.cfg = cfg

	stor := torrstor.NewStorage(cfg.Storage)
	e.storage = stor

	// 4. Настраиваем и поднимаем новый чистый anacrolix клиент
	clientCfg := torrent.NewDefaultClientConfig()
	clientCfg.DefaultStorage = stor
	clientCfg.ListenPort = cfg.ListenPort
	clientCfg.EstablishedConnsPerTorrent = cfg.Storage.ConnectionsLimit
	clientCfg.TotalHalfOpenConns = 500

	clientCfg.IPBlocklist = e.blocklist // сохраняем текущий блоклист

	clientCfg.NoDHT = cfg.DisableDHT
	clientCfg.DisablePEX = cfg.DisablePEX
	clientCfg.NoDefaultPortForwarding = cfg.DisableUPNP
	clientCfg.DisableUTP = cfg.DisableUTP
	clientCfg.DisableTCP = cfg.DisableTCP
	clientCfg.DisableIPv6 = !cfg.EnableIPv6

	if cfg.DownloadRateKB > 0 {
		clientCfg.DownloadRateLimiter = NewRateLimiter(cfg.DownloadRateKB)
	}
	if cfg.UploadRateKB > 0 {
		clientCfg.UploadRateLimiter = NewRateLimiter(cfg.UploadRateKB)
	}

	peerID := GeneratePeerID("-qB4390-")
	clientCfg.PeerID = peerID
	clientCfg.Bep20 = "-qB4390-"
	clientCfg.HTTPUserAgent = "qBittorrent/4.3.9"
	clientCfg.ExtendedHandshakeClientVersion = "qBittorrent/4.3.9"

	client, err := torrent.NewClient(clientCfg)
	if err != nil {
		return fmt.Errorf("failed to restart bittorrent client: %w", err)
	}
	e.client = client

	var listenAddrs []string
	for _, a := range client.ListenAddrs() {
		listenAddrs = append(listenAddrs, a.String())
	}
	log.Infof("[Torrent Engine] Restarted successfully, listening on %v", listenAddrs)
	return nil
}

// Start запускает раздачу в оперативной памяти с теми трекерами, которые были переданы в spec
func (e *Engine) Start(spec *torrent.TorrentSpec) (*Session, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if sess, exists := e.sessions[spec.InfoHash]; exists {
		sess.Touch()
		return sess, nil
	}

	t, _, err := e.client.AddTorrentSpec(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to add torrent spec: %w", err)
	}

	sess := newSession(t, spec, e.storage)
	e.sessions[spec.InfoHash] = sess

	log.Infof("[Torrent Engine] Started active session in RAM: %s", spec.InfoHash.HexString())
	return sess, nil
}

// Get возвращает активную сессию из памяти
func (e *Engine) Get(hash metainfo.Hash) (*Session, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	sess, ok := e.sessions[hash]
	return sess, ok
}

// Stop останавливает раздачу и освобождает память RAM
func (e *Engine) Stop(hash metainfo.Hash) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if sess, exists := e.sessions[hash]; exists {
		sess.Close()
		e.storage.CloseHash(hash)
		delete(e.sessions, hash)
		log.Infof("[Torrent Engine] Stopped session and cleared RAM: %s", hash.HexString())
	}
}

// List возвращает все активные сессии в памяти
func (e *Engine) List() []*Session {
	e.mu.RLock()
	defer e.mu.RUnlock()

	res := make([]*Session, 0, len(e.sessions))
	for _, s := range e.sessions {
		res = append(res, s)
	}
	return res
}

// SetBlocklist применяет готовый скомпилированный список диапазонов
func (e *Engine) SetBlocklist(ranger iplist.Ranger) {
	if e.blocklist == nil {
		e.blocklist = NewDynamicBlocklist(ranger)
		return
	}
	e.blocklist.Set(ranger)
}

// SetBlocklistText принимает текст P2P-блоклиста от плагина и применяет его на лету
func (e *Engine) SetBlocklistText(p2pText string) error {
	ranger, err := ParseBlocklistP2P(strings.NewReader(p2pText))
	if err != nil {
		return err
	}
	e.SetBlocklist(ranger)
	return nil
}

// Close мягко закрывает весь движок
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, s := range e.sessions {
		s.Close()
	}
	e.sessions = make(map[metainfo.Hash]*Session)
	_ = e.storage.Close()
	e.client.Close()

	log.Info("[Torrent Engine] Closed successfully")
	return nil
}

// UpdateConfig сохраняет новые настройки в базу данных и перезапускает движок
func (m *Manager) UpdateConfig(cfg *Config) error {
	// 1. Сохраняем в bbolt базу данных
	if err := m.store.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save torrent config to DB: %w", err)
	}

	// 2. Перезапускаем чистый движок
	if err := m.engine.Restart(cfg); err != nil {
		return err
	}

	// 3. Оповещаем систему и плагины об изменении настроек
	m.bus.Emit("system:config:updated", cfg)
	return nil
}
