package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"server/settings"
)

func setupTrackersTest(t *testing.T, url, defaults string) {
	setupTrackersTestWithURLs(t, url, nil, defaults)
}

func setupTrackersTestWithURLs(t *testing.T, custom string, mirrors []string, defaults string) {
	t.Helper()
	defaultTrackersListURLs = append([]string(nil), mirrors...)
	settings.BTsets = &settings.BTSets{
		TrackersListURL: custom,
		DefaultTrackers: defaults,
	}
	InvalidateTrackersCache()
	t.Cleanup(resetTrackersTestState)
}

func resetTrackersTestState() {
	StopTrackers()
	settings.BTsets = nil
	defaultTrackersListURLs = append([]string(nil), settings.DefaultTrackersListURLs...)
}

func waitForTrackersFetch(t *testing.T, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		trackersMu.RLock()
		ready := loadedTrackers != nil
		trackersMu.RUnlock()
		if ready {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timeout waiting for trackers fetch")
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

	waitForTrackersFetch(t, 2*time.Second)

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

func TestConfiguredTrackersListURLsEmptyUsesDefaults(t *testing.T) {
	defaults := []string{"https://a.example/list.txt", "https://b.example/list.txt"}
	defaultTrackersListURLs = append([]string(nil), defaults...)
	t.Cleanup(resetTrackersTestState)

	settings.BTsets = &settings.BTSets{TrackersListURL: ""}
	got := configuredTrackersListURLs()
	if !reflect.DeepEqual(got, defaults) {
		t.Fatalf("got %#v, want %#v", got, defaults)
	}

	settings.BTsets = nil
	got = configuredTrackersListURLs()
	if !reflect.DeepEqual(got, defaults) {
		t.Fatalf("nil BTsets got %#v, want %#v", got, defaults)
	}
}

func TestConfiguredTrackersListURLsCustomPrepended(t *testing.T) {
	defaultTrackersListURLs = []string{"https://a.example/list.txt", "https://b.example/list.txt"}
	t.Cleanup(resetTrackersTestState)

	settings.BTsets = &settings.BTSets{TrackersListURL: " https://custom.example/list.txt "}
	got := configuredTrackersListURLs()
	want := []string{"https://custom.example/list.txt", "https://a.example/list.txt", "https://b.example/list.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestConfiguredTrackersListURLsDedupe(t *testing.T) {
	defaultTrackersListURLs = []string{"https://a.example/list.txt", "https://b.example/list.txt"}
	t.Cleanup(resetTrackersTestState)

	settings.BTsets = &settings.BTSets{TrackersListURL: "https://a.example/list.txt"}
	got := configuredTrackersListURLs()
	want := []string{"https://a.example/list.txt", "https://b.example/list.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGetDefTrackersEmptyURLUsesMirrors(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote.example:1/announce\n"))
	}))
	defer srv.Close()

	local := "udp://only-local:1337/announce"
	setupTrackersTestWithURLs(t, "", []string{srv.URL}, local)
	waitForTrackersFetch(t, 2*time.Second)

	got := GetDefTrackers()
	want := []string{"udp://remote.example:1/announce", "udp://only-local:1337/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	if hits.Load() != 1 {
		t.Fatalf("mirror hits = %d, want 1 when custom URL is empty", hits.Load())
	}
}

