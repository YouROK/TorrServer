package web

import (
	"net"
	"net/http"
	"net/url"

	"server/log"
	"server/settings"
)

func runHTTPRedirectToHTTPS(addr string) error {
	h := func(w http.ResponseWriter, r *http.Request) {
		target := buildHTTPSRedirectTarget(r)
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
	}
	log.TLogln("Start http server (redirect to https) at", addr)
	return http.ListenAndServe(addr, http.HandlerFunc(h))
}

func buildHTTPSRedirectTarget(r *http.Request) string {
	host := r.Host
	hostName, _, err := net.SplitHostPort(host)
	if err != nil {
		hostName = host
	}
	sslPort := settings.SslPort
	if sslPort == "" {
		sslPort = "8091"
	}
	var httpsHost string
	if sslPort == "443" {
		httpsHost = hostName
	} else {
		httpsHost = net.JoinHostPort(hostName, sslPort)
	}
	path := r.URL.EscapedPath()
	if path == "" {
		path = "/"
	}
	u := &url.URL{
		Scheme:   "https",
		Host:     httpsHost,
		Path:     path,
		RawQuery: r.URL.RawQuery,
	}
	return u.String()
}
