package torznab

import (
	"context"
	"os"
	"testing"
	"time"

	"server/settings"
)

var indexerSearchQueries = []string{
	"матрица 1999",
	"matrix 1999",
}

func TestJacRedStreamMatrixSearches(t *testing.T) {
	if os.Getenv("LIVE_INDEXER") != "1" {
		t.Skip("set LIVE_INDEXER=1 to hit https://jacred.stream (key pp)")
	}

	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{
		EnableTorznabSearch: true,
		TorznabUrls: []settings.TorznabConfig{{
			Host:    JacRedHost,
			Key:     JacRedKey,
			Name:    "JacRed",
			CatType: settings.CategoryAll,
		}},
	}
	t.Cleanup(func() { settings.BTsets = prev })

	client := *httpClient
	client.Timeout = 30 * time.Second
	prevClient := httpClient
	httpClient = &client
	t.Cleanup(func() { httpClient = prevClient })

	for _, q := range indexerSearchQueries {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		list := Search(ctx, q, -1, "", 0, 50)
		cancel()
		if len(list) == 0 {
			t.Errorf("query %q: 0 results", q)
			continue
		}
		t.Logf("query %q: %d hits", q, len(list))
	}
}
