package version

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserAgentFormat(t *testing.T) {
	got := UserAgent()
	want := "TorrServer/" + Version
	if got != want {
		t.Fatalf("UserAgent() = %q, want %q", got, want)
	}
}

func TestDefaultTransportSendsTorrServerUA(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if got != UserAgent() {
		t.Fatalf("User-Agent = %q, want %q", got, UserAgent())
	}
}

func TestWithUserAgentKeepsExplicitUA(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "MSX/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if got != "MSX/1.0" {
		t.Fatalf("User-Agent = %q, want MSX/1.0", got)
	}
}
