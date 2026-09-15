package torrent

import (
	"crypto/sha1"
	"path/filepath"
	"testing"
	"time"

	"silo/internal/config"
	"silo/internal/database"
	"silo/internal/torrent/storage/torrstor"
	"silo/internal/user"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

// createSyntheticSpec создает тестовую раздачу прямо в оперативной памяти (без интернета)
func createSyntheticSpec(t *testing.T) *torrent.TorrentSpec {
	t.Helper()

	info := metainfo.Info{
		PieceLength: 32 * 1024, // Кусок 32 КБ
		Name:        "Test_Series_Season_1",
		Files: []metainfo.FileInfo{
			{Length: 64 * 1024, Path: []string{"Season 1", "S01E01.mkv"}},
			{Length: 64 * 1024, Path: []string{"Season 1", "S01E02.mkv"}},
		},
	}

	totalLength := int64(128 * 1024)
	numPieces := (totalLength + info.PieceLength - 1) / info.PieceLength
	info.Pieces = make([]byte, numPieces*20)

	infoBytes, err := bencode.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to bencode synthetic info: %v", err)
	}

	h := sha1.Sum(infoBytes)
	infoHash := metainfo.Hash(h)

	return &torrent.TorrentSpec{
		InfoBytes:   infoBytes,
		InfoHash:    infoHash,
		DisplayName: "Test Series S01",
	}
}

// setupManagerTest создает тестовую БД, сервисы пользователей, движок и менеджер
func setupManagerTest(t *testing.T) (*Manager, *user.Service, *Engine, *Store, *user.User) {
	t.Helper()

	// 1. Изолированная временная база данных bbolt
	dbPath := filepath.Join(t.TempDir(), "test_manager.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// 2. Сервис пользователей
	cfg := config.DefaultConfig()
	userStore := user.NewStore(db)
	userSvc := user.NewService(userStore, cfg)

	// Авторизуемся под Owner (пароль по умолчанию пустой)
	owner, err := userSvc.Authenticate("")
	if err != nil {
		t.Fatalf("Failed to authenticate owner: %v", err)
	}

	// 3. Чистый движок
	engineCfg := DefaultConfig()
	engineCfg.ListenPort = 0
	engineCfg.Storage = &torrstor.Config{
		Capacity: 10 << 20,
		UseDisk:  false,
	}

	engine, err := NewEngine(engineCfg)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	t.Cleanup(func() { _ = engine.Close() })

	// 4. Менеджер торрентов
	tStore := NewStore(db)
	mgr := NewManager(engine, tStore, userSvc)
	t.Cleanup(func() { mgr.Close() })

	return mgr, userSvc, engine, tStore, owner
}

// TestManager_TrackerPolicy проверяет работу правил трекеров от плагинов
func TestManager_TrackerPolicy(t *testing.T) {
	mgr, _, _, _, _ := setupManagerTest(t)

	spec := &torrent.TorrentSpec{
		Trackers: [][]string{{"udp://initial.tracker:1337/announce"}},
	}

	// 1. Режим Replace: заменяем трекеры
	mgr.SetTrackerPolicy(TrackerModeReplace, []string{"udp://custom.tracker:1337/announce"})
	mgr.applyTrackerPolicy(spec)

	if len(spec.Trackers) != 1 || spec.Trackers[0][0] != "udp://custom.tracker:1337/announce" {
		t.Errorf("Expected replaced tracker, got: %v", spec.Trackers)
	}

	// 2. Режим Append: подмешиваем трекер
	mgr.SetTrackerPolicy(TrackerModeAppend, []string{"udp://second.tracker:1337/announce"})
	mgr.applyTrackerPolicy(spec)

	if len(spec.Trackers) != 2 {
		t.Errorf("Expected 2 trackers after append, got: %d", len(spec.Trackers))
	}

	// 3. Режим Remove: удаляем все трекеры (чистый DHT)
	mgr.SetTrackerPolicy(TrackerModeRemove, nil)
	mgr.applyTrackerPolicy(spec)

	if len(spec.Trackers) != 0 {
		t.Errorf("Expected 0 trackers after remove, got: %v", spec.Trackers)
	}
}

// TestManager_AddAndUnifiedStatus проверяет запуск в RAM, сохранение в БД и переключение статусов
func TestManager_AddAndUnifiedStatus(t *testing.T) {
	mgr, _, engine, tStore, owner := setupManagerTest(t)

	spec := createSyntheticSpec(t)
	hashHex := spec.InfoHash.HexString()

	// 1. Добавляем торрент с сохранением в базу (saveToDB = true)
	st, err := mgr.AddTorrent(owner, spec, "Test Series S01", "http://poster.jpg", "Series", true)
	if err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}

	if st.Hash != hashHex || st.Title != "Test Series S01" {
		t.Errorf("Invalid status returned: %+v", st)
	}

	// 2. Проверяем, что торрент активен в оперативной памяти (Engine)
	if _, inRAM := engine.Get(spec.InfoHash); !inRAM {
		t.Errorf("Session should be active in RAM engine")
	}

	// Даем 100 мс на сохранение метаданных в базу в фоновой горутине
	time.Sleep(100 * time.Millisecond)

	// 3. Проверяем, что физическая карточка создана в базе данных bbolt
	rec, err := tStore.Get(hashHex)
	if err != nil {
		t.Fatalf("Torrent card missing from database: %v", err)
	}
	if rec.Hash != hashHex {
		t.Errorf("Database record mismatch: %+v", rec)
	}

	// 4. ТЕСТ СНА: Усыпляем торрент (останавливаем только в Engine RAM)
	engine.Stop(spec.InfoHash)

	// Убеждаемся, что в RAM его больше нет
	if _, inRAM := engine.Get(spec.InfoHash); inRAM {
		t.Errorf("Session should not be in RAM after Stop")
	}

	// 5. ЕДИНЫЙ СТАТУС: Запрашиваем статус спящего торрента для Owner
	sleepSt, err := mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed for sleeping torrent: %v", err)
	}

	// Он должен автоматически выдать статус TorrentInDB и подтянуть личное название!
	if sleepSt.Stat != TorrentInDB {
		t.Errorf("Expected status TorrentInDB, got: %v", sleepSt.Stat)
	}
	if sleepSt.Title != "Test Series S01" || sleepSt.Hash != hashHex {
		t.Errorf("Metadata mismatch in sleeping status: %+v", sleepSt)
	}
}

