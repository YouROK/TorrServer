package log

import (
	"strings"
	"testing"
)

// TestSanitizeLine проверяет, что переводы строк не подделывают строки лога.
func TestSanitizeLine(t *testing.T) {
	got := sanitizeLine("[Web] Failed login for user 'evil\n[User] Deleted user admin'")
	if strings.ContainsAny(got, "\r\n") {
		t.Errorf("sanitizeLine left line breaks: %q", got)
	}

	plain := "[Web] Server started"
	if sanitizeLine(plain) != plain {
		t.Errorf("sanitizeLine changed a clean message: %q", sanitizeLine(plain))
	}
}
