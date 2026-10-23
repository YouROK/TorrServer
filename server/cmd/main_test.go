package main

import (
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"server/settings"

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
