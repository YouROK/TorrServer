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

	// Изолированная временная база данных bbolt
	dbPath := filepath.Join(t.TempDir(), "test_manager.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// Сервис пользователей
	cfg := config.DefaultConfig()
	userStore := user.NewStore(db)
	userSvc := user.NewService(userStore, cfg)

	// Авторизуемся под Owner (пароль по умолчанию пустой)
	owner, err := userSvc.Authenticate("")
	if err != nil {
		t.Fatalf("Failed to authenticate owner: %v", err)
	}

	// Чистый движок
	engineCfg := DefaultConfig()
	engineCfg.ListenPort = 0
	engineCfg.DisableDHT = true
	engineCfg.Storage = &torrstor.Config{
		Capacity:    10 << 20,
		NextAheadMB: 1,
		ReadaheadMB: 4,
		UseDisk:     false,
	}

	engine, err := NewEngine(engineCfg)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	t.Cleanup(func() { _ = engine.Close() })

	// Менеджер торрентов
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

	// Режим Replace: заменяем трекеры
	mgr.SetTrackerPolicy(TrackerModeReplace, []string{"udp://custom.tracker:1337/announce"})
	mgr.applyTrackerPolicy(spec)

	if len(spec.Trackers) != 1 || spec.Trackers[0][0] != "udp://custom.tracker:1337/announce" {
		t.Errorf("Expected replaced tracker, got: %v", spec.Trackers)
	}

	// Режим Append: подмешиваем трекер
	mgr.SetTrackerPolicy(TrackerModeAppend, []string{"udp://second.tracker:1337/announce"})
	mgr.applyTrackerPolicy(spec)

	if len(spec.Trackers) != 2 {
		t.Errorf("Expected 2 trackers after append, got: %d", len(spec.Trackers))
	}

	// Режим Remove: удаляем все трекеры (чистый DHT)
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

	// Добавляем торрент с сохранением в базу (saveToDB = true)
	st, err := mgr.AddTorrent(owner, spec, "Test Series S01", "http://poster.jpg", "Series", true)
	if err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}

	if st.Hash != hashHex || st.Title != "Test Series S01" {
		t.Errorf("Invalid status returned: %+v", st)
	}

	// Торрент активен в оперативной памяти (Engine)
	if _, inRAM := engine.Get(spec.InfoHash); !inRAM {
		t.Errorf("Session should be active in RAM engine")
	}

	// Даем 100 мс на сохранение метаданных в базу в фоновой горутине
	time.Sleep(100 * time.Millisecond)

	// Физическая карточка создана в базе данных bbolt
	rec, err := tStore.Get(hashHex)
	if err != nil {
		t.Fatalf("Torrent card missing from database: %v", err)
	}
	if rec.Hash != hashHex {
		t.Errorf("Database record mismatch: %+v", rec)
	}

	// Сон: остановка только в Engine RAM
	engine.Stop(spec.InfoHash)

	// Убеждаемся, что в RAM его больше нет
	if _, inRAM := engine.Get(spec.InfoHash); inRAM {
		t.Errorf("Session should not be in RAM after Stop")
	}

	// Единый статус спящего торрента для Owner
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

	// Торрент БЕЗ сохранения в базу (saveToDB = false)
	_, err := mgr.AddTorrent(owner, spec, "Lampa Ephemeral Movie", "", "", false)
	if err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}

	// Раздача крутится в RAM
	if _, inRAM := engine.Get(spec.InfoHash); !inRAM {
		t.Errorf("Ephemeral session should be active in RAM")
	}

	// Остановка (симуляция выключения плеера и автозасыпания)
	engine.Stop(spec.InfoHash)

	// В базе данных должно быть пусто
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

	// Добавляем торрент и ждем сохранения (saveToDB = true)
	_, err := mgr.AddTorrent(owner, spec, "Wake Up Test", "", "", true)
	if err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Сон торрента - полная выгрузка из RAM
	engine.Stop(spec.InfoHash)
	if _, inRAM := engine.Get(spec.InfoHash); inRAM {
		t.Fatalf("Failed to stop session in RAM")
	}

	// Стрим спящего файла должен разбудить торрент
	reader, fileStat, err := mgr.GetStreamReader(owner, hashHex, 0)
	if err != nil {
		t.Fatalf("GetStreamReader failed to wake up torrent: %v", err)
	}
	defer reader.Close()

	if fileStat.Id != 0 || fileStat.Name != "S01E01.mkv" {
		t.Errorf("Invalid file returned: %+v", fileStat)
	}

	// Торрент снова работает в оперативной памяти
	sess, inRAM := engine.Get(spec.InfoHash)
	if !inRAM {
		t.Errorf("Torrent should have been woken up into RAM engine")
	}

	// Счетчик активных читателей должен быть равен 1
	if sess.ActiveReaders() != 1 {
		t.Errorf("Expected 1 active reader, got: %d", sess.ActiveReaders())
	}

	// Закрытие ридера (плеер закончил просмотр)
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

