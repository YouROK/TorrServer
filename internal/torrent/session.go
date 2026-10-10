package torrent

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"silo/internal/bus"
	"silo/internal/log"
	"silo/internal/torrent/storage/torrstor"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type Session struct {
	t         *torrent.Torrent
	storage   *torrstor.Storage
	cache     *torrstor.Cache
	spec      *torrent.TorrentSpec
	stat      TorrentStat
	files     []*TorrentFileStat
	timestamp int64
	mu        sync.RWMutex
	trackers  [][]string

	// Личные метаданные пользователей, которые сейчас смотрят раздачу
	userMeta map[string]EphemeralMeta

	// Имя раздачи из полученных метаданных
	metaName string

	// Скорость и статистика
	lastSpeedCalc   time.Time
	downloadSpeed   float64
	uploadSpeed     float64
	bytesReadUseful int64
	bytesWritten    int64
	preloadedBytes  int64
	preloadSize     int64

	// Активность и автозасыпание
	activeReaders int32
	lastActive    time.Time
	closed        chan struct{}
	ticker        *time.Ticker

	// Фоновые задачи раздачи: закрытие сессии дожидается их завершения
	tasks *bus.Group
}

func newSession(t *torrent.Torrent, spec *torrent.TorrentSpec, stor *torrstor.Storage) *Session {
	s := &Session{
		t:         t,
		spec:      spec,
		storage:   stor,
		stat:      TorrentAdded,
		timestamp: time.Now().Unix(),
		userMeta:  make(map[string]EphemeralMeta),
		trackers:  spec.Trackers,
		// Готовые info-байты несут имя сразу, без ожидания сети
		metaName:      InfoName(spec.InfoBytes),
		lastSpeedCalc: time.Now(),
		lastActive:    time.Now(),
		closed:        make(chan struct{}),
		tasks:         bus.NewGroup(context.Background()),
	}

	// Если метаданные уже были на старте - привязываем кэш сразу
	if stor != nil {
		s.cache = stor.GetCache(spec.InfoHash)
		if s.cache != nil {
			s.cache.SetTorrent(t)
		}
	}

	s.tasks.Go(func(ctx context.Context) { s.watchProgress(ctx) })
	return s
}

// WaitInfo ожидает загрузки метаданных из сети
func (s *Session) WaitInfo(ctx context.Context) error {
	select {
	case <-s.t.GotInfo():
		s.mu.Lock()
		s.stat = TorrentWorking
		s.initFiles()
		s.metaName = s.t.Name()

		// Подтягиваем кэш из хранилища сразу после получения метаданных!
		if s.storage != nil {
			s.cache = s.storage.GetCache(s.spec.InfoHash)
			if s.cache != nil {
				s.cache.SetTorrent(s.t)
			}
		}
		s.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-s.closed:
		return fmt.Errorf("session closed")
	}
}

func (s *Session) initFiles() {
	if s.t.Info() == nil {
		return
	}
	files := s.t.Files()
	s.files = make([]*TorrentFileStat, 0, len(files))

	for i, f := range files {
		cleanPath := f.Path()
		s.files = append(s.files, &TorrentFileStat{
			Id:       i,
			Path:     cleanPath,
			Name:     filepath.Base(cleanPath),
			Length:   f.Length(),
			FileHash: GenerateFileHash(cleanPath, f.Length()),
		})
	}
}

func (s *Session) Files() []*TorrentFileStat {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.files
}

// MetaName возвращает имя раздачи из полученных метаданных
func (s *Session) MetaName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metaName
}

// SetMetaName сохраняет имя из базы, пока движок не получил метаданные
func (s *Session) SetMetaName(name string) {
	if name == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.metaName == "" {
		s.metaName = name
	}
}

// NewReader создает ридер для воспроизведения файла
func (s *Session) NewReader(fileIdx int) (*torrstor.Reader, error) {
	s.Touch()
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.t.Info() == nil {
		return nil, fmt.Errorf("torrent metadata not loaded yet")
	}

	// Гарантируем привязку кэша
	if s.cache == nil && s.storage != nil {
		s.cache = s.storage.GetCache(s.spec.InfoHash)
		if s.cache != nil {
			s.cache.SetTorrent(s.t)
		}
	}

	if s.cache == nil {
		return nil, fmt.Errorf("torrent cache not initialized")
	}

	files := s.t.Files()
	if fileIdx < 0 || fileIdx >= len(files) {
		return nil, fmt.Errorf("file index out of bounds: %d", fileIdx)
	}

	atomic.AddInt32(&s.activeReaders, 1)
	reader := s.cache.NewReader(files[fileIdx])
	return reader, nil
}

func (s *Session) CloseReader(r *torrstor.Reader) {
	if s.cache != nil && r != nil {
		s.cache.CloseReader(r)
	}
	atomic.AddInt32(&s.activeReaders, -1)
	s.Touch()
}

