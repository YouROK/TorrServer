package playback

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestConservativeDefaults(t *testing.T) {
	if err := Init(t.TempDir(), false); err != nil {
		t.Fatal(err)
	}
	config := GetPublicConfig()
	if config.Enabled || config.RoutingMode != RoutingLocal || len(config.Devices) != 0 {
		t.Fatalf("unexpected defaults: %#v", config)
	}
	if _, err := ResolveTarget("anything"); !errors.Is(err, ErrRemotePlaybackDisabled) {
		t.Fatalf("remote routing should be disabled by default, got %v", err)
	}
}

func TestLegacyArrayMigrationStaysDisabled(t *testing.T) {
	dir := t.TempDir()
	legacy := []Device{{ID: "tv", Name: "TV", Endpoint: "http://127.0.0.1:9000"}}
	buf, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, configFileName), buf, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Init(dir, false); err != nil {
		t.Fatal(err)
	}
	config := GetPublicConfig()
	if config.Enabled || config.RoutingMode != RoutingLocal || len(config.Devices) != 1 {
		t.Fatalf("legacy migration enabled new behavior: %#v", config)
	}
}

func TestDevicePersistenceSettingsAndTokenRedaction(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir, false); err != nil {
		t.Fatal(err)
	}

	created, err := Upsert(DeviceInput{
		ID:            "living-room",
		Name:          "Living room",
		Endpoint:      "http://127.0.0.1:9080",
		Token:         "secret",
		StreamBaseURL: "http://192.168.1.10:8090",
		Fullscreen:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "living-room" || created.Name != "Living room" {
		t.Fatalf("unexpected public device: %#v", created)
	}

	public := GetPublicConfig()
	if len(public.Devices) != 1 || public.Devices[0].ID != "living-room" {
		t.Fatalf("unexpected public config: %#v", public)
	}
	managed := GetManagedConfig()
	if len(managed.Devices) != 1 || !managed.Devices[0].HasToken || !managed.Devices[0].Fullscreen {
		t.Fatalf("unexpected managed config: %#v", managed)
	}

	settings, err := UpdateSettings(Settings{
		Enabled:         true,
		RoutingMode:     RoutingPrimary,
		PrimaryDeviceID: "living-room",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !settings.Enabled || settings.RoutingMode != RoutingPrimary {
		t.Fatalf("settings were not saved: %#v", settings)
	}

	// Empty token on update preserves the existing secret.
	if _, err := Upsert(DeviceInput{
		ID:            "living-room",
		Name:          "TV",
		Endpoint:      "http://127.0.0.1:9080",
		StreamBaseURL: "http://192.168.1.10:8090",
		Fullscreen:    false,
	}); err != nil {
		t.Fatal(err)
	}
	managed = GetManagedConfig()
	if !managed.Devices[0].HasToken || managed.Devices[0].Name != "TV" || managed.Devices[0].Fullscreen {
		t.Fatalf("device update was not applied correctly: %#v", managed.Devices[0])
	}

	info, err := os.Stat(filepath.Join(dir, configFileName))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected config permissions: %o", info.Mode().Perm())
	}

	if err := Init(dir, false); err != nil {
		t.Fatal(err)
	}
	device, ok := getDeviceForTest("living-room")
	if !ok || device.Token != "secret" {
		t.Fatalf("token was not persisted: %#v", device)
	}
	reloaded := GetManagedConfig()
	if !reloaded.Enabled || reloaded.RoutingMode != RoutingPrimary || reloaded.PrimaryDeviceID != "living-room" {
		t.Fatalf("routing settings were not persisted: %#v", reloaded)
	}

	// A replacement token takes priority over a stale clear checkbox.
	if _, err := Upsert(DeviceInput{
		ID:            "living-room",
		Name:          "TV",
		Endpoint:      "http://127.0.0.1:9080",
		StreamBaseURL: "http://192.168.1.10:8090",
		Token:         "replacement",
		ClearToken:    true,
	}); err != nil {
		t.Fatal(err)
	}
	device, ok = getDeviceForTest("living-room")
	if !ok || device.Token != "replacement" {
		t.Fatalf("replacement token was discarded: %#v", device)
	}

	if _, err := Upsert(DeviceInput{
		ID:            "living-room",
		Name:          "TV",
		Endpoint:      "http://127.0.0.1:9080",
		StreamBaseURL: "http://192.168.1.10:8090",
		ClearToken:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if GetManagedConfig().Devices[0].HasToken {
		t.Fatal("clear_token did not remove the secret")
	}
}

func TestRoutingModesAndPrimaryDeletion(t *testing.T) {
	if err := Init(t.TempDir(), false); err != nil {
		t.Fatal(err)
	}
	for _, input := range []DeviceInput{
		{ID: "living-room", Name: "Living room", Endpoint: "http://127.0.0.1:9001"},
		{ID: "bedroom", Name: "Bedroom", Endpoint: "http://127.0.0.1:9002"},
	} {
		if _, err := Upsert(input); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := UpdateSettings(Settings{Enabled: true, RoutingMode: RoutingPrimary}); !errors.Is(err, ErrPrimaryDeviceNotSelected) {
		t.Fatalf("primary mode without a device should fail, got %v", err)
	}
	if _, err := UpdateSettings(Settings{
		Enabled:         true,
		RoutingMode:     RoutingPrimary,
		PrimaryDeviceID: "missing",
	}); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unknown primary device should fail, got %v", err)
	}

	if _, err := UpdateSettings(Settings{
		Enabled:         true,
		RoutingMode:     RoutingPrimary,
		PrimaryDeviceID: "living-room",
	}); err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveTarget("bedroom")
	if err != nil || resolved != "living-room" {
		t.Fatalf("browser overrode primary target: %q, %v", resolved, err)
	}

	if _, err := UpdateSettings(Settings{Enabled: true, RoutingMode: RoutingPerBrowser}); err != nil {
		t.Fatal(err)
	}
	resolved, err = ResolveTarget("bedroom")
	if err != nil || resolved != "bedroom" {
		t.Fatalf("per-browser target was not honored: %q, %v", resolved, err)
	}

	if _, err := UpdateSettings(Settings{Enabled: true, RoutingMode: RoutingLocal}); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveTarget("bedroom"); !errors.Is(err, ErrRemotePlaybackDisabled) {
		t.Fatalf("local mode should reject remote routing, got %v", err)
	}

	if _, err := UpdateSettings(Settings{
		Enabled:         true,
		RoutingMode:     RoutingPrimary,
		PrimaryDeviceID: "living-room",
	}); err != nil {
		t.Fatal(err)
	}
	if err := Delete("living-room"); err != nil {
		t.Fatal(err)
	}
	settings := GetPublicConfig().Settings
	if settings.Enabled || settings.RoutingMode != RoutingLocal || settings.PrimaryDeviceID != "" {
		t.Fatalf("deleting primary device did not restore conservative mode: %#v", settings)
	}
}

func TestValidation(t *testing.T) {
	if err := Init(t.TempDir(), false); err != nil {
		t.Fatal(err)
	}

	cases := []DeviceInput{
		{ID: "bad id", Name: "TV", Endpoint: "http://127.0.0.1:9000"},
		{ID: "tv", Name: "", Endpoint: "http://127.0.0.1:9000"},
		{ID: "tv", Name: "TV", Endpoint: "file:///tmp/socket"},
		{ID: "tv", Name: "TV", Endpoint: "http://user:pass@127.0.0.1:9000"},
		{ID: "tv", Name: "TV", Endpoint: "http://127.0.0.1:9000?x=1"},
		{ID: "tv", Name: "TV", Endpoint: "http://127.0.0.1:9000", StreamBaseURL: "ftp://host"},
	}
	for _, input := range cases {
		if _, err := Upsert(input); err == nil {
			t.Fatalf("expected validation error for %#v", input)
		}
	}
	if _, err := UpdateSettings(Settings{Enabled: true, RoutingMode: "unknown"}); err == nil {
		t.Fatal("invalid routing mode was accepted")
	}
}

func TestAgentRequestsIncludeFullscreen(t *testing.T) {
	var mu sync.Mutex
	requests := make([]string, 0, 2)

	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer agent-secret" {
			t.Errorf("unexpected authorization: %q", got)
		}
		mu.Lock()
		requests = append(requests, r.Method+" "+r.URL.Path)
		mu.Unlock()

		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
		case "/play":
			var payload AgentPlayRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode play request: %v", err)
			}
			if payload.StreamURL != "http://ts.local/stream/Movie.mkv?link=hash&index=1&play" {
				t.Errorf("unexpected stream URL: %q", payload.StreamURL)
			}
			if !payload.Fullscreen {
				t.Error("device fullscreen preference was not sent")
			}
			w.WriteHeader(http.StatusAccepted)
		default:
			http.NotFound(w, r)
		}
	}))
	defer agent.Close()

	if err := Init(t.TempDir(), false); err != nil {
		t.Fatal(err)
	}
	if _, err := Upsert(DeviceInput{
		ID:         "tv",
		Name:       "TV",
		Endpoint:   agent.URL,
		Token:      "agent-secret",
		Fullscreen: true,
	}); err != nil {
		t.Fatal(err)
	}

	if err := Test("tv"); err != nil {
		t.Fatal(err)
	}
	if err := Play("tv", AgentPlayRequest{
		Path:      "Movie.mkv",
		Hash:      "hash",
		Index:     1,
		StreamURL: "http://ts.local/stream/Movie.mkv?link=hash&index=1&play",
	}); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 2 || requests[0] != "GET /health" || requests[1] != "POST /play" {
		t.Fatalf("unexpected requests: %#v", requests)
	}
}

func getDeviceForTest(id string) (Device, bool) {
	global.mu.RLock()
	defer global.mu.RUnlock()
	device, ok := global.devices[id]
	return device, ok
}
