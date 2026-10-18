package waf

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"server/settings"

	"github.com/gin-gonic/gin"
)

func withTestWAFDB(t *testing.T, fn func(dir string)) {
	t.Helper()
	dir := t.TempDir()
	oldPath := settings.Path
	oldReadOnly := settings.ReadOnly
	settings.Path = dir
	settings.ReadOnly = false
	t.Cleanup(func() {
		settings.Path = oldPath
		settings.ReadOnly = oldReadOnly
	})
	fn(dir)
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		line      string
		wantOK    bool
		wantErr   bool
		wantFirst string
		wantLast  string
		wantDesc  string
	}{
		{"", false, false, "", "", ""},
		{"# comment", false, false, "", "", ""},
		{"127.0.0.1", true, false, "127.0.0.1", "127.0.0.1", ""},
		{"127.0.0.0-127.0.0.255", true, false, "127.0.0.0", "127.0.0.255", ""},
		{"local:127.0.0.1", true, false, "127.0.0.1", "127.0.0.1", "local"},
		{"local:127.0.0.0-127.0.0.255", true, false, "127.0.0.0", "127.0.0.255", "local"},
		{"10.0.0.0/8", true, false, "10.0.0.0", "10.255.255.255", ""},
		{"lan:10.0.0.0/8", true, false, "10.0.0.0", "10.255.255.255", "lan"},
		{"2001:db8::1", true, false, "2001:db8::1", "2001:db8::1", ""},
		{"local:2001:db8::1", true, false, "2001:db8::1", "2001:db8::1", "local"},
		{"2001:db8::1-2001:db8::ff", true, false, "2001:db8::1", "2001:db8::ff", ""},
		{"2001:db8::/32", true, false, "2001:db8::", "2001:db8:ffff:ffff:ffff:ffff:ffff:ffff", ""},
		{"not-an-ip", false, true, "", "", ""},
		{"bad:not-an-ip", false, true, "", "", ""},
	}

	for _, tt := range tests {
		r, ok, err := parseLine([]byte(tt.line))
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseLine(%q) expected error", tt.line)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseLine(%q) unexpected error: %v", tt.line, err)
			continue
		}
		if ok != tt.wantOK {
			t.Errorf("parseLine(%q) ok=%v want %v", tt.line, ok, tt.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if r.First.String() != tt.wantFirst || r.Last.String() != tt.wantLast {
			t.Errorf("parseLine(%q) = %s-%s, want %s-%s", tt.line, r.First, r.Last, tt.wantFirst, tt.wantLast)
		}
		if r.Description != tt.wantDesc {
			t.Errorf("parseLine(%q) desc=%q want %q", tt.line, r.Description, tt.wantDesc)
		}
	}
}

func TestScanBufFailSoft(t *testing.T) {
	buf := []byte("# comment\n127.0.0.1\nbad-line\n10.0.0.1\n")
	list, warnings := scanBuf(buf)
	if list.NumRanges() != 2 {
		t.Fatalf("NumRanges=%d want 2", list.NumRanges())
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings=%v want 1", warnings)
	}
	if _, ok := list.Lookup(net.ParseIP("127.0.0.1")); !ok {
		t.Fatal("expected 127.0.0.1 in list")
	}
	if _, ok := list.Lookup(net.ParseIP("10.0.0.1")); !ok {
		t.Fatal("expected 10.0.0.1 in list")
	}
}

func TestLookupBadIP(t *testing.T) {
	list := New([]Range{{
		First: net.ParseIP("127.0.0.1").To4(),
		Last:  net.ParseIP("127.0.0.1").To4(),
	}})
	if _, ok := list.Lookup(nil); ok {
		t.Fatal("nil IP should not match")
	}
	if _, ok := list.Lookup(net.ParseIP("127.0.0.1")); !ok {
		t.Fatal("expected match")
	}
	if _, ok := list.Lookup(net.ParseIP("8.8.8.8")); ok {
		t.Fatal("unexpected match")
	}
}

