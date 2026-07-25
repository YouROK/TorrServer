package netbind

import (
	"net"
	"strconv"
	"testing"
)

func TestNormalizeEmpty(t *testing.T) {
	got := Normalize(nil)
	if len(got) != 1 || got[0] != "" {
		t.Fatalf("Normalize(nil) = %#v, want [\"\"]", got)
	}
}

func TestNormalizeDedup(t *testing.T) {
	got := Normalize([]string{"127.0.0.1", "127.0.0.1", "::1"})
	if len(got) != 2 {
		t.Fatalf("Normalize dedup = %#v, want 2 entries", got)
	}
}

func TestAddr(t *testing.T) {
	if got := Addr("", "8090"); got != ":8090" {
		t.Fatalf("Addr empty host = %q", got)
	}
	if got := Addr("127.0.0.1", "8090"); got != "127.0.0.1:8090" {
		t.Fatalf("Addr ipv4 = %q", got)
	}
	if got := Addr("::1", "8090"); got != "[::1]:8090" {
		t.Fatalf("Addr ipv6 = %q", got)
	}
}

func TestListenMultiple(t *testing.T) {
	l, err := Listen([]string{"127.0.0.1"}, "0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	port := l.Addr().(*net.TCPAddr).Port
	conn, err := net.Dial("tcp", Addr("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}