func TestGetDefTrackersFallsBackToNextURL(t *testing.T) {
	failSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer failSrv.Close()

	okSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://mirror.example:1/announce\n"))
	}))
	defer okSrv.Close()

	local := "udp://local.example:80/announce"
	setupTrackersTestWithURLs(t, "", []string{failSrv.URL, okSrv.URL}, local)
	waitForTrackersFetch(t, 2*time.Second)

	got := GetDefTrackers()
	want := []string{"udp://mirror.example:1/announce", "udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGetDefTrackersAllURLsFailUsesLocal(t *testing.T) {
	fail1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer fail1.Close()
	fail2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer fail2.Close()

	local := "udp://fallback.example:80/announce"
	setupTrackersTestWithURLs(t, "", []string{fail1.URL, fail2.URL}, local)
	waitForTrackersFetch(t, 2*time.Second)

	got := GetDefTrackers()
	want := []string{"udp://fallback.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGetDefTrackersCustomTriedBeforeMirrors(t *testing.T) {
	var mirrorHits atomic.Int32
	customSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://custom.example:1/announce\n"))
	}))
	defer customSrv.Close()
	mirrorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mirrorHits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://mirror.example:1/announce\n"))
	}))
	defer mirrorSrv.Close()

	local := "udp://local.example:80/announce"
	setupTrackersTestWithURLs(t, customSrv.URL, []string{mirrorSrv.URL}, local)
	waitForTrackersFetch(t, 2*time.Second)

	got := GetDefTrackers()
	want := []string{"udp://custom.example:1/announce", "udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	if mirrorHits.Load() != 0 {
		t.Fatalf("mirror hits = %d, want 0 when custom URL succeeds", mirrorHits.Load())
	}
}

func TestGetDefTrackersCustomFailsThenMirror(t *testing.T) {
	customSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer customSrv.Close()
	mirrorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://mirror.example:1/announce\n"))
	}))
	defer mirrorSrv.Close()

	local := "udp://local.example:80/announce"
	setupTrackersTestWithURLs(t, customSrv.URL, []string{mirrorSrv.URL}, local)
	waitForTrackersFetch(t, 2*time.Second)

	got := GetDefTrackers()
	want := []string{"udp://mirror.example:1/announce", "udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGetDefTrackersInstantBeforePrefetch(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote.example:1/announce\n"))
	}))
	defer srv.Close()
	defer close(release)

	local := "udp://local.example:80/announce"
	setupTrackersTest(t, srv.URL, local)

	start := time.Now()
	got := GetDefTrackers()
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Fatalf("GetDefTrackers took %v, want instant return before prefetch completes", elapsed)
	}
	want := []string{"udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestPrefetchCompletesInBackground(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote.example:1/announce\n"))
	}))
	defer srv.Close()

	local := "udp://local.example:80/announce"
	setupTrackersTest(t, srv.URL, local)

	got := GetDefTrackers()
	wantLocal := []string{"udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, wantLocal) {
		t.Fatalf("initial got %#v, want %#v", got, wantLocal)
	}

	close(release)
	waitForTrackersFetch(t, 2*time.Second)

	got = GetDefTrackers()
	wantMerged := []string{"udp://remote.example:1/announce", "udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, wantMerged) {
		t.Fatalf("after prefetch got %#v, want %#v", got, wantMerged)
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

	if elapsed > 200*time.Millisecond {
		t.Fatalf("GetDefTrackers took %v, want instant return with local fallback", elapsed)
	}
	want := []string{"udp://fallback.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("immediate fallback = %#v, want %#v", got, want)
	}

	waitForTrackersFetch(t, trackersFetchTimeout+2*time.Second)

	got = GetDefTrackers()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cached fallback = %#v, want %#v", got, want)
	}
}

func TestGetDefTrackersFallbackOnNonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer srv.Close()

	local := "udp://fallback.example:80/announce"
	setupTrackersTest(t, srv.URL, local)

	start := time.Now()
	got := GetDefTrackers()
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("GetDefTrackers blocked on remote fetch failure")
	}

	waitForTrackersFetch(t, 2*time.Second)

	want := []string{"udp://fallback.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("immediate fallback = %#v, want %#v", got, want)
	}
	got = GetDefTrackers()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cached fallback = %#v, want %#v", got, want)
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
	waitForTrackersFetch(t, 2*time.Second)
	_ = GetDefTrackers()
	_ = GetDefTrackers()

	if hits.Load() != 1 {
		t.Fatalf("remote hits = %d, want 1 (sticky one-shot)", hits.Load())
	}
}

