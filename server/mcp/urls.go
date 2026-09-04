package mcp

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/settings"
	"server/torr/state"
)

func requestFromTool(ctxReq *mcpsdk.CallToolRequest) *http.Request {
	if ctxReq == nil {
		return nil
	}
	if ctxReq.Extra != nil {
		if host := strings.TrimSpace(ctxReq.Extra.Header.Get("Host")); host != "" {
			scheme := headerScheme(ctxReq.Extra.Header)
			return &http.Request{Host: host, Header: ctxReq.Extra.Header, URL: &url.URL{Scheme: scheme, Host: host}}
		}
		if xf := strings.TrimSpace(firstCSV(ctxReq.Extra.Header.Get("X-Forwarded-Host"))); xf != "" {
			scheme := headerScheme(ctxReq.Extra.Header)
			return &http.Request{Host: xf, Header: ctxReq.Extra.Header, URL: &url.URL{Scheme: scheme, Host: xf}}
		}
	}
	return nil
}

func headerScheme(h http.Header) string {
	if h == nil {
		return defaultScheme()
	}
	if proto := firstCSV(h.Get("X-Forwarded-Proto")); proto != "" {
		return strings.ToLower(proto)
	}
	return defaultScheme()
}

func defaultScheme() string {
	if settings.Ssl {
		return "https"
	}
	return "http"
}

func firstCSV(v string) string {
	if i := strings.IndexByte(v, ','); i >= 0 {
		return strings.TrimSpace(v[:i])
	}
	return strings.TrimSpace(v)
}

func baseURLFromRequest(r *http.Request) string {
	if r == nil {
		return fallbackBaseURL()
	}
	scheme := defaultScheme()
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := firstCSV(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		scheme = strings.ToLower(proto)
	} else if r.URL != nil && r.URL.Scheme != "" {
		scheme = r.URL.Scheme
	}
	host := r.Host
	if xf := firstCSV(r.Header.Get("X-Forwarded-Host")); xf != "" {
		host = xf
	}
	if host == "" {
		return fallbackBaseURL()
	}
	return scheme + "://" + host
}

func fallbackBaseURL() string {
	port := settings.Port
	if port == "" {
		port = "8090"
	}
	if settings.Ssl {
		if settings.SslPort != "" {
			port = settings.SslPort
		}
		return "https://127.0.0.1:" + port
	}
	return "http://127.0.0.1:" + port
}

func baseURL(req *mcpsdk.CallToolRequest) string {
	return baseURLFromRequest(requestFromTool(req))
}

func playURL(base, hash, path string, fileID int) string {
	name := filepath.Base(path)
	if name == "" || name == "." {
		name = "file"
	}
	return strings.TrimRight(base, "/") + "/stream/" + url.PathEscape(name) +
		"?link=" + url.QueryEscape(hash) + "&index=" + strconv.Itoa(fileID) + "&play"
}

func shortPlayURL(base, hash string, fileID int) string {
	return strings.TrimRight(base, "/") + "/play/" + url.PathEscape(hash) + "/" + strconv.Itoa(fileID)
}

func playlistURL(base, hash string) string {
	if hash == "" {
		return strings.TrimRight(base, "/") + "/playlistall/all.m3u"
	}
	return strings.TrimRight(base, "/") + "/playlist?hash=" + url.QueryEscape(hash)
}

func formatEpisodeCode(season, episode int) string {
	if season <= 0 && episode <= 0 {
		return ""
	}
	if season <= 0 {
		return fmt.Sprintf("E%02d", episode)
	}
	if episode <= 0 {
		return fmt.Sprintf("S%02d", season)
	}
	return fmt.Sprintf("S%02dE%02d", season, episode)
}

func filePlayURLs(base string, st *state.TorrentStatus, f *state.TorrentFileStat) (play, short string) {
	if st == nil || f == nil {
		return "", ""
	}
	return playURL(base, st.Hash, f.Path, f.Id), shortPlayURL(base, st.Hash, f.Id)
}