// TestManager_TitleFromMetadata проверяет подстановку имени раздачи из метаданных,
// когда пользователь не задал свое название (голый хэш или magnet без dn)
func TestManager_TitleFromMetadata(t *testing.T) {
	mgr, _, engine, tStore, owner := setupManagerTest(t)

	spec := createSyntheticSpec(t)
	hashHex := spec.InfoHash.HexString()

	// Добавляем без названия - имя должно прийти из метаданных "Test_Series_Season_1"
	st, err := mgr.AddTorrent(owner, spec, "", "", "", true)
	if err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}

	// Информация уже в InfoBytes, поэтому имя известно сразу
	if st.Title != "Test_Series_Season_1" {
		t.Errorf("Expected name from info bytes, got: %q", st.Title)
	}

	// Ждем фонового сохранения метаданных
	time.Sleep(200 * time.Millisecond)

	rec, err := tStore.Get(hashHex)
	if err != nil {
		t.Fatalf("Torrent card missing from database: %v", err)
	}
	if rec.Name != "Test_Series_Season_1" {
		t.Errorf("Expected metadata name in DB, got: %q", rec.Name)
	}

	// Личная карточка осталась без названия: его подставляет резолв при чтении
	ut, err := mgr.userSvc.GetUserTorrent(owner.ID, hashHex)
	if err != nil {
		t.Fatalf("User torrent missing: %v", err)
	}
	if ut.Title != "" {
		t.Errorf("User title should stay empty, got: %q", ut.Title)
	}

	// Живой статус отдает имя из метаданных
	st, err = mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed: %v", err)
	}
	if st.Title != "Test_Series_Season_1" {
		t.Errorf("Expected metadata title, got: %q", st.Title)
	}

	// Спящий статус тоже отдает имя из метаданных
	engine.Stop(spec.InfoHash)
	sleepSt, err := mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed for sleeping torrent: %v", err)
	}
	if sleepSt.Title != "Test_Series_Season_1" {
		t.Errorf("Expected metadata title for sleeping torrent, got: %q", sleepSt.Title)
	}

	// Личное название пользователя имеет приоритет над именем из метаданных
	if err := mgr.userSvc.UpdateTorrentMeta(owner, hashHex, "My Own Title", "", ""); err != nil {
		t.Fatalf("UpdateTorrentMeta failed: %v", err)
	}
	st, err = mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed: %v", err)
	}
	if st.Title != "My Own Title" {
		t.Errorf("Expected user title to win, got: %q", st.Title)
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

	// Раздача выгружена из памяти Engine
	if _, inRAM := engine.Get(spec.InfoHash); inRAM {
		t.Errorf("Session should have been dropped from RAM engine")
	}

	// Карточка полностью удалена из базы bbolt
	_, err = tStore.Get(hashHex)
	if err != ErrTorrentNotFound {
		t.Errorf("Torrent record should have been deleted from database, got: %v", err)
	}
}

