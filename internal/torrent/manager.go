package torrent

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"silo/internal/torrshash"
	"strconv"
	"strings"
	"sync"
	"time"

	"silo/internal/bus"
	"silo/internal/log"
	"silo/internal/torrent/storage/torrstor"
	"silo/internal/user"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// TrackerMode определяет, как обрабатывать трекеры при старте раздачи
type TrackerMode string

const (
	TrackerModeNone    TrackerMode = "none"    // Не трогать, запускать как есть
	TrackerModeAppend  TrackerMode = "append"  // Подмешать трекеры плагина к существующим
	TrackerModeReplace TrackerMode = "replace" // Стереть исходные трекеры и поставить трекеры плагина
	TrackerModeRemove  TrackerMode = "remove"  // Стереть вообще все трекеры (работать только по DHT)
)

type TrackerPolicy struct {
	Mode     TrackerMode `json:"mode"`
	Trackers []string    `json:"trackers"`
}

type Manager struct {
	engine     *Engine
	store      *Store
	userSvc    *user.Service
	bus        *bus.Client
	mu         sync.RWMutex
	stopWorker chan struct{}

	// Политика трекеров, которой управляют плагины
	trackerPolicy TrackerPolicy

	// Таймаут неактивности перед выгрузкой из RAM
	inactivityTimeout time.Duration
}

func NewManager(engine *Engine, store *Store, userSvc *user.Service) *Manager {
	m := &Manager{
		engine:            engine,
		store:             store,
		userSvc:           userSvc,
		bus:               bus.Get("torrent_manager"),
		stopWorker:        make(chan struct{}),
		inactivityTimeout: 60 * time.Second,
		trackerPolicy: TrackerPolicy{
			Mode: TrackerModeNone,
		},
	}

	// 1. Слушаем шину: если раздача больше никому не принадлежит, удаляем из БД и RAM
	m.bus.On("torrent:drop", func(payload any) {
		if hashStr, ok := payload.(string); ok {
			m.handleTorrentDrop(hashStr)
		}
	})

	// 2. Запускаем фоновый воркер автозасыпания неактивных раздач
	go m.startAutoSleepWorker()

	return m
}

// ============================================================================
// Управление Политикой Трекеров (Вызывается из плагинов)
// ============================================================================

func (m *Manager) SetTrackerPolicy(mode TrackerMode, trackers []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.trackerPolicy = TrackerPolicy{
		Mode:     mode,
		Trackers: trackers,
	}
	log.Infof("[Torrent Manager] Updated tracker policy: mode=%s, trackers=%d", mode, len(trackers))
}

func (m *Manager) GetTrackerPolicy() TrackerPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.trackerPolicy
}

func (m *Manager) applyTrackerPolicy(spec *torrent.TorrentSpec) {
	m.mu.RLock()
	policy := m.trackerPolicy
	m.mu.RUnlock()

	switch policy.Mode {
	case TrackerModeRemove:
		spec.Trackers = nil
	case TrackerModeReplace:
		var tiers [][]string
		for _, tr := range policy.Trackers {
			tiers = append(tiers, []string{tr})
		}
		spec.Trackers = tiers
	case TrackerModeAppend:
		var tiers [][]string
		for _, tr := range policy.Trackers {
			tiers = append(tiers, []string{tr})
		}
		spec.Trackers = append(spec.Trackers, tiers...)
	case TrackerModeNone:
		// Оставляем исходные без изменений
	}
}

// ============================================================================
// Жизненный цикл раздач (Добавление, Запуск, Получение статуса)
// ============================================================================

