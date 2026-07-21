package torznab

import (
	"testing"
)

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
