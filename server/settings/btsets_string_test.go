package settings

import (
	"strings"
	"testing"
)

func TestBTSetsStringMasksKeys(t *testing.T) {
	s := &BTSets{
		TMDBSettings: TMDBConfig{APIKey: "tmdb-secret", APIURL: "https://api.themoviedb.org"},
		TorznabUrls:  []TorznabConfig{{Host: "https://jacred.example", Key: "torznab-secret"}},
	}
	out := s.String()
	for _, secret := range []string{"tmdb-secret", "torznab-secret"} {
		if strings.Contains(out, secret) {
			t.Fatalf("log form leaks %q: %s", secret, out)
		}
	}
	if !strings.Contains(out, "https://jacred.example") {
		t.Fatalf("log form lost non-secret fields: %s", out)
	}
	if s.TMDBSettings.APIKey != "tmdb-secret" || s.TorznabUrls[0].Key != "torznab-secret" {
		t.Fatal("String mutated the settings it formats")
	}
}