// AddTorrent добавляет раздачу (с возможностью не сохранять в БД)
func (m *Manager) AddTorrent(u *user.User, spec *torrent.TorrentSpec, title, poster, category string, saveToDB bool) (*TorrentStatus, error) {
	m.applyTrackerPolicy(spec)
	hashHex := spec.InfoHash.HexString()

	displayTitle := title
	if displayTitle == "" {
		displayTitle = spec.DisplayName
	}

	// 1. Запускаем раздачу в оперативной памяти движка
	session, err := m.engine.Start(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to start engine session: %w", err)
	}

	// 2. Устанавливаем личные метаданные в RAM (для Lampa/NUM)
	session.SetUserMeta(u.ID, displayTitle, poster, category)

	// 3. Если это временный торрент — на этом всё! В базу не пишем.
	if !saveToDB {
		go m.asyncFetchMetadata(session, nil) // Просто ждем метаданные для RAM
		return session.Status(u.ID), nil
	}

	// 4. Если сохраняем в базу — пишем в личный список и в глобальную базу
	if err := m.userSvc.AddTorrent(u, hashHex, displayTitle, poster, category); err != nil {
		return nil, err
	}

	rec := &TorrentRecord{
		Hash:      hashHex,
		Timestamp: time.Now().Unix(),
		Trackers:  flattenTrackers(spec.Trackers),
	}
	_ = m.store.Save(rec)

	go m.asyncFetchMetadata(session, rec)

	return session.Status(u.ID), nil
}

func (m *Manager) asyncFetchMetadata(sess *Session, rec *TorrentRecord) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := sess.WaitInfo(ctx); err != nil {
		log.Warnf("[Torrent Manager] Failed to fetch metadata: %v", err)
		return
	}

	files := sess.Files()

	// Если торрент постоянный (rec != nil), сохраняем физику в БД
	if rec != nil {
		rec.Size = sess.Status("").TorrentSize
		rec.Files = files
		if len(sess.spec.InfoBytes) > 0 {
			rec.InfoBytes = sess.spec.InfoBytes
		}
		_ = m.store.Save(rec)
		log.Infof("[Torrent Manager] Stored metadata in database for: %s (%d files)", rec.Hash, len(files))
	}

	m.bus.Emit("torrent:metadata", map[string]any{
		"hash":  sess.Hash().HexString(),
		"files": len(files),
	})
}

// WakeTorrent будит торрент из базы или продлевает его в движке
func (m *Manager) WakeTorrent(u *user.User, hashHex string) error {
	hash := metainfo.NewHashFromHex(hashHex)

	if sess, ok := m.engine.Get(hash); ok {
		sess.Touch()
		return nil
	}

	rec, err := m.store.Get(hashHex)
	if err != nil {
		return fmt.Errorf("torrent not found in database: %w", err)
	}

	title := "Torrent " + hashHex[:8]
	var poster, category string
	if ut, err := m.userSvc.GetUserTorrent(u.ID, hashHex); err == nil {
		title = ut.Title
		poster = ut.Poster
		category = ut.Category
	}

	spec := &torrent.TorrentSpec{
		InfoHash:    hash,
		DisplayName: title,
		InfoBytes:   rec.InfoBytes,
	}
	if len(rec.InfoBytes) == 0 && len(rec.Trackers) > 0 {
		spec.Trackers = tiersFromList(rec.Trackers)
	}
	m.applyTrackerPolicy(spec)

	sess, err := m.engine.Start(spec)
	if err != nil {
		return fmt.Errorf("failed to wake up torrent: %w", err)
	}

	go m.asyncFetchMetadata(sess, rec)

	sess.SetUserMeta(u.ID, title, poster, category)
	return nil
}