func TestWhitelistBlacklistInteraction(t *testing.T) {
	white, _ := scanBuf([]byte("10.0.0.0/8\n"))
	black, _ := scanBuf([]byte("10.0.0.5\n"))

	ip := net.ParseIP("10.0.0.5")
	minifyIP(&ip)
	if _, ok := white.Lookup(ip); !ok {
		t.Fatal("10.0.0.5 should be in whitelist")
	}
	if _, ok := black.Lookup(ip); !ok {
		t.Fatal("10.0.0.5 should be in blacklist")
	}

	ip2 := net.ParseIP("10.0.0.1")
	minifyIP(&ip2)
	if _, ok := white.Lookup(ip2); !ok {
		t.Fatal("10.0.0.1 should be in whitelist")
	}
	if _, ok := black.Lookup(ip2); ok {
		t.Fatal("10.0.0.1 should not be in blacklist")
	}
}

func TestClientIPIPv6(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "[2001:db8::1]:12345"

	ip := clientIP(c)
	if ip == nil || ip.String() != "2001:db8::1" {
		t.Fatalf("clientIP = %v, want 2001:db8::1", ip)
	}

	c.Request.RemoteAddr = "[2001:db8::2]"
	ip = clientIP(c)
	if ip == nil || ip.String() != "2001:db8::2" {
		t.Fatalf("clientIP without port = %v, want 2001:db8::2", ip)
	}
}

func TestMiddlewareBanStatus(t *testing.T) {
	withTestWAFDB(t, func(dir string) {
		if err := settings.SetWAFConfig(settings.WAFConfig{Blacklist: []string{"127.0.0.1"}}); err != nil {
			t.Fatal(err)
		}
		Load()

		r := gin.New()
		r.Use(WAF())
		r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:9999"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status=%d want 403", w.Code)
		}
		if w.Body.String() != "Banned" {
			t.Fatalf("body=%q", w.Body.String())
		}
	})
}

func TestMiddlewareWhitelistIPv6(t *testing.T) {
	withTestWAFDB(t, func(dir string) {
		if err := settings.SetWAFConfig(settings.WAFConfig{Whitelist: []string{"2001:db8::1"}}); err != nil {
			t.Fatal(err)
		}
		Load()

		r := gin.New()
		r.Use(WAF())
		r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "[2001:db8::1]:9999"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("allowed status=%d body=%q", w.Code, w.Body.String())
		}

		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.RemoteAddr = "[2001:db8::2]:9999"
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusForbidden {
			t.Fatalf("denied status=%d", w2.Code)
		}
	})
}

func TestMiddlewareWhitelistDoesNotBypassDefaultReferer(t *testing.T) {
	withTestWAFDB(t, func(dir string) {
		if err := settings.SetWAFConfig(settings.WAFConfig{Whitelist: []string{"127.0.0.1"}}); err != nil {
			t.Fatal(err)
		}
		Load()

		r := gin.New()
		r.Use(WAF())
		r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:9999"
		req.Header.Set("Referer", "https://"+defaultBlockedReferers[0]+"/embed")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("whitelisted IP with default blocked referer status=%d want 403", w.Code)
		}

		reqOK := httptest.NewRequest(http.MethodGet, "/", nil)
		reqOK.RemoteAddr = "127.0.0.1:9999"
		wOK := httptest.NewRecorder()
		r.ServeHTTP(wOK, reqOK)
		if wOK.Code != 200 {
			t.Fatalf("whitelisted IP without referer status=%d want 200", wOK.Code)
		}
	})
}

