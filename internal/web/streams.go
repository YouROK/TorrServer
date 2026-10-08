package web

import (
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// StreamSnapshot - снимок одного активного HTTP-потока раздачи.
// Один плеер может держать несколько одновременных соединений на один файл
// (перемотка, параллельные range-запросы), поэтому потоков на пару
// (пользователь, раздача) может быть больше одного.
type StreamSnapshot struct {
	ID          int64   `json:"id"`
	UserID      string  `json:"user_id"`
	Hash        string  `json:"hash"`
	FileIdx     int     `json:"file_idx"`
	FileName    string  `json:"file_name"`
	ClientIP    string  `json:"client_ip"`
	ForwardedIP string  `json:"forwarded_ip,omitempty"`
	UserAgent   string  `json:"user_agent,omitempty"`
	StartedAt   int64   `json:"started_at"`
	DurationSec int64   `json:"duration_sec"`
	Bytes       int64   `json:"bytes"`
	SpeedBps    float64 `json:"speed_bps"`
}

// StreamOpen описывает открываемый поток.
type StreamOpen struct {
	UserID      string
	Hash        string
	FileIdx     int
	FileName    string
	ClientIP    string
	ForwardedIP string
	UserAgent   string
}

type streamEntry struct {
	open      StreamOpen
	startedAt time.Time
	bytes     int64

	// Для расчёта мгновенной скорости между опросами админки
	lastBytes int64
	lastAt    time.Time
}

// StreamTracker учитывает активные HTTP-потоки: кто, что и с какого адреса смотрит.
// Позволяет админке видеть, сколько потоков отдаётся пользователю и с каких IP,
// что помогает заметить передачу API-токена третьим лицам.
type StreamTracker struct {
	mu      sync.Mutex
	nextID  int64
	streams map[int64]*streamEntry

	// now подменяется в тестах для детерминированного времени.
	now func() time.Time
}

func NewStreamTracker() *StreamTracker {
	return &StreamTracker{
		streams: make(map[int64]*streamEntry),
		now:     time.Now,
	}
}

// StreamHandle - ссылка на зарегистрированный поток.
type StreamHandle struct {
	tracker *StreamTracker
	id      int64
}

// Open регистрирует новый поток.
func (t *StreamTracker) Open(o StreamOpen) *StreamHandle {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.nextID++
	id := t.nextID
	now := t.now()
	t.streams[id] = &streamEntry{
		open:      o,
		startedAt: now,
		lastAt:    now,
	}
	return &StreamHandle{tracker: t, id: id}
}

func (h *StreamHandle) addBytes(n int64) {
	if h == nil || n <= 0 {
		return
	}
	t := h.tracker
	t.mu.Lock()
	defer t.mu.Unlock()
	if e, ok := t.streams[h.id]; ok {
		e.bytes += n
	}
}

// Close снимает поток с учёта. Безопасен для повторного вызова.
func (h *StreamHandle) Close() {
	if h == nil {
		return
	}
	t := h.tracker
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.streams, h.id)
}

// Idle снимает потоки, которые открыты дольше ttl и не отдали ни байта.
// Обычно хватает Close из defer, но если соединение зависло и обработчик
// не завершился, поток иначе остался бы в списке навсегда.
func (t *StreamTracker) Idle(ttl time.Duration) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	removed := 0
	for id, e := range t.streams {
		if e.bytes == 0 && now.Sub(e.startedAt) > ttl {
			delete(t.streams, id)
			removed++
		}
	}
	return removed
}

// toSnapshot обновляет оценку скорости и строит снимок потока.
// Скорость считается по приросту байт между последовательными опросами админки.
func (t *StreamTracker) toSnapshot(id int64, e *streamEntry) StreamSnapshot {
	now := t.now()

	elapsed := now.Sub(e.lastAt).Seconds()
	delta := e.bytes - e.lastBytes
	e.lastAt = now
	e.lastBytes = e.bytes

	speed := 0.0
	if elapsed > 0 && delta > 0 {
		speed = float64(delta) / elapsed
	}

	total := now.Sub(e.startedAt).Seconds()
	if total < 0 {
		total = 0
	}

	return StreamSnapshot{
		ID:          id,
		UserID:      e.open.UserID,
		Hash:        e.open.Hash,
		FileIdx:     e.open.FileIdx,
		FileName:    e.open.FileName,
		ClientIP:    e.open.ClientIP,
		ForwardedIP: e.open.ForwardedIP,
		UserAgent:   e.open.UserAgent,
		StartedAt:   e.startedAt.Unix(),
		DurationSec: int64(total),
		Bytes:       e.bytes,
		SpeedBps:    speed,
	}
}

// ForUser возвращает активные потоки пользователя, сгруппированные по хэшу раздачи.
func (t *StreamTracker) ForUser(userID string) map[string][]StreamSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	out := make(map[string][]StreamSnapshot)
	for id, e := range t.streams {
		if e.open.UserID != userID {
			continue
		}
		out[e.open.Hash] = append(out[e.open.Hash], t.toSnapshot(id, e))
	}

	for hash := range out {
		list := out[hash]
		sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
		out[hash] = list
	}
	return out
}

// Count возвращает общее число активных потоков.
func (t *StreamTracker) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.streams)
}

// countingWriter считает байты, реально отданные клиенту.
// Встраивание gin.ResponseWriter сохраняет Flusher/Hijacker и прочие методы,
// переопределяется только Write.
type countingWriter struct {
	gin.ResponseWriter
	handle *StreamHandle
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	w.handle.addBytes(int64(n))
	return n, err
}
