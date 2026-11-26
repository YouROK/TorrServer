package dlna

import (
	"testing"

	"server/settings"
)

func TestGetLinkFollowsHTTPSMode(t *testing.T) {
	oldSsl, oldArgs, oldPort, oldSslPort := settings.Ssl, settings.Args, settings.Port, settings.SslPort
	t.Cleanup(func() {
		settings.Ssl, settings.Args, settings.Port, settings.SslPort = oldSsl, oldArgs, oldPort, oldSslPort
	})
	settings.Port, settings.SslPort = "8090", "8091"

	tests := []struct {
		args *settings.ExecArgs
		want string
	}{
		{&settings.ExecArgs{}, "http://192.168.1.10:8090/play"},
		{&settings.ExecArgs{Ssl: true}, "http://192.168.1.10:8090/play"},
		{&settings.ExecArgs{Ssl: true, ForceHTTPS: true, HTTPMedia: true}, "http://192.168.1.10:8090/play"},
		{&settings.ExecArgs{Ssl: true, ForceHTTPS: true}, "https://192.168.1.10:8091/play"},
		{&settings.ExecArgs{Ssl: true, HTTPSOnly: true}, "https://192.168.1.10:8091/play"},
	}
	for _, tt := range tests {
		settings.Ssl, settings.Args = tt.args.Ssl, tt.args
		for _, host := range []string{"192.168.1.10:1900", "http://192.168.1.10:1900"} {
			if got := getLink(host, "play"); got != tt.want {
				t.Errorf("%+v, host %q: getLink = %q, want %q", *tt.args, host, got, tt.want)
			}
		}
	}
}
