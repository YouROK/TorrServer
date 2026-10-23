package server

import (
	"strings"
	"testing"

	"server/settings"
)

func TestStartRejectsUnbindablePorts(t *testing.T) {
	dir := t.TempDir()
	oldArgs, oldPath, oldSsl, oldSets, oldRO, oldWA := settings.Args, settings.Path, settings.Ssl, settings.BTsets, settings.ReadOnly, settings.SearchWA
	t.Cleanup(func() {
		settings.CloseDB()
		settings.Args, settings.Path, settings.Ssl, settings.BTsets = oldArgs, oldPath, oldSsl, oldSets
		settings.ReadOnly, settings.SearchWA = oldRO, oldWA
	})
	settings.Path = dir
	cert := dir + "/cert.pem"
	key := dir + "/key.pem"
	settings.Args = &settings.ExecArgs{
		Path:    dir,
		Port:    "8090",
		IPs:     []string{"127.0.0.1"},
		Ssl:     true,
		SslPort: "1",
		SslCert: cert,
		SslKey:  key,
	}
	err := Start()
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("ssl: %v", err)
	}

	settings.CloseDB()
	settings.Args.Ssl = false
	settings.Args.Port = "1"
	err = Start()
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("http: %v", err)
	}

	settings.CloseDB()
	settings.Args.Ssl = true
	settings.Args.SslPort = ""
	settings.Args.SslCert = ""
	settings.Args.SslKey = ""
	settings.Args.Port = "1"
	err = Start()
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("default ssl port: %v", err)
	}

	if settings.BTsets != nil {
		settings.BTsets.SslPort = 8092
		settings.SetBTSets(settings.BTsets)
	}
	settings.CloseDB()
	settings.Args.SslPort = ""
	settings.Args.Port = "1"
	err = Start()
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("settings ssl port: %v", err)
	}
}
