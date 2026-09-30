package settings

import "testing"

func TestPlainHTTPServesMediaAndLoopbackURL(t *testing.T) {
	oldSsl, oldArgs, oldPort, oldSslPort, oldInternal := Ssl, Args, Port, SslPort, InternalPort
	t.Cleanup(func() { Ssl, Args, Port, SslPort, InternalPort = oldSsl, oldArgs, oldPort, oldSslPort, oldInternal })
	Port, SslPort, InternalPort = "8090", "8091", ""

	tests := []struct {
		ssl, force, media bool
		served            bool
		loopback          string
	}{
		{false, false, false, true, "http://127.0.0.1:8090"},
		{true, false, false, true, "http://127.0.0.1:8090"},
		{true, true, false, false, "https://127.0.0.1:8091"},
		{true, true, true, true, "http://127.0.0.1:8090"},
	}
	for _, tt := range tests {
		Ssl = tt.ssl
		Args = &ExecArgs{Ssl: tt.ssl, ForceHTTPS: tt.force, HTTPMedia: tt.media}
		if got := PlainHTTPServesMedia(); got != tt.served {
			t.Errorf("ssl=%v force=%v media=%v: PlainHTTPServesMedia=%v, want %v", tt.ssl, tt.force, tt.media, got, tt.served)
		}
		if got := LoopbackBaseURL(); got != tt.loopback {
			t.Errorf("ssl=%v force=%v media=%v: LoopbackBaseURL=%q, want %q", tt.ssl, tt.force, tt.media, got, tt.loopback)
		}
	}
}

func TestLoopbackBaseURLPrefersInternalListener(t *testing.T) {
	oldSsl, oldArgs, oldInternal := Ssl, Args, InternalPort
	t.Cleanup(func() { Ssl, Args, InternalPort = oldSsl, oldArgs, oldInternal })

	// strict --force-https: the public HTTP port redirects, the internal one doesn't
	Ssl = true
	Args = &ExecArgs{Ssl: true, ForceHTTPS: true}
	InternalPort = "54321"
	if got := LoopbackBaseURL(); got != "http://127.0.0.1:54321" {
		t.Fatalf("LoopbackBaseURL = %q, want the internal listener", got)
	}
}
