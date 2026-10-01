package web

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"server/settings"
)

func testCert(t *testing.T) tls.Certificate {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
}

// startSplitServer serves "ok" over TLS and the https redirect over plain HTTP on one port.
func startSplitServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	oldPort := settings.SslPort
	settings.SslPort = port
	t.Cleanup(func() {
		shutdownServers()
		settings.SslPort = oldPort
	})

	tlsLn, plainLn := splitTLS(ln)
	cert := testCert(t)
	srv := newServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok "+r.Proto)
	}))
	srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	serve(func() error { return srv.ServeTLS(tlsLn, "", "") })
	redirect := newServer(httpsRedirectHandler())
	serve(func() error { return redirect.Serve(plainLn) })
	return port
}

func insecureClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			ForceAttemptHTTP2: true,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

func TestSplitServesTLS(t *testing.T) {
	port := startSplitServer(t)
	resp, err := insecureClient().Get("https://127.0.0.1:" + port + "/echo")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok HTTP/2.0" {
		t.Fatalf("body = %q, want HTTP/2 response", body)
	}
}

func TestSplitRedirectsPlainHTTP(t *testing.T) {
	port := startSplitServer(t)
	resp, err := insecureClient().Get("http://127.0.0.1:" + port + "/stream/My%20Movie.mkv?link=abc&play")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want 307", resp.StatusCode)
	}
	want := "https://127.0.0.1:" + port + "/stream/My%20Movie.mkv?link=abc&play"
	if got := resp.Header.Get("Location"); got != want {
		t.Fatalf("Location = %q, want %q", got, want)
	}
}

func TestSplitSilentClientDoesNotBlockOthers(t *testing.T) {
	port := startSplitServer(t)
	silent, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		t.Fatal(err)
	}
	defer silent.Close()

	resp, err := insecureClient().Get("https://127.0.0.1:" + port + "/")
	if err != nil {
		t.Fatalf("request blocked by silent client: %v", err)
	}
	resp.Body.Close()
}

func TestShutdownServersIsQuiet(t *testing.T) {
	port := startSplitServer(t)
	resp, err := insecureClient().Get("https://127.0.0.1:" + port + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	shutdownServers()

	select {
	case err := <-waitChan:
		t.Fatalf("shutdown reported error to Wait: %v", err)
	case <-time.After(200 * time.Millisecond):
	}
	if _, err := net.DialTimeout("tcp", "127.0.0.1:"+port, time.Second); err == nil {
		t.Fatal("port still accepting after shutdown")
	}
}

func TestBuildHTTPSRedirectTarget(t *testing.T) {
	old := settings.SslPort
	t.Cleanup(func() { settings.SslPort = old })

	tests := []struct {
		sslPort, host, target, want string
	}{
		{"8091", "192.168.1.2:8090", "/web/?a=1", "https://192.168.1.2:8091/web/?a=1"},
		{"8091", "example.local", "/", "https://example.local:8091/"},
		{"443", "example.com:80", "/x", "https://example.com/x"},
		{"8091", "[::1]:8090", "/", "https://[::1]:8091/"},
		{"8091", "[::1]", "/", "https://[::1]:8091/"},
		{"443", "[fe80::1]:80", "/", "https://[fe80::1]/"},
		{"8091", "h:8090", "/stream/a%20b%2Fc.mkv", "https://h:8091/stream/a%20b%2Fc.mkv"},
	}
	for _, tt := range tests {
		settings.SslPort = tt.sslPort
		r := httptest.NewRequest(http.MethodGet, "http://placeholder"+tt.target, nil)
		r.Host = tt.host
		if got := buildHTTPSRedirectTarget(r); got != tt.want {
			t.Errorf("host %q target %q: got %q, want %q", tt.host, tt.target, got, tt.want)
		}
	}
}

func TestForceHTTPSHandler(t *testing.T) {
	old := settings.SslPort
	settings.SslPort = "8091"
	t.Cleanup(func() { settings.SslPort = old })

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "served")
	})
	tests := []struct {
		httpMedia    bool
		remote, path string
		served       bool
	}{
		// --force-https alone: everything redirects, including loopback and LAN
		// (a local reverse proxy or Docker makes internet clients look like either)
		{false, "127.0.0.1:5000", "/play/abc/1", false},
		{false, "192.168.1.50:5000", "/stream/Movie.mkv", false},
		{false, "203.0.113.9:5000", "/", false},
		// --http-media: media paths are served to any client, the rest redirects
		{true, "203.0.113.9:5000", "/play/abc/1", true},
		{true, "192.168.1.50:5000", "/stream/Movie.mkv", true},
		{true, "192.168.1.50:5000", "/stream", true},
		{true, "192.168.1.50:5000", "/playlist", true},
		{true, "192.168.1.50:5000", "/playlistall/all.m3u", true},
		{true, "192.168.1.50:5000", "/", false},
		{true, "127.0.0.1:5000", "/settings", false},
		{true, "192.168.1.50:5000", "/streamx", false},
		{true, "192.168.1.50:5000", "/playlistfoo", false},
		// GStreamer HLS paths are media, its control endpoints are not
		{true, "192.168.1.50:5000", "/gst/abc/master.m3u8", true},
		{true, "192.168.1.50:5000", "/gst/abc/video.m3u8", true},
		{true, "192.168.1.50:5000", "/gst/abc/init.mp4", true},
		{true, "192.168.1.50:5000", "/gst/abc/seg/00001.m4s", true},
		{true, "192.168.1.50:5000", "/gst/abc/subs/0.m3u8", true},
		{true, "192.168.1.50:5000", "/gst/abc/heartbeat", true},
		{true, "192.168.1.50:5000", "/gst/settings", false},
		{true, "192.168.1.50:5000", "/gst/remove", false},
		{true, "192.168.1.50:5000", "/gst/echo", false},
		{true, "192.168.1.50:5000", "/gst//master.m3u8", false},
		{false, "192.168.1.50:5000", "/gst/abc/master.m3u8", false},
	}
	for _, tt := range tests {
		h := forceHTTPSHandler(next, tt.httpMedia)
		r := httptest.NewRequest(http.MethodGet, "http://host:8090"+tt.path, nil)
		r.RemoteAddr = tt.remote
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		served := w.Code == http.StatusOK && w.Body.String() == "served"
		if served != tt.served {
			t.Errorf("httpMedia=%v %s %s: served=%v (code %d), want %v", tt.httpMedia, tt.remote, tt.path, served, w.Code, tt.served)
		}
		if !tt.served && w.Code != http.StatusTemporaryRedirect {
			t.Errorf("httpMedia=%v %s %s: code %d, want 307", tt.httpMedia, tt.remote, tt.path, w.Code)
		}
	}
}

