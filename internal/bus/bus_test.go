package bus

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestPublishSurvivesPanic проверяет, что паника подписчика не роняет процесс
// и не мешает остальным подписчикам получить событие.
func TestPublishSurvivesPanic(t *testing.T) {
	client := Get("panic_test")
	t.Cleanup(func() { client.UnsubscribeAll() })

	var delivered atomic.Int32

	client.On("test:event", func(payload any) {
		panic("subscriber failed")
	})
	client.On("test:event", func(payload any) {
		delivered.Add(1)
	})

	client.Emit("test:event", "payload")

	deadline := time.Now().Add(2 * time.Second)
	for delivered.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if delivered.Load() != 1 {
		t.Errorf("healthy subscriber got %d events, want 1", delivered.Load())
	}
}
