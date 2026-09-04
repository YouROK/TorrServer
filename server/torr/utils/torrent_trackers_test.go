package utils

import (
	"net/http"
	"net/http/httptest"
	"reflect"
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
	settings.BTsets = nil
	trackersMu.Lock()
	loadedTrackers = nil
	trackersMu.Unlock()
	trackersFetchGen.Add(1)
	prefetchMu.Lock()
	prefetchStartedGen = ^uint64(0)
	prefetchMu.Unlock()
	refreshLoopOnce = sync.Once{}
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
	oldInterval := trackersRefreshInterval
	trackersRefreshInterval = 40 * time.Millisecond
	t.Cleanup(func() { trackersRefreshInterval = oldInterval })

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
	refreshLoopOnce = sync.Once{}
	PrefetchTrackers()

	waitForTrackersFetch(t, 2*time.Second)
	got := GetDefTrackers()
	wantV1 := []string{"udp://remote-v1.example:1/announce", "udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, wantV1) {
		t.Fatalf("initial got %#v, want %#v", got, wantV1)
	}

	deadline := time.Now().Add(2 * time.Second)
	wantV2 := []string{"udp://remote-v2.example:2/announce", "udp://local.example:80/announce"}
	for time.Now().Before(deadline) {
		got = GetDefTrackers()
		if reflect.DeepEqual(got, wantV2) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("after refresh got %#v, want %#v", got, wantV2)
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
	_, _, err := fetchTrackersFromURLs([]string{srv.URL}, local)
	if err == nil {
		t.Fatal("expected refresh fetch to fail")
	}

	if got := GetDefTrackers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("after failed refresh got %#v, want cached %#v", got, want)
	}
}

func TestTrackersRefreshFallsBackToNextURL(t *testing.T) {
	oldInterval := trackersRefreshInterval
	trackersRefreshInterval = 40 * time.Millisecond
	t.Cleanup(func() { trackersRefreshInterval = oldInterval })

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
	refreshLoopOnce = sync.Once{}
	PrefetchTrackers()

	waitForTrackersFetch(t, 2*time.Second)
	got := GetDefTrackers()
	wantV1 := []string{"udp://remote-v1.example:1/announce", "udp://local.example:80/announce"}
	if !reflect.DeepEqual(got, wantV1) {
		t.Fatalf("initial got %#v, want %#v", got, wantV1)
	}

	failFirst.Store(true)
	deadline := time.Now().Add(2 * time.Second)
	wantV2 := []string{"udp://remote-v2.example:2/announce", "udp://local.example:80/announce"}
	for time.Now().Before(deadline) {
		got = GetDefTrackers()
		if reflect.DeepEqual(got, wantV2) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("after refresh fallback got %#v, want %#v", got, wantV2)
}

func TestTrackersRefreshAllFailKeepsCache(t *testing.T) {
	oldInterval := trackersRefreshInterval
	trackersRefreshInterval = 40 * time.Millisecond
	t.Cleanup(func() { trackersRefreshInterval = oldInterval })

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
	refreshLoopOnce = sync.Once{}
	PrefetchTrackers()
	waitForTrackersFetch(t, 2*time.Second)

	want := []string{"udp://remote.example:1/announce", "udp://local.example:80/announce"}
	if got := GetDefTrackers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("before failed refresh got %#v, want %#v", got, want)
	}

	failAll.Store(true)
	time.Sleep(200 * time.Millisecond)

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