// TestManager_ViewedFilesInStatus проверяет, что отметки просмотра попадают
// в статус раздачи и в активной ветке, и в спящей.
func TestManager_ViewedFilesInStatus(t *testing.T) {
	mgr, _, engine, _, owner := setupManagerTest(t)

	spec := createSyntheticSpec(t)
	hashHex := spec.InfoHash.HexString()

	if _, err := mgr.AddTorrent(owner, spec, "Viewed Test", "", "", true); err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}

	if err := mgr.SetFileViewed(owner, hashHex, 1, true); err != nil {
		t.Fatalf("SetFileViewed failed: %v", err)
	}
	if err := mgr.SetFileViewed(owner, hashHex, 0, true); err != nil {
		t.Fatalf("SetFileViewed failed: %v", err)
	}
	// Снятие отметки тоже должно отражаться в статусе
	if err := mgr.SetFileViewed(owner, hashHex, 0, false); err != nil {
		t.Fatalf("SetFileViewed(false) failed: %v", err)
	}

	// Активная ветка: раздача в RAM
	live, err := mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed: %v", err)
	}
	if len(live.ViewedFiles) != 1 || live.ViewedFiles[0] != 1 {
		t.Errorf("live viewed_files = %v, want [1]", live.ViewedFiles)
	}

	// Спящая ветка: раздача выгружена из RAM
	engine.Stop(spec.InfoHash)

	sleeping, err := mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed for sleeping torrent: %v", err)
	}
	if len(sleeping.ViewedFiles) != 1 || sleeping.ViewedFiles[0] != 1 {
		t.Errorf("sleeping viewed_files = %v, want [1]", sleeping.ViewedFiles)
	}

	// Отметки одного пользователя не видны другому
	other, _ := mgr.userSvc.CreateUser(owner, "viewer", "pass", user.RankUser, user.Limits{})
	if err := mgr.userSvc.AddTorrent(other, hashHex, "Other Title", "", ""); err != nil {
		t.Fatalf("AddTorrent for other user failed: %v", err)
	}
	otherSt, err := mgr.GetTorrentStatus(other, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus for other user failed: %v", err)
	}
	if len(otherSt.ViewedFiles) != 0 {
		t.Errorf("other user viewed_files = %v, want empty", otherSt.ViewedFiles)
	}
}

// TestManager_SetAllFilesViewed проверяет массовую отметку и снятие просмотра.
func TestManager_SetAllFilesViewed(t *testing.T) {
	mgr, userSvc, engine, _, owner := setupManagerTest(t)

	spec := createSyntheticSpec(t)
	hashHex := spec.InfoHash.HexString()

	if _, err := mgr.AddTorrent(owner, spec, "Bulk Viewed", "", "", true); err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}

	// Метаданные сохраняются в базу фоновой горутиной, а индексы файлов
	// для отметки берутся из статуса: ждём их появления
	time.Sleep(100 * time.Millisecond)

	// Отмечаем все файлы раздачи
	if err := mgr.SetAllFilesViewed(owner, hashHex, true); err != nil {
		t.Fatalf("SetAllFilesViewed(true) failed: %v", err)
	}

	st, err := mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed: %v", err)
	}
	if len(st.FileStats) == 0 {
		t.Fatal("у раздачи нет файлов, проверка невозможна")
	}
	if len(st.ViewedFiles) != len(st.FileStats) {
		t.Errorf("viewed_files = %v, want все %d файлов", st.ViewedFiles, len(st.FileStats))
	}

	// Снятие отметок у всех
	if err := mgr.SetAllFilesViewed(owner, hashHex, false); err != nil {
		t.Fatalf("SetAllFilesViewed(false) failed: %v", err)
	}
	st, err = mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed: %v", err)
	}
	if len(st.ViewedFiles) != 0 {
		t.Errorf("viewed_files = %v, want пусто", st.ViewedFiles)
	}

	// Работает и для спящей раздачи: индексы берутся из базы
	engine.Stop(spec.InfoHash)
	if err := mgr.SetAllFilesViewed(owner, hashHex, true); err != nil {
		t.Fatalf("SetAllFilesViewed for sleeping torrent failed: %v", err)
	}
	sleeping, err := mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed: %v", err)
	}
	if len(sleeping.ViewedFiles) != len(sleeping.FileStats) {
		t.Errorf("sleeping viewed_files = %v, want все %d файлов", sleeping.ViewedFiles, len(sleeping.FileStats))
	}

	// Повторная отметка не дублирует индексы
	if err := mgr.SetAllFilesViewed(owner, hashHex, true); err != nil {
		t.Fatalf("повторная SetAllFilesViewed failed: %v", err)
	}
	again, err := mgr.GetTorrentStatus(owner, hashHex)
	if err != nil {
		t.Fatalf("GetTorrentStatus failed: %v", err)
	}
	if len(again.ViewedFiles) != len(sleeping.ViewedFiles) {
		t.Errorf("повторная отметка изменила список: %v -> %v", sleeping.ViewedFiles, again.ViewedFiles)
	}

	_ = userSvc
}

