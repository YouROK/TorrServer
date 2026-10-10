package web

import (
	"net"
	"testing"

	"silo/internal/config"
	"silo/internal/ffmpeg/source"
)

// newAccessTestServer создаёт сервер с заданным адресом привязки.
func newAccessTestServer(host string) *Server {
	return &Server{
		cfg: &config.Config{
			Server: config.ServerConfig{Host: host, Port: 8090},
		},
		sourceRegistry: source.NewRegistry(0),
	}
}

func TestIsLocalRequestLoopback(t *testing.T) {
	s := newAccessTestServer("0.0.0.0")

	cases := []struct {
		remote string
		want   bool
	}{
		{remote: "127.0.0.1", want: true},
		{remote: "::1", want: true},
		{remote: "127.0.0.5", want: true},
		{remote: "192.168.1.50", want: false},
		{remote: "10.0.0.7", want: false},
		{remote: "203.0.113.9", want: false},
		{remote: "", want: false},
		{remote: "not-an-ip", want: false},
	}

	for _, c := range cases {
		t.Run(c.remote, func(t *testing.T) {
			if got := s.isLocalRequest(newTestContext(c.remote)); got != c.want {
				t.Errorf("isLocalRequest(%q) = %v, want %v", c.remote, got, c.want)
			}
		})
	}
}

func TestIsLocalRequestBindAddress(t *testing.T) {
	// При привязке к конкретному адресу ffmpeg обращается именно к нему,
	// потому что loopback в этом случае недоступен.
	s := newAccessTestServer("192.168.1.11")

	if !s.isLocalRequest(newTestContext("192.168.1.11")) {
		t.Error("request from the bind address must be local")
	}
	if s.isLocalRequest(newTestContext("192.168.1.12")) {
		t.Error("request from another address must not be local")
	}
	if s.isLocalRequest(newTestContext("203.0.113.9")) {
		t.Error("request from an external address must not be local")
	}
}

func TestIsLocalRequestUnspecifiedBind(t *testing.T) {
	// При привязке к 0.0.0.0 адрес клиента из локальной сети совпадает
	// с адресом сервера, поэтому доверяем только настоящему loopback.
	for _, host := range []string{"0.0.0.0", "::", ""} {
		t.Run(host, func(t *testing.T) {
			s := newAccessTestServer(host)
			if s.bindIP() != nil {
				t.Errorf("bindIP for %q must be nil", host)
			}
			if !s.isLocalRequest(newTestContext("127.0.0.1")) {
				t.Error("loopback must stay local")
			}
			if s.isLocalRequest(newTestContext("192.168.1.50")) {
				t.Error("LAN address must not be local when bound to all interfaces")
			}
		})
	}
}

func TestBindIPv6(t *testing.T) {
	s := newAccessTestServer("[::1]")
	ip := s.bindIP()
	if ip == nil || !ip.Equal(net.ParseIP("::1")) {
		t.Errorf("bindIP = %v, want ::1", ip)
	}
}

func TestDefaultSourceHost(t *testing.T) {
	cases := []struct {
		host string
		want string
	}{
		{host: "0.0.0.0", want: "127.0.0.1"},
		{host: "", want: "127.0.0.1"},
		{host: "192.168.1.11", want: "192.168.1.11"},
	}

	for _, c := range cases {
		t.Run(c.host, func(t *testing.T) {
			s := newAccessTestServer(c.host)
			if got := s.defaultSourceHost(); got != c.want {
				t.Errorf("defaultSourceHost() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestSourceURL(t *testing.T) {
	s := newAccessTestServer("0.0.0.0")

	url := s.SourceURL("abc123")
	want := "http://127.0.0.1:8090/api/internal/source/abc123"
	if url != want {
		t.Errorf("SourceURL = %q, want %q", url, want)
	}
}

func TestSourceURLUsesActualPort(t *testing.T) {
	s := newAccessTestServer("0.0.0.0")
	s.setSourceAddr(&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 45678})

	url := s.SourceURL("id1")
	want := "http://127.0.0.1:45678/api/internal/source/id1"
	if url != want {
		t.Errorf("SourceURL = %q, want %q", url, want)
	}
}

func TestSetSourceAddrWithBindAddress(t *testing.T) {
	s := newAccessTestServer("192.168.1.11")
	s.setSourceAddr(&net.TCPAddr{IP: net.ParseIP("192.168.1.11"), Port: 9000})

	url := s.SourceURL("id2")
	want := "http://192.168.1.11:9000/api/internal/source/id2"
	if url != want {
		t.Errorf("SourceURL = %q, want %q", url, want)
	}
}

func TestSetSourceAddrDualStack(t *testing.T) {
	// Go слушает 0.0.0.0 как dual-stack и возвращает адрес [::]
	s := newAccessTestServer("0.0.0.0")
	s.setSourceAddr(&net.TCPAddr{IP: net.IPv6unspecified, Port: 8080})

	url := s.SourceURL("id3")
	want := "http://127.0.0.1:8080/api/internal/source/id3"
	if url != want {
		t.Errorf("SourceURL = %q, want %q", url, want)
	}
}

func TestSourceRegistryIssuesLeases(t *testing.T) {
	s := newAccessTestServer("0.0.0.0")

	lease := s.sourceRegistry.Issue("owner", "aabb", 1)
	if lease.ID == "" {
		t.Fatal("lease id is empty")
	}

	got, ok := s.sourceRegistry.Get(lease.ID)
	if !ok {
		t.Fatal("issued lease was not found")
	}
	if got.Hash != "aabb" || got.FileIdx != 1 {
		t.Errorf("lease = %+v, want hash aabb index 1", got)
	}
}
