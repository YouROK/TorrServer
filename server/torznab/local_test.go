package torznab

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchCapsAndSearchLocal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(sampleCapsXML))
	}))
	defer srv.Close()
	prev := httpClient
	httpClient = srv.Client()
	t.Cleanup(func() { httpClient = prev })

	caps, err := FetchCaps(context.Background(), srv.URL, "key")
	if err != nil || caps == nil || caps.Limits.Max != 100 {
		t.Fatalf("caps %v %v", caps, err)
	}
	if err := Test(unsetCtx(), srv.URL, "key"); err != nil {
		t.Fatal(err)
	}
	if _, err := FetchCaps(unsetCtx(), "://bad", "key"); err == nil {
		t.Fatal("bad host")
	}
}

func unsetCtx() context.Context { return nil }

func TestFetchCapsFailures(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"non-200": func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", http.StatusForbidden) },
		"error":   func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`<error code="100"/>`)) },
		"empty":   func(w http.ResponseWriter, r *http.Request) {},
		"oversized": func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(bytes.Repeat([]byte(" "), 2<<20))
			_, _ = w.Write([]byte(sampleCapsXML))
		},
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			url := useIndexer(t, h)
			if caps, err := FetchCaps(context.Background(), url, "k"); err == nil {
				t.Fatalf("caps = %v, want error", caps)
			}
		})
	}
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	if _, err := FetchCaps(context.Background(), url, "k"); err == nil {
		t.Fatal("closed server")
	}
}
