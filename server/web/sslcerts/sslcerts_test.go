package sslcerts

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"server/settings"
)

func withTempPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := settings.Path
	settings.Path = dir
	t.Cleanup(func() { settings.Path = old })
	return dir
}

func leaf(t *testing.T, certFile, keyFile string) *x509.Certificate {
	t.Helper()
	pair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	c, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestEnsureCertGeneratesWhenEmpty(t *testing.T) {
	dir := withTempPath(t)

	cert, key, changed, err := EnsureCert("", "", []string{"192.168.1.10", "not-an-ip"})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}
	if want := filepath.Join(dir, certFileName); !samePath(cert, want) {
		t.Fatalf("cert = %q, want %q", cert, want)
	}
	if !IsGenerated(cert, key) {
		t.Fatal("IsGenerated = false for generated pair")
	}

	c := leaf(t, cert, key)
	if err := c.VerifyHostname("192.168.1.10"); err != nil {
		t.Errorf("LAN IP not in SANs: %v", err)
	}
	if err := c.VerifyHostname("127.0.0.1"); err != nil {
		t.Errorf("loopback not in SANs: %v", err)
	}
	if err := c.VerifyHostname("localhost"); err != nil {
		t.Errorf("localhost not in SANs: %v", err)
	}
	if c.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		t.Error("ECDSA cert should not have KeyEncipherment usage")
	}
}

func TestGeneratedKeyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions only")
	}
	withTempPath(t)

	_, key, err := MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(key)
	if err != nil {
		t.Fatal(err)
	}
	if perm := st.Mode().Perm(); perm != 0o600 {
		t.Fatalf("key perm = %o, want 600", perm)
	}
}

func TestEnsureCertTightensLegacyKeyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions only")
	}
	withTempPath(t)

	cert, key, err := MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(key, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, changed, err := EnsureCert(cert, key, nil); err != nil || changed {
		t.Fatalf("EnsureCert = changed %v, err %v; want unchanged, nil", changed, err)
	}
	st, _ := os.Stat(key)
	if perm := st.Mode().Perm(); perm != 0o600 {
		t.Fatalf("key perm = %o, want 600", perm)
	}
}

func TestEnsureCertKeepsValidUserCert(t *testing.T) {
	withTempPath(t)
	userDir := t.TempDir()

	// build a valid pair outside settings.Path, i.e. "user supplied"
	certPEM, keyPEM, err := generateSelfSignedCert(nil)
	if err != nil {
		t.Fatal(err)
	}
	cert := filepath.Join(userDir, "fullchain.pem")
	key := filepath.Join(userDir, "privkey.pem")
	os.WriteFile(cert, certPEM, 0o644)
	os.WriteFile(key, keyPEM, 0o600)

	gotCert, gotKey, changed, err := EnsureCert(cert, key, nil)
	if err != nil {
		t.Fatal(err)
	}
	if changed || gotCert != cert || gotKey != key {
		t.Fatalf("user cert replaced: %q %q changed=%v", gotCert, gotKey, changed)
	}
}

func TestEnsureCertNeverReplacesInvalidUserCert(t *testing.T) {
	dir := withTempPath(t)
	userDir := t.TempDir()
	cert := filepath.Join(userDir, "fullchain.pem")
	key := filepath.Join(userDir, "privkey.pem")
	os.WriteFile(cert, []byte("garbage"), 0o644)
	os.WriteFile(key, []byte("garbage"), 0o600)

	_, _, changed, err := EnsureCert(cert, key, nil)
	if err == nil {
		t.Fatal("expected error for invalid user cert")
	}
	if changed {
		t.Fatal("changed = true for invalid user cert")
	}
	if _, err := os.Stat(filepath.Join(dir, certFileName)); !os.IsNotExist(err) {
		t.Fatal("self-signed cert was generated as a fallback for a user cert")
	}
}

func TestEnsureCertMissingUserCert(t *testing.T) {
	withTempPath(t)
	_, _, _, err := EnsureCert("/nonexistent/cert.pem", "/nonexistent/key.pem", nil)
	if err == nil {
		t.Fatal("expected error for missing user cert")
	}
}

func TestEnsureCertRegeneratesBrokenGeneratedCert(t *testing.T) {
	withTempPath(t)

	cert, key, err := MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cert, []byte("truncated"), 0o644); err != nil {
		t.Fatal(err)
	}

	gotCert, gotKey, changed, err := EnsureCert(cert, key, nil)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("paths are identical, changed should be false")
	}
	if err := VerifyCertKeyFiles(gotCert, gotKey); err != nil {
		t.Fatalf("regenerated pair invalid: %v", err)
	}
}
