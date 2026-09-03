package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func withWAFSettingsPath(t *testing.T) string {
	t.Helper()
	oldPath := Path
	oldReadOnly := ReadOnly
	Path = t.TempDir()
	ReadOnly = false
	t.Cleanup(func() {
		Path = oldPath
		ReadOnly = oldReadOnly
	})
	return filepath.Join(Path, "settings.json")
}

func TestWAFConfigAbsentAndExplicitlyEmpty(t *testing.T) {
	withWAFSettingsPath(t)

	cfg, ok, err := GetWAFConfig()
	if err != nil || ok || cfg.Version != WAFConfigVersion {
		t.Fatalf("absent config = %+v, ok=%v, err=%v", cfg, ok, err)
	}
	if err := SetWAFConfig(WAFConfig{}); err != nil {
		t.Fatal(err)
	}
	cfg, ok, err = GetWAFConfig()
	if err != nil || !ok || cfg.Version != WAFConfigVersion {
		t.Fatalf("empty config = %+v, ok=%v, err=%v", cfg, ok, err)
	}
}

func TestWAFConfigRoundTripPreservesSettings(t *testing.T) {
	filename := withWAFSettingsPath(t)
	if err := os.WriteFile(filename, []byte(`{"BitTorr":{"EnableDebug":true},"other":"keep","large":9007199254740993}`), 0o600); err != nil {
		t.Fatal(err)
	}
	want := WAFConfig{
		Whitelist: []string{"1.2.3.4", "10.0.0.0/8"},
		Blacklist: []string{"5.6.7.8"},
		Referers:  []string{"example.com"},
	}
	if err := SetWAFConfig(want); err != nil {
		t.Fatal(err)
	}

	fileData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(fileData, &root); err != nil {
		t.Fatal(err)
	}
	if _, ok := root["BitTorr"]; !ok {
		t.Fatal("BitTorr settings were removed")
	}
	if string(root["other"]) != `"keep"` {
		t.Fatalf("unrelated setting changed: %s", root["other"])
	}
	if string(root["large"]) != "9007199254740993" {
		t.Fatalf("unrelated numeric setting changed: %s", root["large"])
	}
	if _, ok := root[wafConfigKey]; !ok {
		t.Fatalf("%q is not a top-level settings.json object", wafConfigKey)
	}

	got, ok, err := GetWAFConfig()
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if !slices.Equal(got.Whitelist, want.Whitelist) ||
		!slices.Equal(got.Blacklist, want.Blacklist) ||
		!slices.Equal(got.Referers, want.Referers) {
		t.Fatalf("got=%+v want=%+v", got, want)
	}
}

func TestWAFConfigRejectsStringLists(t *testing.T) {
	filename := withWAFSettingsPath(t)
	if err := os.WriteFile(filename, []byte(`{"waf":{"version":1,"whitelist":"127.0.0.1","blacklist":[],"referers":[]}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := GetWAFConfig(); err == nil {
		t.Fatal("expected string list to be rejected")
	}
}

func TestWAFConfigErrorsAndReadOnly(t *testing.T) {
	filename := withWAFSettingsPath(t)
	if err := os.WriteFile(filename, []byte(`{"waf":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := GetWAFConfig(); err == nil {
		t.Fatal("expected malformed settings.json read error")
	}
	if err := SetWAFConfig(WAFConfig{}); err == nil {
		t.Fatal("expected malformed settings.json write error")
	}

	if err := os.WriteFile(filename, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	ReadOnly = true
	if err := SetWAFConfig(WAFConfig{}); err == nil {
		t.Fatal("expected read-only error")
	}
}