// GetTorrentStatus отдает статус раздачи с учетом запрашивающего пользователя
func (m *Manager) GetTorrentStatus(u *user.User, hashHex string) (*TorrentStatus, error) {
	hash := metainfo.NewHashFromHex(hashHex)

	// Если активен в RAM — отдаем живую статистику (она сама подтянет личные данные юзера)
	sess, ok := m.engine.Get(hash)
	if ok {
		st := sess.Status(u.ID)
		// при добавлении торрента нет данных о нем и нужно брать из базы данные файлы
		rec, recErr := m.store.Get(hashHex)

		if len(st.FileStats) == 0 && recErr == nil && rec != nil {
			st.FileStats = rec.Files
		}
		if st.TorrentSize == 0 && recErr == nil && rec != nil {
			st.TorrentSize = rec.Size
		}
		st.Torrs = packTorrs(hashHex, st.Title, st.Poster, st.Category, st.TorrentSize, flattenTrackers(sess.Trackers()))
		return st, nil
	}

	// Если спит — достаем физику из базы
	rec, err := m.store.Get(hashHex)
	if err != nil {
		return nil, err
	}

	// И достаем личные данные пользователя (название, постер)
	ut, err := m.userSvc.GetUserTorrent(u.ID, hashHex)
	if err != nil {
		// Если торрент есть в базе, но не принадлежит этому юзеру
		return nil, fmt.Errorf("torrent not found in your library")
	}

	return &TorrentStatus{
		Title:       ut.Title,
		Poster:      ut.Poster,
		Category:    ut.Category,
		Hash:        rec.Hash,
		Stat:        TorrentInDB,
		StatString:  TorrentInDB.String(),
		TorrentSize: rec.Size,
		FileStats:   rec.Files,
		Timestamp:   rec.Timestamp,
		Torrs:       packTorrs(rec.Hash, ut.Title, ut.Poster, ut.Category, rec.Size, rec.Trackers),
	}, nil
}

// ============================================================================
// Стриминг: Подготовка потока байт для веб-слоя
// ============================================================================

// streamReader оборачивает torrstor.Reader для поддержки интерфейса io.ReadSeekCloser
// и автоматически уменьшает счетчик ActiveReaders при закрытии плеером
type streamReader struct {
	*torrstor.Reader
	sess *Session
	once sync.Once
}

func (sr *streamReader) Close() error {
	sr.once.Do(func() {
		sr.sess.CloseReader(sr.Reader)
	})
	return nil
}

// GetStreamReader обеспечивает воспроизведение файла (будит торрент из БД при необходимости)
func (m *Manager) GetStreamReader(u *user.User, hashHex string, fileIdx int) (io.ReadSeekCloser, *TorrentFileStat, error) {
	hash := metainfo.NewHashFromHex(hashHex)

	sess, ok := m.engine.Get(hash)
	if !ok {
		// Торрент спит в базе! Будим его:
		rec, err := m.store.Get(hashHex)
		if err != nil {
			return nil, nil, fmt.Errorf("torrent not found in database: %w", err)
		}

		// Fallback-название (если вдруг юзер открыл торрент не из своей библиотеки)
		title := "Torrent " + hashHex[:8]
		var poster, category string

		// Достаем личные данные пользователя (красивое название, постер)
		if ut, err := m.userSvc.GetUserTorrent(u.ID, hashHex); err == nil {
			title = ut.Title
			poster = ut.Poster
			category = ut.Category
		}

		spec := &torrent.TorrentSpec{
			InfoHash:    hash,
			DisplayName: title,
			InfoBytes:   rec.InfoBytes,
		}

		if len(rec.InfoBytes) == 0 && len(rec.Trackers) > 0 {
			spec.Trackers = tiersFromList(rec.Trackers)
		}
		m.applyTrackerPolicy(spec)

		sess, err = m.engine.Start(spec)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to wake up torrent: %w", err)
		}

		// Восстанавливаем личные метаданные в RAM для этого пользователя
		sess.SetUserMeta(u.ID, title, poster, category)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := sess.WaitInfo(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed waiting for info: %w", err)
	}

	files := sess.Files()
	if fileIdx < 0 || fileIdx >= len(files) {
		return nil, nil, fmt.Errorf("file index out of bounds: %d", fileIdx)
	}

	rawReader, err := sess.NewReader(fileIdx)
	if err != nil {
		return nil, nil, err
	}

	wrappedReader := &streamReader{
		Reader: rawReader,
		sess:   sess,
	}

	return wrappedReader, files[fileIdx], nil
}

// ============================================================================
// Фоновый воркер: Автозасыпание при неактивности
// ============================================================================

func (m *Manager) startAutoSleepWorker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkInactiveSessions()
		case <-m.stopWorker:
			return
		}
	}
}

