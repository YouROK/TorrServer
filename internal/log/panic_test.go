package log

import (
	"strings"
	"testing"
)

// TestPanicWritesStack проверяет, что паника попадает в лог со стеком:
// по одному тексту ошибки место падения не найти.
func TestPanicWritesStack(t *testing.T) {
	buf := &syncBuffer{}
	mu.Lock()
	prev := defaultLogger
	defaultLogger = newLoggerForTest(buf)
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		defaultLogger = prev
		mu.Unlock()
	})

	func() {
		defer func() {
			if r := recover(); r != nil {
				Panic("[Test] Handler panicked", r)
			}
		}()
		panic("boom")
	}()

	out := buf.String()
	if !strings.Contains(out, "Handler panicked") {
		t.Errorf("log has no panic message:\n%s", out)
	}
	if !strings.Contains(out, "boom") {
		t.Errorf("log has no panic value:\n%s", out)
	}
	if !strings.Contains(out, "panic_test.go") {
		t.Errorf("log has no stack trace pointing to the test file:\n%s", out)
	}
}
