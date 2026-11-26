package torr

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

// TestWriteStatusNilClient — ДЕТЕРМИНИРОВАННЫЙ контроль на nil-указателе.
// Состояние bts.client == nil реально возникает на проде между Disconnect()
// (btserver.go:81) и Connect() (btserver.go:69) во время SetSettings.
//
// ДО фикса: паника nil pointer на apihelper.go:326 → torrent client.go:1461.
// ПОСЛЕ фикса: паники нет, в вывод попадает сообщение о неподключённом клиенте.
func TestWriteStatusNilClient(t *testing.T) {
	bt := NewBTS()
	InitApiHelper(bt)
	bt.client = nil

	var buf bytes.Buffer
	WriteStatus(&buf)
	if !strings.Contains(buf.String(), "not connected") {
		t.Fatalf("ожидалось сообщение о неподключённом клиенте, получено: %q", buf.String())
	}
}

// TestNewTorrentNilClient — второе место чтения bt.client (torrent.go).
// Проверка nil там есть, но она была за 17 строк до использования, то есть
// не защищала от чересполосицы Disconnect(). Тест фиксирует авторитетную
// проверку непосредственно перед вызовом.
func TestNewTorrentNilClient(t *testing.T) {
	bt := NewBTS()
	InitApiHelper(bt)
	bt.client = nil

	if _, err := NewTorrent(&torrent.TorrentSpec{}, bt); err == nil {
		t.Fatal("ожидалась ошибка при nil-клиенте, получено nil")
	}
}

// TestConcurrentDisconnectAndWriteStatus — настоящая гонка: переподключение
// клиента против читателей /stat.
//
// Без -race (нужен CGO/gcc, на этой машине недоступен) результат
// недетерминирован, поэтому тест проверяет главное — отсутствие паники и
// завершение. Реальную гонку данных ловит -race в CI на Linux.
func TestConcurrentDisconnectAndWriteStatus(t *testing.T) {
	// configure() читает settings.BTsets напрямую (btserver.go:90), а
	// configureProxy() — settings.Args (btserver.go:202). Оба глобала заполняет
	// main() при старте, поэтому тест обязан их выставить, иначе Connect() падает
	// в nil — это дефект окружения теста, а не проверяемый баг. Так же поступает
	// rutor/race_test.go.
	initTestGlobals(t)

	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer bt.Disconnect()

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			bt.Disconnect()
			_ = bt.Connect()
		}
		close(stop)
	}()

	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var buf bytes.Buffer
			for {
				select {
				case <-stop:
					return
				default:
				}
				buf.Reset()
				WriteStatus(&buf)
			}
		}()
	}
	wg.Wait()
}

// initTestGlobals выставляет глобалы, которые заполняет main() при старте:
// configure() читает settings.BTsets (btserver.go:90), configureProxy() —
// settings.Args (btserver.go:202). Без них Connect() падает в nil, и это дефект
// окружения теста, а не проверяемый баг. Так же поступает rutor/race_test.go.
//
// Глобалы НЕ возвращаются в nil после теста: горутины torrent-клиента переживают
// Disconnect(), и обнуление settings.BTsets в defer давало ложное срабатывание
// -race. Для остальных тестов пакета оставленные глобалы безвредны.
func initTestGlobals(t *testing.T) {
	t.Helper()
	if settings.BTsets == nil {
		settings.BTsets = &settings.BTSets{}
	}
	if settings.Args == nil {
		settings.Args = &settings.ExecArgs{}
	}
}

// TestNewTorrentDoesNotDeadlock фиксирует требование к дисциплине блокировок в
// NewTorrent: чтение bt.client берёт bt.mu.RLock, а код ниже берёт bt.mu.Lock()
// для карты торрентов. Если read-lock удерживается через defer, RWMutex
// блокируется сам на себе и КАЖДОЕ добавление торрента намертво зависает.
//
// Тест обязан проверять успешный путь AddTorrentSpec: при ошибке мы выходим
// до bt.mu.Lock() и дедлок не воспроизводится, поэтому спецификация строится
// из настоящего metainfo.
func TestNewTorrentDoesNotDeadlock(t *testing.T) {
	initTestGlobals(t)

	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer bt.Disconnect()

	info := metainfo.Info{
		Name:        "deadlock-probe",
		PieceLength: 1 << 10,
		Length:      1 << 10,
		Pieces:      make([]byte, 20),
	}
	mi := metainfo.MetaInfo{InfoBytes: bencode.MustMarshal(info)}
	// так же собирает сам клиент в spec.go из MetaInfo
	spec := &torrent.TorrentSpec{
		InfoBytes:   mi.InfoBytes,
		InfoHash:    mi.HashInfoBytes(),
		DisplayName: info.Name,
	}

	done := make(chan error, 1)
	go func() {
		_, err := NewTorrent(spec, bt)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			// спецификация могла не приняться — это ослабляет тест, поэтому
			// сообщаем явно, а не молча считаем проверку успешной
			t.Logf("ВНИМАНИЕ: AddTorrentSpec отверг спецификацию (%v), путь до bt.mu.Lock() не выполнен", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("NewTorrent не вернулся за 30 c: bt.mu.RLock удерживается через defer, а ниже берётся bt.mu.Lock() — дедлок")
	}
}