func TestPrefetchDedupeSameGeneration(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote.example:1/announce\n"))
	}))
	defer srv.Close()

	setupTrackersTest(t, srv.URL, "udp://local.example:80/announce")
	PrefetchTrackers()
	PrefetchTrackers()
	PrefetchTrackers()
	waitForTrackersFetch(t, 2*time.Second)

	if hits.Load() != 1 {
		t.Fatalf("remote hits = %d, want 1 (deduped prefetch)", hits.Load())
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
	waitForTrackersFetch(t, 2*time.Second)
	if hits.Load() != 1 {
		t.Fatalf("hits after first load = %d, want 1", hits.Load())
	}

	InvalidateTrackersCache()
	waitForTrackersFetch(t, 2*time.Second)

	if hits.Load() != 2 {
		t.Fatalf("hits after invalidate = %d, want 2", hits.Load())
	}
	got := GetDefTrackers()
	want := []string{"udp://remote.example:1/announce", "udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGetDefTrackersNeverBlocksNewTorrentPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(trackersFetchTimeout + 2*time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://should-not-appear:1/announce\n"))
	}))
	defer srv.Close()

	setupTrackersTest(t, srv.URL, "udp://instant.example:80/announce")

	const iterations = 5
	for i := 0; i < iterations; i++ {
		start := time.Now()
		got := GetDefTrackers()
		if time.Since(start) > 200*time.Millisecond {
			t.Fatalf("iteration %d: GetDefTrackers blocked for %v", i, time.Since(start))
		}
		want := []string{"udp://instant.example:80/announce"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("iteration %d: got %#v, want %#v", i, got, want)
		}
	}
}

func TestTrackersPeriodicRefresh(t *testing.T) {
	var version atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := version.Add(1)
		w.WriteHeader(http.StatusOK)
		if v == 1 {
			_, _ = w.Write([]byte("udp://remote-v1.example:1/announce\n"))
			return
		}
		_, _ = w.Write([]byte("udp://remote-v2.example:2/announce\n"))
	}))
	defer srv.Close()

	setupTrackersTest(t, srv.URL, "udp://local.example:80/announce")
	waitForTrackersFetch(t, 2*time.Second)

	wantV1 := []string{"udp://remote-v1.example:1/announce", "udp://local.example:80/announce"}
	if got := GetDefTrackers(); !reflect.DeepEqual(got, wantV1) {
		t.Fatalf("initial got %#v, want %#v", got, wantV1)
	}

	refreshTrackers()

	wantV2 := []string{"udp://remote-v2.example:2/announce", "udp://local.example:80/announce"}
	if got := GetDefTrackers(); !reflect.DeepEqual(got, wantV2) {
		t.Fatalf("after refresh got %#v, want %#v", got, wantV2)
	}
}

func TestStaleTrackersFetchDoesNotCancelCurrent(t *testing.T) {
	release := make(chan struct{})
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote.example:1/announce\n"))
	}))
	defer srv.Close()

	setupTrackersTest(t, srv.URL, "udp://local.example:80/announce")
	deadline := time.Now().Add(2 * time.Second)
	for hits.Load() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("prefetch did not reach the server")
		}
		time.Sleep(5 * time.Millisecond)
	}

	_, _, _, err := fetchTrackersFromURLs(context.Background(), trackersFetchGen.Load()-1, []string{srv.URL}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("stale fetch err = %v, want canceled", err)
	}

	close(release)
	waitForTrackersFetch(t, 2*time.Second)
	want := []string{"udp://remote.example:1/announce", "udp://local.example:80/announce"}
	if got := GetDefTrackers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestTrackersRefreshKeepsCacheOnFailure(t *testing.T) {
	var failRefresh atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failRefresh.Load() {
			http.Error(w, "blocked", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote.example:1/announce\n"))
	}))
	defer srv.Close()

	setupTrackersTest(t, srv.URL, "udp://local.example:80/announce")
	waitForTrackersFetch(t, 2*time.Second)

	want := []string{"udp://remote.example:1/announce", "udp://local.example:80/announce"}
	if got := GetDefTrackers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("before failed refresh got %#v, want %#v", got, want)
	}

	failRefresh.Store(true)
	local := configuredDefaultTrackers()
	_, _, _, err := fetchTrackersFromURLs(context.Background(), trackersFetchGen.Load(), []string{srv.URL}, local)
	if err == nil {
		t.Fatal("expected refresh fetch to fail")
	}

	if got := GetDefTrackers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("after failed refresh got %#v, want cached %#v", got, want)
	}
}