func (m *Manager) checkInactiveSessions() {
	sessions := m.engine.List()
	now := time.Now()

	for _, sess := range sessions {
		if sess.ActiveReaders() == 0 && now.Sub(sess.LastActive()) > m.inactivityTimeout {
			hash := sess.Hash()
			log.Infof("[Torrent Manager] Session %s inactive for %v. Putting to sleep (purging RAM)...", hash.HexString(), m.inactivityTimeout)

			m.engine.Stop(hash)
			m.bus.Emit("torrent:idle", hash.HexString())
		}
	}
}

func (m *Manager) handleTorrentDrop(hashHex string) {
	hash := metainfo.NewHashFromHex(hashHex)
	m.engine.Stop(hash)
	_ = m.store.Delete(hashHex)
	log.Infof("[Torrent Manager] Completely dropped torrent %s from engine and database", hashHex)
}

func (m *Manager) Close() {
	close(m.stopWorker)
	m.bus.UnsubscribeAll()
	_ = m.engine.Close()
	log.Info("[Torrent Manager] Closed successfully")
}

// ============================================================================
// Torrents
// ============================================================================

// ListTorrents возвращает список всех раздач пользователя с их актуальными статусами
func (m *Manager) ListTorrents(u *user.User) ([]*TorrentStatus, error) {
	userTorrents, err := m.userSvc.ListTorrents(u)
	if err != nil {
		return nil, err
	}

	var result []*TorrentStatus
	for _, ut := range userTorrents {
		st, err := m.GetTorrentStatus(u, ut.TorrentHash)
		if err == nil {
			result = append(result, st)
		}
	}
	return result, nil
}

func (m *Manager) RemoveTorrent(u *user.User, hashHex string) error {
	return m.userSvc.RemoveTorrent(u, hashHex)
}

func (m *Manager) SetFileViewed(u *user.User, hashHex string, fileIdx int, viewed bool) error {
	return m.userSvc.SetFileViewed(u, hashHex, fileIdx, viewed)
}

func (m *Manager) SetBlocklistText(text string) error {
	return m.engine.SetBlocklistText(text)
}

// EngineStats — живой снимок состояния движка и процесса для дашборда
type EngineStats struct {
	ActiveSessions int      `json:"active_sessions"`
	ActiveReaders  int      `json:"active_readers"`
	SessionHashes  []string `json:"session_hashes"`
	Goroutines     int      `json:"goroutines"`
	MemAllocBytes  uint64   `json:"mem_alloc_bytes"`
	MemSysBytes    uint64   `json:"mem_sys_bytes"`
}

// Stats возвращает живую статистику для админ-дашборда
func (m *Manager) Stats() EngineStats {
	sessions := m.engine.List()

	st := EngineStats{
		ActiveSessions: len(sessions),
		SessionHashes:  make([]string, 0, len(sessions)),
		Goroutines:     runtime.NumGoroutine(),
	}

	for _, sess := range sessions {
		st.ActiveReaders += sess.ActiveReaders()
		st.SessionHashes = append(st.SessionHashes, sess.Hash().HexString())
	}

	// Читаем память процесса
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	st.MemAllocBytes = ms.Alloc
	st.MemSysBytes = ms.Sys

	return st
}

// GetConfig возвращает текущий конфиг движка из базы данных
func (m *Manager) GetConfig() (*Config, error) {
	return m.store.GetConfig()
}

// SaveConfig сохраняет конфиг движка в базу данных.
// Изменения применяются после перезапуска сервера.
func (m *Manager) SaveConfig(cfg *Config) error {
	if err := m.store.SaveConfig(cfg); err != nil {
		return err
	}
	log.Info("[Torrent Manager] Engine config saved to database (applies after restart)")
	return nil
}

// flattenTrackers превращает тиры [][]string в плоский список
func flattenTrackers(tiers [][]string) []string {
	var list []string
	for _, tier := range tiers {
		list = append(list, tier...)
	}
	return list
}

// tiersFromList оборачивает плоский список трекеров в тиры
func tiersFromList(list []string) [][]string {
	tiers := make([][]string, 0, len(list))
	for _, tr := range list {
		tiers = append(tiers, []string{tr})
	}
	return tiers
}

