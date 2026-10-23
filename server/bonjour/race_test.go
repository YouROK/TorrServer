package bonjour

import (
	"net"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"

	"server/settings"
)

func TestSanitizeInstanceBounds(t *testing.T) {
	if got := sanitizeInstance(strings.Repeat("a", 80)); len(got) != 63 || strings.TrimSpace(got) != got {
		t.Fatalf("long name %q len %d", got, len(got))
	}
	// stripLocalSuffix eats a trailing dot, so "..." collapses to "." then "-".
	if got := sanitizeInstance("..."); got != "-" {
		t.Fatalf("dots = %q", got)
	}
	if got := sanitizeInstance(" .local "); got != "TorrServer" {
		t.Fatalf("blank .local = %q", got)
	}
	if got := sanitizeInstance("  Name.local.  "); got != "Name" {
		t.Fatalf("suffix = %q", got)
	}
}

func TestUsableIface(t *testing.T) {
	up := net.FlagUp | net.FlagMulticast
	cases := []struct {
		name  string
		flags net.Flags
		want  bool
	}{
		{"en0", up, true},
		{"lo0", up | net.FlagLoopback, runtime.GOOS == "windows"},
		{"en1", net.FlagMulticast, runtime.GOOS == "windows"},
		{"utun0", up, false},
		{"docker0", up, false},
		{"br-1", up, false},
		{"veth0", up, false},
		{"awdl0", up, false},
	}
	for _, tc := range cases {
		got := usableIface(net.Interface{Name: tc.name, Flags: tc.flags})
		if got != tc.want {
			t.Errorf("usableIface(%s) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestMdnsHostname(t *testing.T) {
	host := mdnsHostname()
	if host == "" || strings.Contains(host, ".") || strings.Contains(host, " ") {
		t.Fatalf("mdnsHostname() = %q", host)
	}
}

func TestBaseTXT(t *testing.T) {
	oldSsl, oldPort := settings.Ssl, settings.SslPort
	t.Cleanup(func() { settings.Ssl, settings.SslPort = oldSsl, oldPort })
	settings.Ssl, settings.SslPort = true, "8091"
	txt := baseTXT()
	if !slices.Contains(txt, "https=8091") || !slices.Contains(txt, "path=/") {
		t.Fatalf("txt = %#v", txt)
	}
	settings.Ssl = false
	if slices.Contains(baseTXT(), "https=8091") {
		t.Fatal("https txt without ssl")
	}
}

func TestStartStopConcurrent(t *testing.T) {
	oldPort, oldSsl, oldSslPort, oldSets := settings.Port, settings.Ssl, settings.SslPort, settings.BTsets
	settings.Port = "8090"
	settings.Ssl = true
	settings.SslPort = "8091"
	settings.BTsets = &settings.BTSets{FriendlyName: "Race Test.local"}
	t.Cleanup(func() {
		Stop()
		settings.Port, settings.Ssl, settings.SslPort, settings.BTsets = oldPort, oldSsl, oldSslPort, oldSets
	})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			Start()
		}()
		go func() {
			defer wg.Done()
			Stop()
		}()
	}
	wg.Wait()
	Stop()
	if servers != nil {
		t.Fatalf("servers = %#v after Stop", servers)
	}
}
