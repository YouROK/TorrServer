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

type fakeAnnouncer struct{}

func (fakeAnnouncer) Shutdown() {}

func TestStartStopConcurrent(t *testing.T) {
	oldNew := newAnnouncer
	newAnnouncer = func(string, string, int, string, []string, []string, []net.Interface) (announcer, error) {
		return fakeAnnouncer{}, nil
	}
	t.Cleanup(func() { newAnnouncer = oldNew })

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

func TestAdvertiseMockInterfaces(t *testing.T) {
	oldList, oldAddrs, oldNew := listIfaces, addrsByIface, newAnnouncer
	oldPort, oldSsl, oldSslPort, oldSets := settings.Port, settings.Ssl, settings.SslPort, settings.BTsets
	var registered []string
	newAnnouncer = func(_, service string, _ int, _ string, _, _ []string, _ []net.Interface) (announcer, error) {
		registered = append(registered, service)
		return fakeAnnouncer{}, nil
	}
	t.Cleanup(func() {
		listIfaces, addrsByIface, newAnnouncer = oldList, oldAddrs, oldNew
		settings.Port, settings.Ssl, settings.SslPort, settings.BTsets = oldPort, oldSsl, oldSslPort, oldSets
		Stop()
	})

	en := net.Interface{Index: 2, Name: "en0", Flags: net.FlagUp | net.FlagMulticast}
	listIfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Index: 1, Name: "lo0", Flags: net.FlagUp | net.FlagLoopback | net.FlagMulticast},
			{Index: 3, Name: "utun0", Flags: net.FlagUp | net.FlagMulticast},
			en,
		}, nil
	}
	addrsByIface = func(i *net.Interface) ([]net.Addr, error) {
		if i.Name != "en0" {
			return nil, nil
		}
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)},
			&net.IPNet{IP: net.ParseIP("169.254.1.1"), Mask: net.CIDRMask(16, 32)},
			&net.IPNet{IP: net.ParseIP("192.0.2.10"), Mask: net.CIDRMask(24, 32)},
			&net.IPNet{IP: net.ParseIP("2001:db8::1"), Mask: net.CIDRMask(32, 128)},
			&net.IPAddr{IP: net.ParseIP("192.0.2.11")},
		}, nil
	}
	ifaces, ips := advertiseAddrs()
	if len(ifaces) != 1 || ifaces[0].Name != "en0" {
		t.Fatalf("ifaces %#v", ifaces)
	}
	if len(ips) != 3 {
		t.Fatalf("ips %#v", ips)
	}

	listIfaces = func() ([]net.Interface, error) { return nil, errMock }
	if ifs, ips := advertiseAddrs(); ifs != nil || ips != nil {
		t.Fatalf("error path %#v %#v", ifs, ips)
	}
	listIfaces = func() ([]net.Interface, error) { return []net.Interface{en}, nil }

	settings.Port = "8090"
	settings.Ssl = true
	settings.SslPort = "8091"
	settings.BTsets = &settings.BTSets{FriendlyName: "Living Room.local"}
	Start()
	if want := []string{serviceTorrServer, serviceHTTP, serviceHTTPS}; !slices.Equal(registered, want) {
		t.Fatalf("registered %v, want %v", registered, want)
	}
	Stop()
	if servers != nil {
		t.Fatal("servers left registered")
	}
}

var errMock = errString("mock")

type errString string

func (e errString) Error() string { return string(e) }