// TestManager_EphemeralTorrent проверяет ВРЕМЕННЫЕ торренты (не сохраняются в БД)
func TestManager_EphemeralTorrent(t *testing.T) {
	mgr, _, engine, tStore, owner := setupManagerTest(t)

	spec := createSyntheticSpec(t)
	hashHex := spec.InfoHash.HexString()

	// 1. Добавляем торрент БЕЗ сохранения в базу (saveToDB = false)
	_, err := mgr.AddTorrent(owner, spec, "Lampa Ephemeral Movie", "", "", false)
	if err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}

	// 2. Проверяем, что он крутится в RAM
	if _, inRAM := engine.Get(spec.InfoHash); !inRAM {
		t.Errorf("Ephemeral session should be active in RAM")
	}

	// 3. Останавливаем (симулируем выключение плеера и автозасыпание)
	engine.Stop(spec.InfoHash)

	// 4. Проверяем базу данных: ТАМ ДОЛЖНО БЫТЬ ПУСТО!
	_, err = tStore.Get(hashHex)
	if err != ErrTorrentNotFound {
		t.Errorf("Ephemeral torrent leaked into database! Expected ErrTorrentNotFound, got: %v", err)
	}
}

// TestManager_WakeUpOnStream проверяет автоматическое пробуждение торрента при запросе видео
func TestManager_WakeUpOnStream(t *testing.T) {
	mgr, _, engine, _, owner := setupManagerTest(t)

	spec := createSyntheticSpec(t)
	hashHex := spec.InfoHash.HexString()

	// 1. Добавляем торрент и ждем сохранения (saveToDB = true)
	_, err := mgr.AddTorrent(owner, spec, "Wake Up Test", "", "", true)
	if err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// 2. Усыпляем торрент (полная выгрузка из RAM)
	engine.Stop(spec.InfoHash)
	if _, inRAM := engine.Get(spec.InfoHash); inRAM {
		t.Fatalf("Failed to stop session in RAM")
	}

	// 3. Плеер запрашивает стрим спящего файла: менеджер ДОЛЖЕН САМ ЕГО РАЗБУДИТЬ!
	reader, fileStat, err := mgr.GetStreamReader(owner, hashHex, 0)
	if err != nil {
		t.Fatalf("GetStreamReader failed to wake up torrent: %v", err)
	}
	defer reader.Close()

	if fileStat.Id != 0 || fileStat.Name != "S01E01.mkv" {
		t.Errorf("Invalid file returned: %+v", fileStat)
	}

	// 4. Проверяем: торрент снова проснулся и работает в оперативной памяти!
	sess, inRAM := engine.Get(spec.InfoHash)
	if !inRAM {
		t.Errorf("Torrent should have been woken up into RAM engine")
	}

	// Счетчик активных читателей должен быть равен 1
	if sess.ActiveReaders() != 1 {
		t.Errorf("Expected 1 active reader, got: %d", sess.ActiveReaders())
	}

	// 5. Закрываем ридер (плеер закончил просмотр)
	reader.Close()
	if sess.ActiveReaders() != 0 {
		t.Errorf("Expected 0 active readers after Close, got: %d", sess.ActiveReaders())
	}
}

