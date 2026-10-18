package mcp

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPListTools(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Mount(r)

	ts := httptest.NewServer(r)
	defer ts.Close()

	ctx := context.Background()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "torrserver-test", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, &mcpsdk.StreamableClientTransport{
		Endpoint:             ts.URL + "/mcp",
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Errorf("close session: %v", err)
		}
	})

	res, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	want := map[string]bool{
		"get_server_info":    false,
		"list_torrents":      false,
		"get_torrent":        false,
		"add_torrent":        false,
		"update_torrent":     false,
		"remove_torrent":     false,
		"drop_torrent":       false,
		"get_play_url":       false,
		"get_playlist_url":   false,
		"list_viewed":        false,
		"mark_viewed":        false,
		"unmark_viewed":      false,
		"get_next_unwatched": false,
		"search_torrents":    false,
	}
	for _, tool := range res.Tools {
		if _, ok := want[tool.Name]; ok {
			want[tool.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("missing tool %q", name)
		}
	}

	info, err := session.CallTool(ctx, &mcpsdk.CallToolParams{Name: "get_server_info", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("get_server_info: %v", err)
	}
	if info.IsError {
		t.Fatalf("get_server_info is error: %+v", info)
	}
}
