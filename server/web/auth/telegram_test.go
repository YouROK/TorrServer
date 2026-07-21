package auth

import (
	"testing"
)

func TestExtractJSONInt64Field(t *testing.T) {
	got := extractJSONInt64Field(`{"id":123456789,"first_name":"A"}`, "id")
	if got != 123456789 {
		t.Fatalf("got %d", got)
	}
}

func TestValidateTelegramInitDataRejectsEmpty(t *testing.T) {
	if _, ok := ValidateTelegramInitData("", "token"); ok {
		t.Fatal("expected reject")
	}
	if _, ok := ValidateTelegramInitData("hash=abc", ""); ok {
		t.Fatal("expected reject")
	}
}
