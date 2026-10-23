package main

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"server"
	"server/settings"
	"server/torr/offlinetest"

	"github.com/alexflint/go-arg"
)

func TestIPsSeparateFlag(t *testing.T) {
	var params args
	p, err := arg.NewParser(arg.Config{}, &params)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"--ip", "127.0.0.1", "--ip", "192.168.1.100"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"127.0.0.1", "192.168.1.100"}
	if len(params.IPs) != len(want) {
		t.Fatalf("IPs = %#v, want %#v", params.IPs, want)
	}
	for i := range want {
		if params.IPs[i] != want[i] {
			t.Fatalf("IPs = %#v, want %#v", params.IPs, want)
		}
	}
}

func TestArgsVersionAndFlags(t *testing.T) {
	var params args
	if !strings.Contains(params.Version(), "TorrServer") {
		t.Fatalf("Version() = %q", params.Version())
	}
	p, err := arg.NewParser(arg.Config{}, &params)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse([]string{
		"-p", "9000",
		"--ssl",
		"--sslport", "9443",
		"--https-only",
		"--force-https",
		"--http-media",
		"-m", "10",
		"--pprof",
		"--proxyurl", "socks5://127.0.0.1:1080",
		"--proxymode", "peers",
	})
	if err != nil {
		t.Fatal(err)
	}
	if params.Port != "9000" || params.SslPort != "9443" || !params.Ssl || !params.HTTPSOnly || !params.ForceHTTPS || !params.HTTPMedia {
		t.Fatalf("%+v", params)
	}
	if params.MaxSize != "10" || !params.Pprof || params.ProxyURL != "socks5://127.0.0.1:1080" || params.ProxyMode != "peers" {
		t.Fatalf("%+v", params)
	}
}

func TestArgsRejectsUnknownFlag(t *testing.T) {
	var params args
	p, err := arg.NewParser(arg.Config{}, &params)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"--not-a-flag"}); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestPreconfigDontKillIgnoresHUP(t *testing.T) {
	// The signal loop lives for the process, so BTsets must stay non-nil.
	if settings.BTsets == nil {
		settings.BTsets = &settings.BTSets{}
	}
	Preconfig(true)
	time.Sleep(20 * time.Millisecond)
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Signal(syscall.SIGHUP); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
}

func TestRunFlags(t *testing.T) {
	prev := net.DefaultResolver
	t.Cleanup(func() { net.DefaultResolver = prev })

	if code := run([]string{"--help"}); code != 0 {
		t.Fatal(code)
	}
	if code := run([]string{"--version"}); code != 0 {
		t.Fatal(code)
	}
	if code := run([]string{"--not-a-flag"}); code != 2 {
		t.Fatal(code)
	}

	dir := t.TempDir()
	base := []string{"-k", "-a", "-r", "-d", dir, "-p", "1", "-t", dir, "-m", "1", "-4", "8.8.8.8", "-6", "::1", "--torrentaddr", "127.0.0.1:0", "--proxyurl", "socks5://127.0.0.1:1", "--proxymode", "nope", "--pprof"}
	if code := run(append(append([]string{}, base...), "--force-https")); code != 1 {
		t.Fatal("force-https", code)
	}
	if code := run([]string{"-k", "-d", dir, "--https-only"}); code != 1 {
		t.Fatal("https-only", code)
	}
	if code := run([]string{"-k", "-d", dir, "--ssl", "--https-only", "--http-media"}); code != 1 {
		t.Fatal("combined", code)
	}
	if code := run([]string{"-k", "-d", dir, "--http-media"}); code != 1 {
		t.Fatal("http-media", code)
	}
}

func TestRunStartsAndStops(t *testing.T) {
	prev := net.DefaultResolver
	t.Cleanup(func() { net.DefaultResolver = prev })

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	dir := t.TempDir()
	done := make(chan int, 1)
	go func() {
		done <- run([]string{
			"-k", "-d", dir, "-p", strconv.Itoa(port),
			"-i", "127.0.0.1", "-4", "8.8.8.8",
			"-t", dir, "--torrentaddr", "127.0.0.1:0",
		})
	}()

	deadline := time.Now().Add(45 * time.Second)
	var up bool
	// Dial the IP directly. run() replaces net.DefaultResolver, and http.Get would race with that write.
	client := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			// A nil Resolver reads net.DefaultResolver, which run() replaces.
			return (&net.Dialer{Resolver: &net.Resolver{}}).DialContext(ctx, "tcp", "127.0.0.1:"+strconv.Itoa(port))
		},
	}}
	for time.Now().Before(deadline) {
		resp, err := client.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/echo")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				up = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !up {
		t.Fatal("server did not answer /echo")
	}
	show := offlinetest.WriteShow(t)
	var meta bytes.Buffer
	if err := show.Meta.Write(&meta); err != nil {
		t.Fatal(err)
	}
	torrentPath := filepath.Join(dir, "show.torrent")
	if err := os.WriteFile(torrentPath, meta.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(torrentPath); os.IsNotExist(err) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if _, err := os.Stat(torrentPath); err == nil {
		t.Fatal("watcher left the torrent file")
	}
	server.Stop()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatal(code)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("shutdown timed out")
	}
	if settings.BTsets == nil {
		settings.BTsets = &settings.BTSets{}
	}
}

func TestDNSCheckerFallback(t *testing.T) {
	prev := net.DefaultResolver
	t.Cleanup(func() { net.DefaultResolver = prev })
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(context.Context, string, string) (net.Conn, error) {
			return nil, context.DeadlineExceeded
		},
	}
	d := NewDNSChecker(DNSConfig{
		PrimaryServers:  []string{"127.0.0.1:1"},
		FallbackServers: []string{"127.0.0.1:1"},
		Timeout:         200 * time.Millisecond,
		CacheDuration:   time.Minute,
	})
	if d.testSystemDNS() {
		t.Fatal("system dns should fail")
	}
	d.initCustomResolver()
	if _, err := d.LookupHostWithFallback("example.invalid"); err == nil {
		t.Fatal("expected lookup error")
	}
	d.addToCache("example.invalid", []string{"1.2.3.4"})
	if addrs, ok := d.getFromCache("example.invalid"); !ok || len(addrs) != 1 {
		t.Fatal("cache")
	}
	if _, err := d.LookupHostWithFallback("example.invalid"); err != nil {
		t.Fatal(err)
	}
	go watchDir(filepath.Join(t.TempDir(), "missing"), nil)
}
