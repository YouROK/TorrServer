package mcp

import (
	"context"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/log"
	"server/version"
)

const instructions = `TorrServer streams torrents over HTTP without requiring a full download.

Library categories: movie, tv, music, other (empty means uncategorized).
Viewed status is per torrent file (1-based file_index), not per show. Use mark_viewed after the user watches something, and get_next_unwatched to pick the next TV episode from filenames (SxxEyy / 1x05 / Season N).
Play URLs are ordinary HTTP links for VLC, mpv, or a browser. Prefer add_torrent with save=true so the torrent stays in the library.
Do not wipe the library or shut down the server; those operations are not exposed here.
Search (RuTor/Torznab) is optional and may be disabled in settings.`

type httpReqKey struct{}

func newServer() *mcpsdk.Server {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "torrserver",
		Title:   "TorrServer",
		Version: version.Version,
	}, &mcpsdk.ServerOptions{
		Instructions: instructions,
	})
	registerTools(server)
	return server
}

func newHTTPHandler(server *mcpsdk.Server) http.Handler {
	inner := mcpsdk.NewStreamableHTTPHandler(func(_ *http.Request) *mcpsdk.Server {
		return server
	}, &mcpsdk.StreamableHTTPOptions{
		Stateless: true,
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Host") == "" && r.Host != "" {
			r.Header.Set("Host", r.Host)
		}
		ctx := context.WithValue(r.Context(), httpReqKey{}, r)
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Handler() http.Handler {
	return newHTTPHandler(newServer())
}

func logMCPReady() {
	log.TLogln("MCP Streamable HTTP endpoint /mcp")
}