func TestTrackersRefreshFallsBackToNextURL(t *testing.T) {
	var failFirst atomic.Bool
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failFirst.Load() {
			http.Error(w, "blocked", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote-v1.example:1/announce\n"))
	}))
	defer srv1.Close()

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote-v2.example:2/announce\n"))
	}))
	defer srv2.Close()

	setupTrackersTestWithURLs(t, "", []string{srv1.URL, srv2.URL}, "udp://local.example:80/announce")
	waitForTrackersFetch(t, 2*time.Second)

	wantV1 := []string{"udp://remote-v1.example:1/announce", "udp://local.example:80/announce"}
	if got := GetDefTrackers(); !reflect.DeepEqual(got, wantV1) {
		t.Fatalf("initial got %#v, want %#v", got, wantV1)
	}

	failFirst.Store(true)
	refreshTrackers()

	wantV2 := []string{"udp://remote-v2.example:2/announce", "udp://local.example:80/announce"}
	if got := GetDefTrackers(); !reflect.DeepEqual(got, wantV2) {
		t.Fatalf("after refresh fallback got %#v, want %#v", got, wantV2)
	}
}

func TestTrackersRefreshAllFailKeepsCache(t *testing.T) {
	var failAll atomic.Bool
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failAll.Load() {
			http.Error(w, "blocked", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("udp://remote.example:1/announce\n"))
	}))
	defer srv1.Close()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer srv2.Close()

	setupTrackersTestWithURLs(t, "", []string{srv1.URL, srv2.URL}, "udp://local.example:80/announce")
	waitForTrackersFetch(t, 2*time.Second)

	want := []string{"udp://remote.example:1/announce", "udp://local.example:80/announce"}
	if got := GetDefTrackers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("before failed refresh got %#v, want %#v", got, want)
	}

	failAll.Store(true)
	refreshTrackers()

	if got := GetDefTrackers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("after all-URL refresh failure got %#v, want cached %#v", got, want)
	}
}