// TestManager_AutoSleep проверяет выгрузку неактивных раздач по таймауту
func TestManager_AutoSleep(t *testing.T) {
	mgr, _, engine, _, owner := setupManagerTest(t)

	// Ставим короткий таймаут неактивности для теста
	mgr.inactivityTimeout = 50 * time.Millisecond

	spec := createSyntheticSpec(t)
	// saveToDB = true
	_, err := mgr.AddTorrent(owner, spec, "Auto Sleep Test", "", "", true)
	if err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}

	// Убеждаемся, что раздача в RAM
	if _, inRAM := engine.Get(spec.InfoHash); !inRAM {
		t.Fatalf("Session should be active in RAM")
	}

	// Ждем истечения таймаута
	time.Sleep(70 * time.Millisecond)

	// Запускаем проверку неактивности
	mgr.checkInactiveSessions()

	// Торрент должен автоматически уснуть (выгрузиться из RAM)!
	if _, inRAM := engine.Get(spec.InfoHash); inRAM {
		t.Errorf("Session should have been purged from RAM by auto-sleep")
	}
}

// TestManager_DropOnUserDelete проверяет полное удаление из БД и RAM по сигналу шины
func TestManager_DropOnUserDelete(t *testing.T) {
	mgr, userSvc, engine, tStore, owner := setupManagerTest(t)

	spec := createSyntheticSpec(t)
	hashHex := spec.InfoHash.HexString()

	// Добавляем торрент (saveToDB = true)
	_, _ = mgr.AddTorrent(owner, spec, "Drop Test", "", "", true)
	time.Sleep(50 * time.Millisecond)

	// Пользователь удаляет торрент из своего списка.
	// Так как других пользователей нет, сервис бросит в шину событие torrent:drop!
	err := userSvc.RemoveTorrent(owner, hashHex)
	if err != nil {
		t.Fatalf("RemoveTorrent failed: %v", err)
	}

	// Даем шине 50 мс обработать событие
	time.Sleep(50 * time.Millisecond)

	// 1. Проверяем: раздача выгружена из памяти Engine
	if _, inRAM := engine.Get(spec.InfoHash); inRAM {
		t.Errorf("Session should have been dropped from RAM engine")
	}

	// 2. Проверяем: карточка полностью удалена из базы bbolt
	_, err = tStore.Get(hashHex)
	if err != ErrTorrentNotFound {
		t.Errorf("Torrent record should have been deleted from database, got: %v", err)
	}
}
