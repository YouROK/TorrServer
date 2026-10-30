package web

import (
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"server/settings"
	"server/version"
)

// offlineWebSettings points the shared settings DB and the package BTS at a
// loopback-only, network-free configuration and restores everything after.
func offlineWebSettings(t *testing.T) {
	t.Helper()
	sslTestRouter(t) // opens the package's one settings DB
	settings.Ssl = false

	oldSets, oldArgs, oldIPs, oldPort, oldInternal := settings.BTsets, settings.Args, settings.IPs, settings.Port, settings.InternalPort
	oldPub4, oldPub6, oldTor := settings.PubIPv4, settings.PubIPv6, settings.TorAddr
	t.Cleanup(func() {
		settings.BTsets, settings.Args, settings.IPs, settings.Port, settings.InternalPort = oldSets, oldArgs, oldIPs, oldPort, oldInternal
		settings.PubIPv4, settings.PubIPv6, settings.TorAddr = oldPub4, oldPub6, oldTor
	})

	settings.BTsets = &settings.BTSets{
		CacheSize: 32 << 20, PreloadCache: 100, ReaderReadAHead: 50, ConnectionsLimit: 25,
		TorrentDisconnectTimeout: 120, ResponsiveMode: true,
		DisableDHT: true, DisablePEX: true, DisableUPNP: true, DisableUTP: true,
	}
	settings.PubIPv4, settings.PubIPv6, settings.TorAddr = "8.8.8.8", "", "127.0.0.1:0"
	settings.Args = &settings.ExecArgs{}
	settings.IPs = []string{"127.0.0.1"}
	settings.Port = freePort(t)
}

func getBody(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: %d %s", url, resp.StatusCode, body)
	}
	return string(body)
}

func TestStartServesEchoAndStopReleasesPorts(t *testing.T) {
	offlineWebSettings(t)

	if err := Start(); err != nil {
		t.Fatal(err)
	}
	stopped := false
	t.Cleanup(func() {
		if !stopped {
			go func() { _ = Wait() }()
			Stop()
		}
	})

	port := settings.Port
	if got := getBody(t, "http://127.0.0.1:"+port+"/echo"); got != version.Version {
		t.Fatalf("/echo on the public port = %q, want %q", got, version.Version)
	}
	if settings.InternalPort == "" {
		t.Fatal("internal loopback port not set")
	}
	internal := settings.InternalPort
	if got := getBody(t, "http://127.0.0.1:"+internal+"/echo"); got != version.Version {
		t.Fatalf("/echo on the internal port = %q", got)
	}
	if body := getBody(t, "http://127.0.0.1:"+port+"/swagger/index.html"); body == "" {
		t.Fatal("swagger index is empty")
	}
	resp, err := (&http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}).
		Get("http://127.0.0.1:" + port + "/swagger")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != "/swagger/index.html" {
		t.Fatalf("/swagger: %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}

	// Stop blocks on the unbuffered waitChan until Wait receives.
	waited := make(chan error, 1)
	go func() { waited <- Wait() }()
	stopDone := make(chan struct{})
	go func() { Stop(); close(stopDone) }()
	select {
	case err := <-waited:
		if err != nil {
			t.Fatalf("Wait = %v, want nil", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("Wait did not return after Stop")
	}
	select {
	case <-stopDone:
	case <-time.After(15 * time.Second):
		t.Fatal("Stop did not return")
	}
	stopped = true

	if settings.InternalPort != "" {
		t.Fatalf("InternalPort = %q after Stop", settings.InternalPort)
	}
	for _, p := range []string{port, internal} {
		if conn, err := net.DialTimeout("tcp", "127.0.0.1:"+p, time.Second); err == nil {
			_ = conn.Close()
			t.Fatalf("port %s still accepting after Stop", p)
		}
	}
}

func TestStartFailsWhenHTTPPortIsTaken(t *testing.T) {
	offlineWebSettings(t)
	busy, err := net.Listen("tcp", "127.0.0.1:"+settings.Port)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = busy.Close() }()
	t.Cleanup(BTS.Disconnect)

	if err := Start(); err == nil {
		t.Fatal("Start succeeded with the HTTP port taken")
	}
}

func TestStartFailsOnUnusableCertificate(t *testing.T) {
	offlineWebSettings(t)
	settings.Ssl = true
	settings.BTsets.SslCert = t.TempDir() + "/missing-cert.pem"
	settings.BTsets.SslKey = t.TempDir() + "/missing-key.pem"
	t.Cleanup(func() { settings.Ssl = false })

	if err := Start(); err == nil {
		t.Fatal("Start succeeded with a missing certificate")
	}
}