func TestGetDefTrackersEmptyURLNoMirrorsSkipsFetch(t *testing.T) {
	local := "udp://only-local:1337/announce"
	setupTrackersTest(t, "", local)

	got := GetDefTrackers()
	want := []string{"udp://only-local:1337/announce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestSanitizeTrackersSameIPAndRawHTTPS(t *testing.T) {
	got := SanitizeTrackers([]string{
		"udp://198.51.100.1:1337/announce",
		"http://198.51.100.1:1337/announce",
		"https://1.2.3.4/announce",
		"wss://203.0.113.8/announce",
		"http://retracker.mgts.by:80/announce",
		"ws://tracker.example/announce",
	})
	want := []string{
		"udp://198.51.100.1:1337/announce",
		"http://retracker.mgts.by:80/announce",
		"ws://tracker.example/announce",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestSanitizeTrackersKeepsRemoteAndLocal(t *testing.T) {
	var remote []string
	for i := 1; i <= 20; i++ {
		remote = append(remote, fmt.Sprintf("  udp://198.51.100.%d:1337/announce/  ", i))
	}
	remote = append(remote,
		"udp://198.51.100.1:1337/announce",
		"http://198.51.100.1:1337/announce",
		"https://198.51.100.50:443/announce",
		"https://1.2.3.4/announce",
		"wss://203.0.113.8:443/announce",
		"# comment",
		"",
		"   ",
		"not a url",
		"ftp://files.example/announce",
		"http://",
		"http://tracker.example:99999/announce",
		"udp://tracker.example:1337/announce",
		"udp://tracker.example:1337/announce/",
	)
	var local []string
	for i := 1; i <= 14; i++ {
		local = append(local, fmt.Sprintf("udp://local%d.example:80/announce", i))
	}
	merged, remoteCount := mergeTrackers(remote, local)
	if remoteCount != 21 || len(merged) != 35 {
		t.Fatalf("remote %d, len %d, want 21 remote + 14 local", remoteCount, len(merged))
	}
	for _, tr := range merged {
		u := strings.ToLower(tr)
		if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "wss://") {
			if strings.Contains(tr, "198.51.100.") || strings.Contains(tr, "1.2.3.4") || strings.Contains(tr, "203.0.113.") {
				t.Fatalf("raw-ip tls tracker kept: %s", tr)
			}
		}
	}
	seen := map[string]struct{}{}
	for _, tr := range merged {
		if _, ok := seen[tr]; ok {
			t.Fatalf("duplicate %s", tr)
		}
		seen[tr] = struct{}{}
	}
	for i := 1; i <= 14; i++ {
		want := fmt.Sprintf("udp://local%d.example:80/announce", i)
		if _, ok := seen[want]; !ok {
			t.Fatalf("local tracker dropped: %s in %#v", want, merged)
		}
	}
	if _, ok := seen["udp://198.51.100.1:1337/announce"]; ok {
		if _, dup := seen["http://198.51.100.1:1337/announce"]; dup {
			t.Fatal("udp and http on the same ip:port both kept")
		}
	}
}

func TestGetDefTrackersMergesSanitizedRemote(t *testing.T) {
	body := strings.Join([]string{
		"# best",
		"udp://198.51.100.1:1337/announce",
		"  udp://198.51.100.1:1337/announce/  ",
		"http://198.51.100.1:1337/announce",
		"https://203.0.113.5:443/announce",
		"ftp://nope.example/announce",
		"",
		"udp://198.51.100.2:6969/announce",
		"udp://local.example:80/announce",
	}, "\n")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	local := "udp://local.example:80/announce\nhttps://1.2.3.4/announce\n# c\nnot-a-url\nwss://tracker.example/announce"
	setupTrackersTest(t, srv.URL, local)
	waitForTrackersFetch(t, 2*time.Second)

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

func TestFetchTrackersOversizedBody(t *testing.T) {
	body := "udp://198.51.100.9:1337/announce\n" + strings.Repeat("A", trackersBodyLimit*2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	setupTrackersTest(t, srv.URL, "udp://local.example:80/announce")
	start := time.Now()
	waitForTrackersFetch(t, 2*time.Second)
	if time.Since(start) > 2*time.Second {
		t.Fatalf("oversized body blocked for %s", time.Since(start))
	}
	got := GetDefTrackers()
	if len(got) != 2 {
		t.Fatalf("got %#v", got)
	}
	if got[0] != "udp://198.51.100.9:1337/announce" {
		t.Fatalf("got %#v", got)
	}
}

func TestFetchTrackersHTTP500And404(t *testing.T) {
	for _, code := range []int{http.StatusInternalServerError, http.StatusNotFound} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
			}))
			defer srv.Close()
			local := "udp://fallback.example:80/announce"
			setupTrackersTest(t, srv.URL, local)
			waitForTrackersFetch(t, 2*time.Second)
			got := GetDefTrackers()
			want := []string{local}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
		})
	}
}

func TestFetchTrackersContextCancel(t *testing.T) {
	started := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() {
		<-started
		cancel()
	}()
	start := time.Now()
	_, _, _, err := fetchTrackersFromURLs(ctx, trackersFetchGen.Load(), []string{srv.URL}, []string{"udp://local.example:1/announce"})
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if time.Since(start) > time.Second {
		t.Fatalf("cancel took %s", time.Since(start))
	}
}

func TestTrackersConcurrentReload(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := n.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "udp://198.51.100.%d:1337/announce\n", (id%20)+1)
	}))
	defer srv.Close()

	oldPath := settings.Path
	dir := t.TempDir()
	settings.Path = dir
	t.Cleanup(func() { settings.Path = oldPath })
	if err := os.WriteFile(filepath.Join(dir, "trackers.txt"), []byte("udp://file.example:1/announce\nhttps://1.2.3.4/announce\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	setupTrackersTest(t, srv.URL, "udp://local.example:80/announce")
	waitForTrackersFetch(t, 2*time.Second)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				_ = GetDefTrackers()
				_ = GetTrackerFromFile()
			}
		}()
	}
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 5 {
				refreshTrackers()
				InvalidateTrackersCache()
			}
		}()
	}
	wg.Wait()
}