// TestManager_StreamReaderRejectsForeignTorrent проверяет, что чужую раздачу
// нельзя открыть по одному хэшу, минуя библиотеку пользователя.
func TestManager_StreamReaderRejectsForeignTorrent(t *testing.T) {
	mgr, userSvc, _, _, owner := setupManagerTest(t)

	spec := createSyntheticSpec(t)
	hashHex := spec.InfoHash.HexString()

	if _, err := mgr.AddTorrent(owner, spec, "Owner Movie", "", "", true); err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	other, err := userSvc.CreateUser(owner, "stranger", "pass", user.RankUser, user.Limits{})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if _, _, err := mgr.GetStreamReader(other, hashHex, 0); err != ErrNotOwned {
		t.Errorf("GetStreamReader for foreign torrent = %v, want %v", err, ErrNotOwned)
	}

	if _, err := mgr.GetTorrentStatus(other, hashHex); err != ErrNotOwned {
		t.Errorf("GetTorrentStatus for foreign torrent = %v, want %v", err, ErrNotOwned)
	}

	if err := mgr.WakeTorrent(other, hashHex); err != ErrNotOwned {
		t.Errorf("WakeTorrent for foreign torrent = %v, want %v", err, ErrNotOwned)
	}

	// Владелец по-прежнему получает доступ
	reader, _, err := mgr.GetStreamReader(owner, hashHex, 0)
	if err != nil {
		t.Fatalf("GetStreamReader for owner failed: %v", err)
	}
	_ = reader.Close()
}

// TestManager_EphemeralTorrentStaysReadable проверяет, что временная раздача
// (save_to_db = false) читается её владельцем: записи в библиотеке у неё нет.
func TestManager_EphemeralTorrentStaysReadable(t *testing.T) {
	mgr, _, _, tStore, owner := setupManagerTest(t)

	spec := createSyntheticSpec(t)
	hashHex := spec.InfoHash.HexString()

	if _, err := mgr.AddTorrent(owner, spec, "Ephemeral", "", "", false); err != nil {
		t.Fatalf("AddTorrent failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Убеждаемся, что в библиотеке и базе записи действительно нет
	if _, err := mgr.userSvc.GetUserTorrent(owner.ID, hashHex); err == nil {
		t.Fatal("Ephemeral torrent must not be stored in the library")
	}
	if _, err := tStore.Get(hashHex); err != ErrTorrentNotFound {
		t.Fatalf("Ephemeral torrent must not be in database, got %v", err)
	}

	reader, _, err := mgr.GetStreamReader(owner, hashHex, 0)
	if err != nil {
		t.Fatalf("GetStreamReader for own ephemeral torrent failed: %v", err)
	}
	_ = reader.Close()

	if _, err := mgr.GetTorrentStatus(owner, hashHex); err != nil {
		t.Fatalf("GetTorrentStatus for own ephemeral torrent failed: %v", err)
	}
}