func (s *Session) Preload(fileIdx int, size int64) {
	s.Touch()
	s.mu.Lock()
	if s.t.Info() == nil || size <= 0 {
		s.mu.Unlock()
		return
	}
	s.stat = TorrentPreload
	s.preloadSize = size
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		if s.stat == TorrentPreload {
			s.stat = TorrentWorking
		}
		s.mu.Unlock()
	}()

	files := s.t.Files()
	if fileIdx < 0 || fileIdx >= len(files) {
		return
	}
	file := files[fileIdx]

	if size > file.Length() {
		size = file.Length()
	}

	pieceLen := s.t.Info().PieceLength
	startEndSize := pieceLen
	if startEndSize < 8<<20 {
		startEndSize = 8 << 20
	}

	// Хвост файла
	if file.Length() > startEndSize*2 {
		tailReader := file.NewReader()
		tailReader.SetResponsive()
		tailReader.SetReadahead(0)
		tailReader.SetNext(0)
		tailReader.SetZone(startEndSize)

		tailOffset := file.Length() - startEndSize
		_, _ = tailReader.Seek(tailOffset, io.SeekStart)

		buf := make([]byte, 32768)
		for tailOffset < file.Length() {
			n, err := tailReader.Read(buf)
			tailOffset += int64(n)
			if err != nil {
				break
			}
		}
		tailReader.Close()
	}

	// Начало файла
	headReader := file.NewReader()
	headReader.SetResponsive()
	readahead := pieceLen * 4
	headReader.SetReadahead(readahead)
	headReader.SetNext(readahead)
	headReader.SetZone(size)

	var offset int64
	buf := make([]byte, 32768)
	for offset < size {
		n, err := headReader.Read(buf)
		offset += int64(n)
		if err != nil {
			break
		}
	}
	headReader.Close()
}

func (s *Session) watchProgress(ctx context.Context) {
	s.ticker = time.NewTicker(time.Second)
	defer s.ticker.Stop()

	for {
		select {
		case <-s.ticker.C:
			s.calculateSpeed()
		case <-s.closed:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Session) calculateSpeed() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.t != nil && s.t.Info() != nil {
		st := s.t.Stats()
		deltaDl := st.BytesRead.Int64() - s.bytesReadUseful
		deltaUp := st.BytesWritten.Int64() - s.bytesWritten
		elapsed := time.Since(s.lastSpeedCalc).Seconds()

		if elapsed > 0 {
			s.downloadSpeed = float64(deltaDl) / elapsed
			s.uploadSpeed = float64(deltaUp) / elapsed
		}

		s.bytesReadUseful = st.BytesRead.Int64()
		s.bytesWritten = st.BytesWritten.Int64()

		if s.cache != nil {
			s.preloadedBytes = s.cache.GetState().Filled
		}
	} else {
		s.downloadSpeed = 0
		s.uploadSpeed = 0
	}

	s.lastSpeedCalc = time.Now()
}

func (s *Session) Touch() {
	s.mu.Lock()
	s.lastActive = time.Now()
	s.mu.Unlock()
}

func (s *Session) ActiveReaders() int {
	return int(atomic.LoadInt32(&s.activeReaders))
}

func (s *Session) LastActive() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastActive
}

func (s *Session) Hash() metainfo.Hash {
	return s.spec.InfoHash
}

// SetUserMeta сохраняет временное название и постер для конкретного пользователя
func (s *Session) SetUserMeta(userID, title, poster, category string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userMeta[userID] = EphemeralMeta{
		Title:    title,
		Poster:   poster,
		Category: category,
	}
}

// UserTitle возвращает личное название пользователя, без имени из метаданных
func (s *Session) UserTitle(userID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userMeta[userID].Title
}

// Status возвращает статус раздачи с учетом личных метаданных запрашивающего пользователя
func (s *Session) Status(userID string) *TorrentStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hashHex := s.spec.InfoHash.HexString()

	// Достаем личные данные пользователя (если есть)
	meta := s.userMeta[userID]
	displayTitle := ResolveTitle(meta.Title, s.metaName, hashHex)

	st := &TorrentStatus{
		Title:          displayTitle,
		Poster:         meta.Poster,
		Category:       meta.Category,
		Hash:           s.spec.InfoHash.HexString(),
		Timestamp:      s.timestamp,
		Stat:           s.stat,
		StatString:     s.stat.String(),
		PreloadedBytes: s.preloadedBytes,
		PreloadSize:    s.preloadSize,
		DownloadSpeed:  s.downloadSpeed,
		UploadSpeed:    s.uploadSpeed,
		FileStats:      s.files,
	}

	if s.t != nil && s.t.Info() != nil {
		st.Name = s.t.Name()
		st.LoadedSize = s.t.BytesCompleted()
		st.TorrentSize = s.t.Length()

		tst := s.t.Stats()
		st.TotalPeers = tst.TotalPeers
		st.ActivePeers = tst.ActivePeers
		st.ConnectedSeeders = tst.ConnectedSeeders
		st.HalfOpenPeers = tst.HalfOpenPeers
		st.BytesRead = tst.BytesRead.Int64()
		st.BytesWritten = tst.BytesWritten.Int64()
	}

	return st
}

func (s *Session) Close() {
	select {
	case <-s.closed:
		return
	default:
		close(s.closed)
	}

	// Сначала фоновые задачи: они не должны обращаться к движку и хранилищу после их остановки
	if s.tasks != nil && !s.tasks.Close(3*time.Second) {
		log.Warnf("[Torrent Engine] Session tasks did not stop in time: %s", s.spec.InfoHash.HexString())
	}

	if s.t != nil {
		s.t.Drop()
	}
}

// AddTrackers безопасно подключает новые трекеры к работающей раздаче
func (s *Session) AddTrackers(trackers []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(trackers) == 0 {
		return
	}

	var tiers [][]string
	for _, tr := range trackers {
		tiers = append(tiers, []string{tr})
	}

	s.trackers = append(s.trackers, tiers...)

	if s.t != nil {
		s.t.AddTrackers(tiers)
	}
}

// Trackers возвращает актуальный список трекеров раздачи
func (s *Session) Trackers() [][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.trackers
}

// CacheState возвращает снимок состояния кэша раздачи.
func (s *Session) CacheState() *torrstor.CacheState {
	if s.cache == nil {
		return nil
	}
	return s.cache.GetState()
}

// PeerStates возвращает снимок состояния подключенных пиров раздачи.
func (s *Session) PeerStates() []torrent.PeerStatus {
	if s.t == nil {
		return nil
	}
	return s.t.PeerStatuses()
}
