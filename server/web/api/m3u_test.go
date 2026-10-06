package api

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-contrib/location/v2"
	"github.com/gin-gonic/gin"

	sets "server/settings"
	"server/web/sslcerts"
)

func TestMediaBaseURL(t *testing.T) {
	oldPath, oldPort, oldSets := sets.Path, sets.Port, sets.BTsets
	t.Cleanup(func() { sets.Path, sets.Port, sets.BTsets = oldPath, oldPort, oldSets })
	sets.Path = t.TempDir()
	sets.Port = "8090"

	genCert, genKey, err := sslcerts.MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	userDir := t.TempDir()
	userCert, userKey := filepath.Join(userDir, "c.pem"), filepath.Join(userDir, "k.pem")
	for src, dst := range map[string]string{genCert: userCert, genKey: userKey} {
		b, _ := os.ReadFile(src)
		os.WriteFile(dst, b, 0o600)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(location.Default())
	r.GET("/playlist", func(c *gin.Context) { c.String(http.StatusOK, mediaBaseURL(c)) })

	oldSsl, oldArgs := sets.Ssl, sets.Args
	t.Cleanup(func() { sets.Ssl, sets.Args = oldSsl, oldArgs })
	sets.Ssl = true

	tests := []struct {
		name       string
		cert, key  string
		forceHTTPS bool
		httpMedia  bool
		host       string
		tls        bool
		want       string
	}{
		{"self-signed, http serves media -> http port", genCert, genKey, false, false, "192.168.1.2:8091", true, "http://192.168.1.2:8090"},
		{"ipv6 host", genCert, genKey, false, false, "[fe80::1]:8091", true, "http://[fe80::1]:8090"},
		{"self-signed with --force-https --http-media -> http port", genCert, genKey, true, true, "192.168.1.2:8091", true, "http://192.168.1.2:8090"},
		{"self-signed with --force-https only stays https", genCert, genKey, true, false, "192.168.1.2:8091", true, "https://192.168.1.2:8091"},
		{"user cert stays https", userCert, userKey, false, false, "tv.example.com:8091", true, "https://tv.example.com:8091"},
		{"plain http unchanged", genCert, genKey, false, false, "192.168.1.2:8090", false, "http://192.168.1.2:8090"},
	}
	for _, tt := range tests {
		sets.BTsets = &sets.BTSets{SslCert: tt.cert, SslKey: tt.key}
		sets.Args = &sets.ExecArgs{ForceHTTPS: tt.forceHTTPS, HTTPMedia: tt.httpMedia}
		req := httptest.NewRequest(http.MethodGet, "/playlist", nil)
		req.Host = tt.host
		if tt.tls {
			req.TLS = &tls.ConnectionState{}
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if got := w.Body.String(); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestMediaBaseEndpoint(t *testing.T) {
	oldPath, oldPort, oldSets, oldSsl, oldArgs, oldAuth := sets.Path, sets.Port, sets.BTsets, sets.Ssl, sets.Args, sets.HttpAuth
	t.Cleanup(func() {
		sets.Path, sets.Port, sets.BTsets, sets.Ssl, sets.Args, sets.HttpAuth = oldPath, oldPort, oldSets, oldSsl, oldArgs, oldAuth
	})
	sets.Path = t.TempDir()
	sets.Port = "8090"
	sets.Ssl = true
	cert, key, err := sslcerts.MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	sets.BTsets = &sets.BTSets{SslCert: cert, SslKey: key}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(location.Default())
	SetupRoute(r)

	get := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/mediabase", nil)
		req.Host = "192.168.1.2:8091"
		req.TLS = &tls.ConnectionState{}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	tests := []struct {
		name              string
		forceHTTPS, media bool
		want              string
	}{
		{"self-signed, http serves media", false, false, `{"base":"http://192.168.1.2:8090"}`},
		{"strict --force-https keeps https", true, false, `{"base":"https://192.168.1.2:8091"}`},
		{"--force-https --http-media", true, true, `{"base":"http://192.168.1.2:8090"}`},
	}
	for _, tt := range tests {
		sets.Args = &sets.ExecArgs{ForceHTTPS: tt.forceHTTPS, HTTPMedia: tt.media}
		w := get()
		if w.Code != http.StatusOK || w.Body.String() != tt.want {
			t.Errorf("%s: %d %s, want %s", tt.name, w.Code, w.Body.String(), tt.want)
		}
		if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
			t.Errorf("%s: Cache-Control = %q, want no-store", tt.name, cc)
		}
	}

	// the settings change at runtime: a trusted cert takes effect without restart
	userDir := t.TempDir()
	userCert, userKey := filepath.Join(userDir, "c.pem"), filepath.Join(userDir, "k.pem")
	b, _ := os.ReadFile(cert)
	os.WriteFile(userCert, b, 0o644)
	b, _ = os.ReadFile(key)
	os.WriteFile(userKey, b, 0o600)
	sets.BTsets = &sets.BTSets{SslCert: userCert, SslKey: userKey}
	sets.Args = &sets.ExecArgs{}
	if w := get(); w.Body.String() != `{"base":"https://192.168.1.2:8091"}` {
		t.Errorf("trusted cert: %s, want https base", w.Body.String())
	}

	// requires authentication when --httpauth is on
	sets.HttpAuth = true
	if w := get(); w.Code != http.StatusUnauthorized {
		t.Errorf("without credentials: %d, want 401", w.Code)
	}
}
