package utils

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"server/settings"
)

func setupTrackersTest(t *testing.T, url, defaults string) {
	t.Helper()
	InvalidateTrackersCache()
	settings.BTsets = &settings.BTSets{
		TrackersListURL: url,
		DefaultTrackers: defaults,
	}
	t.Cleanup(func() {
		InvalidateTrackersCache()
		settings.BTsets = nil
	})
}

func TestGetDefTrackersSuccess(t *testing.T) {
	remote := "udp://198.51.100.1:1337/announce\nudp://198.51.100.2:6969/announce\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(remote))
	}))
	defer srv.Close()

	local := "udp://local.example:80/announce\nwss://tracker.example/announce"
	setupTrackersTest(t, srv.URL, local)

	got := GetDefTrackers()
	want := []string{
		"udp://198.51.100.1:1337/announce",
		"udp://198.51.100.2:6969/announce",
		"udp://local.example:80/announce",
		"wss://tracker.example/announce",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGetDefTrackersEmptyURLSkipsFetch(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://should-not-appear:1/announce\n"))
	}))
	defer srv.Close()

	local := "udp://only-local:1337/announce"
	setupTrackersTest(t, "", local)
	// Ensure empty URL is used even if a server exists.
	_ = srv

	got := GetDefTrackers()
	want := []string{"udp://only-local:1337/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	if hits.Load() != 0 {
		t.Fatalf("remote hits = %d, want 0 when URL is empty", hits.Load())
	}
}

func TestGetDefTrackersFallbackOnTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(trackersFetchTimeout + 2*time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://should-not-appear:1/announce\n"))
	}))
	defer srv.Close()

	local := "udp://fallback.example:80/announce"
	setupTrackersTest(t, srv.URL, local)

	start := time.Now()
	got := GetDefTrackers()
	elapsed := time.Since(start)

	if elapsed > trackersFetchTimeout+2*time.Second {
		t.Fatalf("GetDefTrackers took %v, want around %v timeout", elapsed, trackersFetchTimeout)
	}
	want := []string{"udp://fallback.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fallback = %#v, want %#v", got, want)
	}
}

func TestGetDefTrackersFallbackOnNonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer srv.Close()

	local := "udp://fallback.example:80/announce"
	setupTrackersTest(t, srv.URL, local)

	got := GetDefTrackers()
	want := []string{"udp://fallback.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fallback = %#v, want %#v", got, want)
	}
}

func TestGetDefTrackersNoRefetchAfterFailure(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer srv.Close()

	setupTrackersTest(t, srv.URL, "udp://fallback.example:80/announce")
	_ = GetDefTrackers()
	_ = GetDefTrackers()
	_ = GetDefTrackers()

	if hits.Load() != 1 {
		t.Fatalf("remote hits = %d, want 1 (sticky one-shot)", hits.Load())
	}
}

func TestInvalidateTrackersCacheAllowsRefetch(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote.example:1/announce\n"))
	}))
	defer srv.Close()

	setupTrackersTest(t, srv.URL, "udp://local.example:80/announce")
	_ = GetDefTrackers()
	if hits.Load() != 1 {
		t.Fatalf("hits after first load = %d, want 1", hits.Load())
	}

	InvalidateTrackersCache()
	got := GetDefTrackers()
	if hits.Load() != 2 {
		t.Fatalf("hits after invalidate = %d, want 2", hits.Load())
	}
	want := []string{"udp://remote.example:1/announce", "udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
