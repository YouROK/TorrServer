package web

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	_ "server/docs"
	"server/settings"
)

func TestSwaggerAndEcho(t *testing.T) {
	gin.SetMode(gin.TestMode)
	doc, err := filteredSwaggerJSON()
	if err != nil || !strings.Contains(doc, "paths") {
		t.Fatalf("doc %v %s", err, doc)
	}
	r := gin.New()
	r.GET("/echo", echo)
	r.GET("/swagger/*any", swaggerHandler())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/echo", nil))
	if w.Code != http.StatusOK || w.Body.Len() == 0 {
		t.Fatalf("echo %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "paths") {
		t.Fatalf("swagger %d %s", w.Code, w.Body.String())
	}
	_ = GetLocalIps()

	oldList, oldAddrs := listIfaces, ifaceAddrs
	t.Cleanup(func() { listIfaces, ifaceAddrs = oldList, oldAddrs })
	en := net.Interface{Name: "en0", Flags: net.FlagUp}
	listIfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "lo0", Flags: net.FlagUp | net.FlagLoopback},
			en,
			{Name: "down0"},
		}, nil
	}
	ifaceAddrs = func(i *net.Interface) ([]net.Addr, error) {
		if i.Name != "en0" {
			return nil, nil
		}
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)},
			&net.IPNet{IP: net.ParseIP("169.254.1.1"), Mask: net.CIDRMask(16, 32)},
			&net.IPNet{IP: net.ParseIP("192.0.2.10"), Mask: net.CIDRMask(24, 32)},
			&net.IPAddr{IP: net.ParseIP("2001:db8::1")},
		}, nil
	}
	ips := GetLocalIps()
	if len(ips) != 2 {
		t.Fatalf("ips %#v", ips)
	}
	oldIPs := settings.IPs
	t.Cleanup(func() { settings.IPs = oldIPs })
	settings.IPs = []string{""}
	if len(certIPs()) != 2 {
		t.Fatal(certIPs())
	}
	settings.IPs = []string{"192.0.2.10"}
	if got := certIPs(); len(got) != 1 || got[0] != "192.0.2.10" {
		t.Fatalf("bound %#v", got)
	}
}
