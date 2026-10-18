package waf

import (
	"strings"
	"testing"
)

func TestHostMatchesBlocked(t *testing.T) {
	blocked := []string{"bylampa.online", "example.com"}

	tests := []struct {
		host string
		want bool
	}{
		{"bylampa.online", true},
		{"zerkalo.bylampa.online", true},
		{"www.bylampa.online", true},
		{"notbylampa.online", false},
		{"bylampa.online.evil.com", false},
		{"example.com", true},
		{"sub.example.com", true},
		{"localhost", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := hostMatchesBlocked(tt.host, blocked); got != tt.want {
			t.Errorf("hostMatchesBlocked(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}

func TestIsBlockedReferer(t *testing.T) {
	blocked := []string{"bylampa.online"}

	if host, ok := isBlockedReferer("https://zerkalo.bylampa.online/player", "", blocked); !ok || host != "zerkalo.bylampa.online" {
		t.Fatalf("expected blocked referer, got host=%q ok=%v", host, ok)
	}
	if _, ok := isBlockedReferer("", "https://bylampa.online", blocked); !ok {
		t.Fatal("expected blocked origin")
	}
	if _, ok := isBlockedReferer("", "", blocked); ok {
		t.Fatal("expected empty headers to pass")
	}
	if _, ok := isBlockedReferer("https://myserver.local/stream", "", blocked); ok {
		t.Fatal("expected unrelated referer to pass")
	}
}

func TestBlockedReferersFromConfig(t *testing.T) {
	got := blockedReferersFromConfig([]byte("example.com\nbylampa.online\n"))
	want := append(append([]string(nil), defaultBlockedReferers...), "example.com")
	if len(got) != len(want) {
		t.Fatalf("blockedReferersFromConfig() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("blockedReferersFromConfig()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	empty := blockedReferersFromConfig(nil)
	if len(empty) != len(defaultBlockedReferers) {
		t.Fatalf("blockedReferersFromConfig(nil) = %v, want defaults %v", empty, defaultBlockedReferers)
	}
	for i, host := range defaultBlockedReferers {
		if empty[i] != host {
			t.Fatalf("blockedReferersFromConfig(nil)[%d] = %q, want %q", i, empty[i], host)
		}
	}
}

func TestScanRefererBuf(t *testing.T) {
	got := scanRefererBuf([]byte("# comment\nbylampa.online\n\n# another\nexample.com\n"))
	want := []string{"bylampa.online", "example.com"}
	if len(got) != len(want) {
		t.Fatalf("scanRefererBuf() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("scanRefererBuf()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestNormalizeRefererRule(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{"Example.COM.", "example.com", true},
		{"example.com:443", "example.com", true},
		{"https://Example.COM/", "example.com", true},
		{"http://example.com:8080", "example.com", true},
		{"127.0.0.1", "127.0.0.1", true},
		{"[2001:db8::1]:443", "2001:db8::1", true},
		{"https://example.com/path", "", false},
		{"ftp://example.com", "", false},
		{"*.example.com", "", false},
		{"example.com/path", "", false},
		{"-bad.example", "", false},
		{"bad_.example", "", false},
	}
	for _, tt := range tests {
		got, ok := normalizeRefererRule(tt.input)
		if got != tt.want || ok != tt.ok {
			t.Errorf("normalizeRefererRule(%q) = %q, %v; want %q, %v", tt.input, got, ok, tt.want, tt.ok)
		}
	}
}

func TestScanRefererBufWarnings(t *testing.T) {
	hosts, warnings := scanRefererBufValidated([]byte("example.com\n*.bad.example\nhttps://good.example/\n"))
	if len(hosts) != 2 || hosts[0] != "example.com" || hosts[1] != "good.example" {
		t.Fatalf("hosts=%v", hosts)
	}
	if len(warnings) != 1 || warnings[0].Line != 2 || warnings[0].Code != "invalid_referer" {
		t.Fatalf("warnings=%+v", warnings)
	}
}

func TestHostFromHeaderNormalizesPortAndTrailingDot(t *testing.T) {
	if got := hostFromHeader("https://Example.COM.:443/path"); got != "example.com" {
		t.Fatalf("hostFromHeader()=%q", got)
	}
}

func TestRefererScannerReportsOversizedLine(t *testing.T) {
	_, warnings := scanRefererBufValidated([]byte(strings.Repeat("a", maxScannerTokenSize+1)))
	if len(warnings) != 1 || warnings[0].Code != "line_too_long" {
		t.Fatalf("warnings=%+v", warnings)
	}
}