// packTorrs упаковывает карточку раздачи в строку torrs://
func packTorrs(hash, title, poster, category string, size int64, trackers []string) string {
	th := torrshash.New(hash)
	th.AddField(torrshash.TagTitle, title)
	th.AddField(torrshash.TagPoster, poster)
	th.AddField(torrshash.TagCategory, category)
	th.AddField(torrshash.TagSize, strconv.FormatInt(size, 10))
	for _, tr := range trackers {
		th.AddField(torrshash.TagTracker, tr)
	}

	packed, err := torrshash.Pack(th)
	if err != nil {
		return ""
	}
	return "torrs://" + packed
}

// sizeFromTorrs читает размер из поля TagSize упакованного хэша
func sizeFromTorrs(th *torrshash.TorrsHash) int64 {
	for _, f := range th.Fields {
		if f.Tag == torrshash.TagSize {
			v, _ := strconv.ParseInt(f.Value, 10, 64)
			return v
		}
	}
	return 0
}

// trackersFor возвращает плоский список трекеров: из живой сессии или из базы
func (m *Manager) trackersFor(hashHex string) []string {
	hash := metainfo.NewHashFromHex(hashHex)
	if sess, ok := m.engine.Get(hash); ok {
		return flattenTrackers(sess.Trackers())
	}
	if rec, err := m.store.Get(hashHex); err == nil {
		return rec.Trackers
	}
	return nil
}

// ExportLibrary упаковывает библиотеку пользователя в строки torrs://
func (m *Manager) ExportLibrary(u *user.User) ([]string, error) {
	uts, err := m.userSvc.ListTorrents(u)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0, len(uts))
	for _, ut := range uts {
		size := int64(0)
		if rec, err := m.store.Get(ut.TorrentHash); err == nil {
			size = rec.Size
		}
		line := packTorrs(ut.TorrentHash, ut.Title, ut.Poster, ut.Category, size, m.trackersFor(ut.TorrentHash))
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

// ImportLibrary добавляет строки torrs:// в библиотеку пользователя
func (m *Manager) ImportLibrary(u *user.User, lines []string) (int, error) {
	imported := 0

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "torrs://")

		th, err := torrshash.Unpack(line)
		if err != nil {
			log.Warnf("[Torrent Manager] import: failed to unpack line: %v", err)
			continue
		}

		// Уже есть в библиотеке — пропускаем
		if _, err := m.userSvc.GetUserTorrent(u.ID, th.Hash); err == nil {
			continue
		}

		if err := m.userSvc.AddTorrent(u, th.Hash, th.Title(), th.Poster(), th.Category()); err != nil {
			log.Warnf("[Torrent Manager] import: failed to add %s: %v", th.Hash, err)
			continue
		}

		// Физическая запись в базе с трекерами для пробуждения без DHT
		if _, err := m.store.Get(th.Hash); err != nil {
			rec := &TorrentRecord{
				Hash:      th.Hash,
				Size:      sizeFromTorrs(th),
				Timestamp: time.Now().Unix(),
				Trackers:  th.Trackers(),
			}
			_ = m.store.Save(rec)
		}
		imported++
	}

	return imported, nil
}

// PreloadTorrent запускает предзагрузку файла раздачи
// Доступно только владельцу карточки в библиотеке.
func (m *Manager) PreloadTorrent(u *user.User, hashHex string, fileIdx int) error {
	if _, err := m.userSvc.GetUserTorrent(u.ID, hashHex); err != nil {
		return fmt.Errorf("torrent not found in your library")
	}

	cfg, err := m.store.GetConfig()
	if err != nil {
		return fmt.Errorf("dont open config: %v", err)
	}
	preloadSize := cfg.PreloadSize
	if preloadSize == 0 {
		//preloadSize = 16 * 1024 * 1024
		return nil
	}

	hash := metainfo.NewHashFromHex(hashHex)
	sess, ok := m.engine.Get(hash)
	if !ok {
		return fmt.Errorf("torrent not found in engine")
	}

	files := sess.Files()
	if fileIdx < 0 || fileIdx >= len(files) {
		return fmt.Errorf("file index out of bounds: %d", fileIdx)
	}

	go sess.Preload(fileIdx, preloadSize)

	return nil
}
