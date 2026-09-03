package torznab

import (
	"encoding/xml"
	"testing"

	"server/settings"
)

func TestEffectiveCat(t *testing.T) {
	tests := []struct {
		name       string
		requestCat string
		categories string
		catType    settings.CategoryType
		want       string
	}{
		{name: "ui cat wins", requestCat: "3000", categories: "2000", catType: settings.CategoryManual, want: "3000"},
		{name: "default movies tv", requestCat: "", categories: "", catType: settings.CategoryDefault, want: "5000,2000"},
		{name: "empty type is default", requestCat: "", categories: "", catType: "", want: "5000,2000"},
		{name: "manual categories", requestCat: "", categories: "1000,7000", catType: settings.CategoryManual, want: "1000,7000"},
		{name: "all omits cat", requestCat: "", categories: "1000", catType: settings.CategoryAll, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := effectiveCat(tt.requestCat, tt.categories, tt.catType)
			if got != tt.want {
				t.Fatalf("effectiveCat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProwlarrIndexerFallback(t *testing.T) {
	const body = `<?xml version="1.0"?><rss version="2.0"><channel>
<item>
  <title>Prowlarr Only</title>
  <prowlarrindexer>MyProwlarr</prowlarrindexer>
</item>
<item>
  <title>Jackett Wins</title>
  <jackettindexer>JackettName</jackettindexer>
  <prowlarrindexer>ShouldIgnore</prowlarrindexer>
</item>
</channel></rss>`

	var resp TorznabResponse
	if err := xml.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Channel.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(resp.Channel.Items))
	}

	trackerFor := func(item TorznabItem, label string) string {
		tracker := item.Indexer
		if tracker == "" {
			tracker = item.Prowlarr
		}
		if tracker == "" {
			tracker = label
		}
		return tracker
	}

	if got := trackerFor(resp.Channel.Items[0], "fallback.example"); got != "MyProwlarr" {
		t.Fatalf("prowlarr-only tracker = %q, want MyProwlarr", got)
	}
	if got := trackerFor(resp.Channel.Items[1], "fallback.example"); got != "JackettName" {
		t.Fatalf("jackett tracker = %q, want JackettName", got)
	}
}

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		want    string
		wantErr bool
	}{
		{
			name: "jackett torznab base",
			host: "http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/",
			want: "http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/api",
		},
		{
			name: "already ends with api",
			host: "http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/api",
			want: "http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/api",
		},
		{
			name: "already ends with api slash",
			host: "http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/api/",
			want: "http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/api",
		},
		{
			name: "prowlarr indexer id",
			host: "http://localhost:9696/1",
			want: "http://localhost:9696/1/api",
		},
		{
			name: "scheme default",
			host: "localhost:9117/api/v2.0/indexers/all/results/torznab/",
			want: "http://localhost:9117/api/v2.0/indexers/all/results/torznab/api",
		},
		{
			name:    "empty",
			host:    "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := normalizeHost(tt.host)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", u)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := u.String(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
