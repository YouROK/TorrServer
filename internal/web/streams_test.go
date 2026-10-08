package web

import (
	"testing"
	"time"
)

// Трекер обязан считать потоки по пользователю и группировать их по раздаче.
func TestStreamTrackerGroupsByUserAndHash(t *testing.T) {
	tr := NewStreamTracker()
	now := time.Unix(1_700_000_000, 0)
	tr.now = func() time.Time { return now }

	h1 := tr.Open(StreamOpen{UserID: "u_1", Hash: "aaa", FileIdx: 0, FileName: "S01E01.mkv", ClientIP: "10.0.0.1"})
	h2 := tr.Open(StreamOpen{UserID: "u_1", Hash: "aaa", FileIdx: 0, FileName: "S01E01.mkv", ClientIP: "10.0.0.2"})
	h3 := tr.Open(StreamOpen{UserID: "u_1", Hash: "bbb", FileIdx: 3, FileName: "movie.mp4", ClientIP: "10.0.0.1"})
	tr.Open(StreamOpen{UserID: "u_2", Hash: "aaa", FileIdx: 0, ClientIP: "10.0.0.9"})

	if got := tr.Count(); got != 4 {
		t.Fatalf("Count() = %d, want 4", got)
	}

	forUser := tr.ForUser("u_1")
	if len(forUser) != 2 {
		t.Fatalf("ForUser(u_1) returned %d hashes, want 2", len(forUser))
	}
	if len(forUser["aaa"]) != 2 {
		t.Errorf("hash aaa: got %d streams, want 2", len(forUser["aaa"]))
	}
	if len(forUser["bbb"]) != 1 {
		t.Errorf("hash bbb: got %d streams, want 1", len(forUser["bbb"]))
	}

	// Чужие потоки не должны попадать в выборку
	for _, list := range forUser {
		for _, st := range list {
			if st.UserID != "u_1" {
				t.Errorf("ForUser(u_1) leaked stream of %q", st.UserID)
			}
		}
	}

	// Close убирает только свой поток
	h1.Close()
	h2.Close()
	h3.Close()
	if got := tr.Count(); got != 1 {
		t.Errorf("Count() after close = %d, want 1", got)
	}
	if got := tr.ForUser("u_1"); len(got) != 0 {
		t.Errorf("ForUser(u_1) after close = %v, want empty", got)
	}
}

// Повторный Close не должен паниковать и ломать учёт других потоков.
func TestStreamTrackerDoubleCloseIsSafe(t *testing.T) {
	tr := NewStreamTracker()
	h := tr.Open(StreamOpen{UserID: "u_1", Hash: "aaa"})
	other := tr.Open(StreamOpen{UserID: "u_1", Hash: "aaa"})

	h.Close()
	h.Close()

	if got := tr.Count(); got != 1 {
		t.Errorf("Count() = %d, want 1 (only the other stream)", got)
	}
	other.Close()
	if got := tr.Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
}

// Скорость считается по приросту байт между опросами.
func TestStreamTrackerSpeedAndDuration(t *testing.T) {
	tr := NewStreamTracker()
	now := time.Unix(1_700_000_000, 0)
	tr.now = func() time.Time { return now }

	h := tr.Open(StreamOpen{UserID: "u_1", Hash: "aaa", ClientIP: "10.0.0.1"})

	// Первый опрос: байт ещё нет, скорость нулевая
	first := tr.ForUser("u_1")["aaa"]
	if len(first) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(first))
	}
	if first[0].SpeedBps != 0 {
		t.Errorf("initial speed = %v, want 0", first[0].SpeedBps)
	}

	// Прошло 2 секунды, отдали 4 МиБ → ~2 МиБ/с
	now = now.Add(2 * time.Second)
	h.addBytes(4 * 1024 * 1024)

	second := tr.ForUser("u_1")["aaa"][0]
	wantSpeed := float64(4*1024*1024) / 2.0
	if diff := second.SpeedBps - wantSpeed; diff > 1 || diff < -1 {
		t.Errorf("speed = %v, want ~%v", second.SpeedBps, wantSpeed)
	}
	if second.Bytes != 4*1024*1024 {
		t.Errorf("bytes = %d, want %d", second.Bytes, 4*1024*1024)
	}
	if second.DurationSec != 2 {
		t.Errorf("duration = %d, want 2", second.DurationSec)
	}

	// Скорость должна затухать, когда трафик прекратился
	now = now.Add(10 * time.Second)
	third := tr.ForUser("u_1")["aaa"][0]
	if third.SpeedBps != 0 {
		t.Errorf("speed after idle = %v, want 0", third.SpeedBps)
	}

	h.Close()
}

// Уборщик снимает только «пустые» зависшие потоки и не трогает активные.
func TestStreamTrackerIdleCleanup(t *testing.T) {
	tr := NewStreamTracker()
	now := time.Unix(1_700_000_000, 0)
	tr.now = func() time.Time { return now }

	stale := tr.Open(StreamOpen{UserID: "u_1", Hash: "aaa"})
	active := tr.Open(StreamOpen{UserID: "u_1", Hash: "bbb"})
	active.addBytes(1024)

	now = now.Add(time.Hour)

	removed := tr.Idle(30 * time.Minute)
	if removed != 1 {
		t.Errorf("Idle() removed %d, want 1", removed)
	}

	forUser := tr.ForUser("u_1")
	if len(forUser["aaa"]) != 0 {
		t.Errorf("stale stream must be removed")
	}
	if len(forUser["bbb"]) != 1 {
		t.Errorf("active stream must survive")
	}

	stale.Close()
	active.Close()
}

// IPv6-адреса и адрес из заголовка прокси должны попадать в снимок как есть.
func TestStreamTrackerKeepsClientAndForwardedIP(t *testing.T) {
	tr := NewStreamTracker()
	h := tr.Open(StreamOpen{
		UserID:      "u_1",
		Hash:        "aaa",
		ClientIP:    "192.168.1.50",
		ForwardedIP: "203.0.113.7",
	})
	defer h.Close()

	st := tr.ForUser("u_1")["aaa"][0]
	if st.ClientIP != "192.168.1.50" {
		t.Errorf("ClientIP = %q", st.ClientIP)
	}
	if st.ForwardedIP != "203.0.113.7" {
		t.Errorf("ForwardedIP = %q", st.ForwardedIP)
	}
}
