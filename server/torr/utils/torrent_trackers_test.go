package utils

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetDefTrackersSuccess(t *testing.T) {
	resetLoadedTrackersForTest()
	t.Cleanup(resetLoadedTrackersForTest)

	remote := "udp://198.51.100.1:1337/announce\nudp://198.51.100.2:6969/announce\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(remote))
	}))
	defer srv.Close()

	trackersListURL = srv.URL
	got := GetDefTrackers()

	wantPrefix := []string{
		"udp://198.51.100.1:1337/announce",
		"udp://198.51.100.2:6969/announce",
	}
	if len(got) < len(wantPrefix)+len(defTrackers) {
		t.Fatalf("got %d trackers, want at least %d", len(got), len(wantPrefix)+len(defTrackers))
	}
	if !reflect.DeepEqual(got[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("remote prefix = %#v, want %#v", got[:len(wantPrefix)], wantPrefix)
	}
	if !reflect.DeepEqual(got[len(wantPrefix):], copyDefTrackers()) {
		t.Fatalf("defTrackers suffix mismatch")
	}
}

func TestGetDefTrackersFallbackOnTimeout(t *testing.T) {
	resetLoadedTrackersForTest()
	t.Cleanup(resetLoadedTrackersForTest)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(trackersFetchTimeout + 2*time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://should-not-appear:1/announce\n"))
	}))
	defer srv.Close()

	trackersListURL = srv.URL
	start := time.Now()
	got := GetDefTrackers()
	elapsed := time.Since(start)

	if elapsed > trackersFetchTimeout+2*time.Second {
		t.Fatalf("GetDefTrackers took %v, want around %v timeout", elapsed, trackersFetchTimeout)
	}
	if !reflect.DeepEqual(got, copyDefTrackers()) {
		t.Fatalf("fallback = %#v, want built-in defTrackers", got)
	}
}

func TestGetDefTrackersFallbackOnNonOK(t *testing.T) {
	resetLoadedTrackersForTest()
	t.Cleanup(resetLoadedTrackersForTest)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer srv.Close()

	trackersListURL = srv.URL
	got := GetDefTrackers()
	if !reflect.DeepEqual(got, copyDefTrackers()) {
		t.Fatalf("fallback = %#v, want built-in defTrackers", got)
	}
}

func TestGetDefTrackersNoRefetchAfterFailure(t *testing.T) {
	resetLoadedTrackersForTest()
	t.Cleanup(resetLoadedTrackersForTest)

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer srv.Close()

	trackersListURL = srv.URL
	_ = GetDefTrackers()
	_ = GetDefTrackers()
	_ = GetDefTrackers()

	if hits.Load() != 1 {
		t.Fatalf("remote hits = %d, want 1 (sticky one-shot)", hits.Load())
	}
}