func TestServerErrorLogThrottlesTLSErrors(t *testing.T) {
	var got []string
	l := &serverErrorLog{out: func(v ...any) { got = append(got, fmt.Sprint(v...)) }}
	lines := []string{
		"http: TLS handshake error from 192.168.0.169:52203: remote error: tls: unknown certificate\n",
		"http: TLS handshake error from 192.168.0.169:52204: remote error: tls: unknown certificate\n",
		"http: TLS handshake error from [fe80::1]:40000: EOF\n",
		"http: TLS handshake error from 192.168.0.169:52205: remote error: tls: unknown certificate\n",
		"http: panic serving 1.2.3.4:5: boom\n",
		"http: panic serving 1.2.3.4:5: boom\n",
	}
	for _, line := range lines {
		l.Write([]byte(line))
	}
	if len(got) != 4 {
		t.Fatalf("got %d lines, want 4 (one per client + both non-TLS lines):\n%s", len(got), strings.Join(got, "\n"))
	}
	if !strings.Contains(got[0], "from 192.168.0.169: remote error: tls: unknown certificate") ||
		!strings.Contains(got[0], "--sslcert") {
		t.Errorf("first line missing client or hint: %q", got[0])
	}
	if !strings.Contains(got[1], "from fe80::1: EOF") || strings.Contains(got[1], "--sslcert") {
		t.Errorf("ipv6 line wrong: %q", got[1])
	}
	if got[2] != "http: panic serving 1.2.3.4:5: boom" || got[3] != got[2] {
		t.Errorf("non-TLS lines must pass through unchanged: %q", got[2:])
	}

	// after the quiet period the client is reported again
	l.seen["192.168.0.169"] = time.Now().Add(-tlsErrorQuietPeriod - time.Second)
	l.Write([]byte(lines[0]))
	if len(got) != 5 {
		t.Fatal("client not reported again after the quiet period")
	}
}

func TestInternalServerBypassesForceHTTPS(t *testing.T) {
	oldInternal := settings.InternalPort
	t.Cleanup(func() { settings.InternalPort = oldInternal })

	startInternalServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "media")
	}))
	if settings.InternalPort == "" {
		t.Fatal("internal port not set")
	}
	base := settings.LoopbackBaseURL()
	if base != "http://127.0.0.1:"+settings.InternalPort {
		t.Fatalf("LoopbackBaseURL = %q, want the internal listener", base)
	}
	resp, err := insecureClient().Get(base + "/play/abc/1")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(body) != "media" {
		t.Fatalf("internal request: %d %q, want 200 media", resp.StatusCode, body)
	}

	port := settings.InternalPort
	shutdownServers()
	if _, err := net.DialTimeout("tcp", "127.0.0.1:"+port, time.Second); err == nil {
		t.Fatal("internal listener still accepting after shutdown")
	}
}

func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	return port
}

func TestStartServersHTTPSOnly(t *testing.T) {
	cert := testCert(t)
	keyDER, err := x509.MarshalECPrivateKey(cert.PrivateKey.(*ecdsa.PrivateKey))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath, keyPath := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}), 0o600)
	os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600)

	oldSsl, oldArgs, oldIPs, oldPort, oldSslPort, oldSets, oldInternal := settings.Ssl, settings.Args,
		settings.IPs, settings.Port, settings.SslPort, settings.BTsets, settings.InternalPort
	t.Cleanup(func() {
		if stopRenew != nil {
			close(stopRenew)
			stopRenew = nil
		}
		shutdownServers()
		settings.Ssl, settings.Args, settings.IPs, settings.Port, settings.SslPort, settings.BTsets,
			settings.InternalPort = oldSsl, oldArgs, oldIPs, oldPort, oldSslPort, oldSets, oldInternal
	})
	settings.Ssl = true
	settings.Args = &settings.ExecArgs{Ssl: true, HTTPSOnly: true}
	settings.IPs = []string{"127.0.0.1"}
	settings.Port, settings.SslPort = freePort(t), freePort(t)
	settings.BTsets = &settings.BTSets{SslCert: certPath, SslKey: keyPath}

	if err := startServers(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	})); err != nil {
		t.Fatal(err)
	}

	resp, err := insecureClient().Get("https://127.0.0.1:" + settings.SslPort + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(body) != "ok" {
		t.Fatalf("https: %d %q, want 200 ok", resp.StatusCode, body)
	}
	if conn, err := net.DialTimeout("tcp", "127.0.0.1:"+settings.Port, time.Second); err == nil {
		conn.Close()
		t.Fatal("plain HTTP port is open with --https-only")
	}
	if settings.InternalPort == "" {
		t.Fatal("internal loopback listener not started")
	}
}
