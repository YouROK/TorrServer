package version

import (
	"net/http"
	"strings"
)

// Product is the RFC 9110 product token used for outbound HTTP.
const Product = "TorrServer"

func init() {
	http.DefaultTransport = WithUserAgent(http.DefaultTransport)
}

// UserAgent is the canonical app-level User-Agent, e.g. "TorrServer/MatriX.144".
// BitTorrent tracker/peer HTTP keeps qBittorrent in torr.BTServer — that is
// separate from these application sub-requests (Torznab, .torrent fetch, TMDB, …).
func UserAgent() string {
	return Product + "/" + Version
}

// SetRequest sets the canonical User-Agent on an outbound request.
func SetRequest(req *http.Request) {
	if req == nil {
		return
	}
	req.Header.Set("User-Agent", UserAgent())
}

type userAgentTransport struct {
	base http.RoundTripper
}

// WithUserAgent wraps rt so Go's default "Go-http-client/…" User-Agent is
// replaced with UserAgent(). An already-set product UA (browser, proxy, tests)
// is left unchanged.
func WithUserAgent(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	if wrapped, ok := rt.(*userAgentTransport); ok {
		return wrapped
	}
	return &userAgentTransport{base: rt}
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if replaceGoUserAgent(req.Header.Get("User-Agent")) {
		req = req.Clone(req.Context())
		req.Header.Set("User-Agent", UserAgent())
	}
	return t.base.RoundTrip(req)
}

func replaceGoUserAgent(ua string) bool {
	return ua == "" || strings.HasPrefix(ua, "Go-http-client/")
}
