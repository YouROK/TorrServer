package torznab

import (
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