func TestStoreUpdateHotReload(t *testing.T) {
	withTestWAFDB(t, func(dir string) {
		Load()
		snap, err := Update(ListsUpdate{
			Whitelist: "8.8.8.8\n",
			Blacklist: "",
			Referers:  "evil.example\n",
		})
		if err != nil {
			t.Fatal(err)
		}
		if !snap.IPEnabled {
			t.Fatal("expected IP enabled")
		}
		if _, ok := snap.WhiteIP.Lookup(net.ParseIP("8.8.8.8")); !ok {
			t.Fatal("whitelist missing 8.8.8.8")
		}
		found := false
		for _, h := range snap.Referers {
			if h == "evil.example" {
				found = true
			}
		}
		if !found {
			t.Fatalf("referers=%v missing evil.example", snap.Referers)
		}

		cfg, ok, err := settings.GetWAFConfig()
		if err != nil || !ok {
			t.Fatalf("GetWAFConfig() ok=%v err=%v", ok, err)
		}
		if !slices.Equal(cfg.Whitelist, []string{"8.8.8.8"}) ||
			!slices.Equal(cfg.Referers, []string{"evil.example"}) {
			t.Fatalf("settings.json config=%+v", cfg)
		}
	})
}

func TestStoreMalformedSettingsFailsSoft(t *testing.T) {
	withTestWAFDB(t, func(dir string) {
		if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"waf":`), 0o600); err != nil {
			t.Fatal(err)
		}
		Load()
		snap := GetSnapshot()
		if len(snap.Referers) != len(defaultBlockedReferers) {
			t.Fatalf("referers=%v want built-in defaults", snap.Referers)
		}
		if len(snap.Warnings) != 1 || snap.Warnings[0].List != "storage" || snap.Warnings[0].Code != "read_failed" {
			t.Fatalf("warnings=%+v", snap.Warnings)
		}
	})
}

func TestExistingMiddlewareObservesHotReload(t *testing.T) {
	withTestWAFDB(t, func(_ string) {
		Load()
		r := gin.New()
		r.Use(WAF())
		r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

		request := func() int {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "127.0.0.1:9999"
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			return w.Code
		}
		if got := request(); got != http.StatusOK {
			t.Fatalf("before update status=%d", got)
		}
		if _, err := Update(ListsUpdate{Blacklist: "127.0.0.1"}); err != nil {
			t.Fatal(err)
		}
		if got := request(); got != http.StatusForbidden {
			t.Fatalf("after update status=%d", got)
		}
	})
}

func TestMiddlewareBlacklistOverridesWhitelist(t *testing.T) {
	withTestWAFDB(t, func(_ string) {
		if _, err := Update(ListsUpdate{
			Whitelist: "127.0.0.1",
			Blacklist: "127.0.0.1",
		}); err != nil {
			t.Fatal(err)
		}
		r := gin.New()
		r.Use(WAF())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:9999"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status=%d want 403", w.Code)
		}
	})
}

func TestMiddlewareRejectsMalformedClientIPWhenACLEnabled(t *testing.T) {
	withTestWAFDB(t, func(_ string) {
		if _, err := Update(ListsUpdate{Whitelist: "127.0.0.1"}); err != nil {
			t.Fatal(err)
		}
		r := gin.New()
		r.Use(WAF())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "not-an-ip"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status=%d want 403", w.Code)
		}
	})
}

func TestConcurrentUpdatesNeverMixConfigurations(t *testing.T) {
	withTestWAFDB(t, func(_ string) {
		const updates = 20
		var wg sync.WaitGroup
		errs := make(chan error, updates)
		for i := 0; i < updates; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				token := fmt.Sprintf("%d", i)
				snap, err := Update(ListsUpdate{
					Whitelist: "10.0.0." + token,
					Blacklist: "192.0.2." + token,
					Referers:  "host" + token + ".example",
				})
				if err != nil {
					errs <- err
					return
				}
				if snap.WhitelistText != "10.0.0."+token ||
					snap.BlacklistText != "192.0.2."+token ||
					snap.ReferersText != "host"+token+".example" {
					errs <- fmt.Errorf("mixed snapshot for %s: %+v", token, snap)
				}
			}(i)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			t.Error(err)
		}
	})
}
