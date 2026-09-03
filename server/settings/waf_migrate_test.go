package settings

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestMigrateWAFListsNoOpWithoutFiles(t *testing.T) {
	withWAFSettingsPath(t)

	MigrateWAFLists()

	_, ok, err := GetWAFConfig()
	if err != nil || ok {
		t.Fatalf("expected no waf key, ok=%v err=%v", ok, err)
	}
}

func TestMigrateWAFListsImportsAndRenames(t *testing.T) {
	withWAFSettingsPath(t)

	wip := filepath.Join(Path, legacyWIPFile)
	bip := filepath.Join(Path, legacyBIPFile)
	wipBody := "# comment\nlocal:127.0.0.1\n\n10.0.0.0-10.0.0.255\n"
	bipBody := "203.0.113.1\n"
	if err := os.WriteFile(wip, []byte(wipBody), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bip, []byte(bipBody), 0o600); err != nil {
		t.Fatal(err)
	}

	MigrateWAFLists()

	cfg, ok, err := GetWAFConfig()
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	wantWhite := []string{"# comment", "local:127.0.0.1", "10.0.0.0-10.0.0.255"}
	wantBlack := []string{"203.0.113.1"}
	if !slices.Equal(cfg.Whitelist, wantWhite) {
		t.Fatalf("whitelist=%v want %v", cfg.Whitelist, wantWhite)
	}
	if !slices.Equal(cfg.Blacklist, wantBlack) {
		t.Fatalf("blacklist=%v want %v", cfg.Blacklist, wantBlack)
	}
	if len(cfg.Referers) != 0 {
		t.Fatalf("referers=%v want empty", cfg.Referers)
	}

	if fileExists(wip) {
		t.Fatal("wip.txt should be renamed away")
	}
	if fileExists(bip) {
		t.Fatal("bip.txt should be renamed away")
	}
	if !fileExists(wip + ".bak") {
		t.Fatal("expected wip.txt.bak")
	}
	if !fileExists(bip + ".bak") {
		t.Fatal("expected bip.txt.bak")
	}
	bak, err := os.ReadFile(wip + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if string(bak) != wipBody {
		t.Fatalf("wip backup content mismatch: %q", bak)
	}
}

func TestMigrateWAFListsOnlyWIP(t *testing.T) {
	withWAFSettingsPath(t)
	if err := os.WriteFile(filepath.Join(Path, legacyWIPFile), []byte("127.0.0.1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	MigrateWAFLists()

	cfg, ok, err := GetWAFConfig()
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if !slices.Equal(cfg.Whitelist, []string{"127.0.0.1"}) || len(cfg.Blacklist) != 0 {
		t.Fatalf("cfg=%+v", cfg)
	}
	if fileExists(filepath.Join(Path, legacyWIPFile)) || !fileExists(filepath.Join(Path, legacyWIPFile+".bak")) {
		t.Fatal("expected only wip.txt.bak")
	}
}

func TestMigrateWAFListsSkipsWhenWAFAlreadyPresent(t *testing.T) {
	withWAFSettingsPath(t)
	if err := SetWAFConfig(WAFConfig{Whitelist: []string{"8.8.8.8"}}); err != nil {
		t.Fatal(err)
	}
	wip := filepath.Join(Path, legacyWIPFile)
	if err := os.WriteFile(wip, []byte("127.0.0.1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	MigrateWAFLists()

	cfg, ok, err := GetWAFConfig()
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if !slices.Equal(cfg.Whitelist, []string{"8.8.8.8"}) {
		t.Fatalf("config overwritten: %+v", cfg)
	}
	if !fileExists(wip) {
		t.Fatal("legacy file should be left alone when waf already exists")
	}
	if fileExists(wip + ".bak") {
		t.Fatal("should not create backup when migration is skipped")
	}
}

func TestMigrateWAFListsReadOnly(t *testing.T) {
	withWAFSettingsPath(t)
	wip := filepath.Join(Path, legacyWIPFile)
	if err := os.WriteFile(wip, []byte("127.0.0.1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ReadOnly = true

	MigrateWAFLists()

	_, ok, err := GetWAFConfig()
	if err != nil || ok {
		t.Fatalf("expected no waf write in read-only, ok=%v err=%v", ok, err)
	}
	if !fileExists(wip) {
		t.Fatal("legacy file should remain in read-only mode")
	}
}

func TestMigrateWAFListsWriteFailureKeepsFiles(t *testing.T) {
	withWAFSettingsPath(t)
	wip := filepath.Join(Path, legacyWIPFile)
	if err := os.WriteFile(wip, []byte("127.0.0.1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(Path, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(Path, 0o755) })

	MigrateWAFLists()

	if !fileExists(wip) {
		t.Fatal("source file must not be renamed when write fails")
	}
	if fileExists(wip + ".bak") {
		t.Fatal("backup must not be created when write fails")
	}
	_, ok, err := GetWAFConfig()
	if err == nil && ok {
		t.Fatal("waf should not have been written when directory is not writable")
	}
}

func TestMigrateWAFListsReplacesExistingBak(t *testing.T) {
	withWAFSettingsPath(t)
	wip := filepath.Join(Path, legacyWIPFile)
	bak := wip + ".bak"
	if err := os.WriteFile(wip, []byte("1.1.1.1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bak, []byte("old-backup\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	MigrateWAFLists()

	if fileExists(wip) {
		t.Fatal("wip.txt should be gone")
	}
	data, err := os.ReadFile(bak)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "1.1.1.1\n" {
		t.Fatalf("bak should be replaced with current wip content, got %q", data)
	}
}
