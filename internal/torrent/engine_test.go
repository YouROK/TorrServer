package torrent

import (
	"context"
	"testing"
	"time"

	"silo/internal/torrent/storage/torrstor"
)

// TestGenerateFileHash проверяет устойчивость хэша при смене имени папки релиза
func TestGenerateFileHash(t *testing.T) {
	// Релиз 1: папка "Series.720p"
	hash1 := GenerateFileHash("Series.720p/Season 1/S01E01.mkv", 1450000)

	// Релиз 2: папка "Series.1080p.Remux" (другое имя папки, но файл и размер те же!)
	hash2 := GenerateFileHash("Series.1080p.Remux/S01E01.mkv", 1450000)

	// Релиз 3: серия 2 (другой файл)
	hash3 := GenerateFileHash("Series.1080p.Remux/S01E02.mkv", 1450000)

	if hash1 != hash2 {
		t.Errorf("File hashes must match for identical filename and size! got %s != %s", hash1, hash2)
	}

	if hash1 == hash3 {
		t.Errorf("File hashes must differ for different filenames! got %s == %s", hash1, hash3)
	}
}

// TestEngineFullLifecycle тестирует полный цикл: запуск, чтение, сессии, RAM кэш и остановку
func TestEngineFullLifecycle(t *testing.T) {
	// 1. Конфигурируем чистый движок с оперативным кэшем
	cfg := DefaultConfig()
	cfg.ListenPort = 0 // случайный свободный порт
	cfg.Storage = &torrstor.Config{
		Capacity:         10 << 20, // 10 MB кэша
		UseDisk:          false,    // чисто в RAM!
		ConnectionsLimit: 25,
		ReaderReadAHead:  95,
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	// 2. Создаем синтетическую раздачу и запускаем в памяти
	spec := createSyntheticSpec(t)
	title := "My Awesome Test Series"
	userID := "test_user"

	// Теперь движок принимает только спецификацию
	session, err := engine.Start(spec)
	if err != nil {
		t.Fatalf("Failed to start torrent session: %v", err)
	}

	// Имитируем работу менеджера: задаем личные метаданные пользователя
	session.SetUserMeta(userID, title, "", "")

	// 3. Ожидаем загрузки метаданных (для синтетической раздачи это доли миллисекунды)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := session.WaitInfo(ctx); err != nil {
		t.Fatalf("WaitInfo failed: %v", err)
	}

	// 4. Проверяем файлы раздачи и вычисленные хэши
	files := session.Files()
	if len(files) != 2 {
		t.Fatalf("Expected 2 files in session, got: %d", len(files))
	}

	if files[0].Name != "S01E01.mkv" || files[0].FileHash == "" {
		t.Errorf("Invalid file 0 stats: %+v", files[0])
	}
	if files[1].Name != "S01E02.mkv" || files[1].FileHash == "" {
		t.Errorf("Invalid file 1 stats: %+v", files[1])
	}

	// 5. Проверяем статус раздачи для конкретного пользователя
	status := session.Status(userID)
	if status.Title != title {
		t.Errorf("Status title mismatch: got %s, want %s", status.Title, title)
	}
	if status.Stat != TorrentWorking {
		t.Errorf("Expected status TorrentWorking, got: %v", status.Stat)
	}
	if status.TorrentSize != 128*1024 {
		t.Errorf("Torrent size mismatch: got %d, want %d", status.TorrentSize, 128*1024)
	}

	// 6. Проверяем создание ридеров и трекинг активности (ActiveReaders)
	if session.ActiveReaders() != 0 {
		t.Errorf("Expected 0 active readers initially, got: %d", session.ActiveReaders())
	}

	reader, err := session.NewReader(0)
	if err != nil {
		t.Fatalf("Failed to create reader for file 0: %v", err)
	}

	if session.ActiveReaders() != 1 {
		t.Errorf("Expected 1 active reader after NewReader, got: %d", session.ActiveReaders())
	}

	// Закрываем ридер и проверяем сброс счетчика
	session.CloseReader(reader)
	if session.ActiveReaders() != 0 {
		t.Errorf("Expected 0 active readers after CloseReader, got: %d", session.ActiveReaders())
	}

	// 7. Проверяем методы фасадного движка (Get, List)
	if _, ok := engine.Get(spec.InfoHash); !ok {
		t.Errorf("Engine.Get failed to find active session by hash")
	}

	activeSessions := engine.List()
	if len(activeSessions) != 1 {
		t.Errorf("Expected 1 session in Engine.List, got: %d", len(activeSessions))
	}

	// 8. Останавливаем раздачу и проверяем выгрузку из RAM
	engine.Stop(spec.InfoHash)

	if _, ok := engine.Get(spec.InfoHash); ok {
		t.Errorf("Session should be removed from Engine after Stop")
	}

	if len(engine.List()) != 0 {
		t.Errorf("Expected 0 sessions after Stop, got: %d", len(engine.List()))
	}
}

// TestHelpersAndRateLimiter проверяет вспомогательные функции
func TestHelpersAndRateLimiter(t *testing.T) {
	// Лимитер
	lim := NewRateLimiter(1024) // 1 МБ/с
	if lim == nil {
		t.Errorf("Expected non-nil rate limiter")
	}

	// PeerID
	peerID := GeneratePeerID("-VT0001-")
	if len(peerID) != 20 {
		t.Errorf("PeerID must be exactly 20 bytes, got length %d: %s", len(peerID), peerID)
	}
}
