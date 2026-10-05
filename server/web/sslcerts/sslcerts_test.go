package sslcerts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

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

func TestEnsureCertRenewsWhenIPMoved(t *testing.T) {
	withTempPath(t)
	cert, key, err := MakeCertKeyFiles([]string{"10.0.0.5"})
	if err != nil {
		t.Fatal(err)
	}

	// DHCP moved the host: none of the current IPs is covered
	if _, _, _, err := EnsureCert(cert, key, []string{"192.168.7.7"}); err != nil {
		t.Fatal(err)
	}
	c := leaf(t, cert, key)
	for _, ip := range []string{"192.168.7.7", "10.0.0.5"} {
		if err := c.VerifyHostname(ip); err != nil {
			t.Errorf("%s not in SANs after renewal: %v", ip, err)
		}
	}
}

func TestEnsureCertRenewsNearExpiry(t *testing.T) {
	withTempPath(t)
	old := selfSignedValidity
	selfSignedValidity = 24 * time.Hour
	cert, key, err := MakeCertKeyFiles(nil)
	selfSignedValidity = old
	if err != nil {
		t.Fatal(err)
	}

	if _, _, _, err := EnsureCert(cert, key, nil); err != nil {
		t.Fatal(err)
	}
	if left := time.Until(leaf(t, cert, key).NotAfter); left < 300*24*time.Hour {
		t.Fatalf("cert not renewed, %s left", left)
	}
}

func TestEnsureCertKeepsHealthyGeneratedCert(t *testing.T) {
	withTempPath(t)
	cert, key, err := MakeCertKeyFiles([]string{"10.0.0.5"})
	if err != nil {
		t.Fatal(err)
	}
	before := leaf(t, cert, key).SerialNumber

	if _, _, _, err := EnsureCert(cert, key, []string{"10.0.0.5"}); err != nil {
		t.Fatal(err)
	}
	if leaf(t, cert, key).SerialNumber.Cmp(before) != 0 {
		t.Fatal("healthy cert was regenerated")
	}
}

func TestLoaderReloadsChangedFiles(t *testing.T) {
	withTempPath(t)
	cert, key, err := MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	l, err := NewLoader(func() (string, string) { return cert, key })
	if err != nil {
		t.Fatal(err)
	}
	first, _ := l.GetCertificate(nil)

	// regenerate in place, bump mtime in case the FS has coarse timestamps
	if _, _, err := MakeCertKeyFiles(nil); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Minute)
	os.Chtimes(cert, future, future)
	l.lastCheck = time.Time{}

	second, _ := l.GetCertificate(nil)
	if second.Leaf.SerialNumber.Cmp(first.Leaf.SerialNumber) == 0 {
		t.Fatal("loader did not pick up the new certificate")
	}
}

func TestLoaderKeepsPreviousOnBrokenFile(t *testing.T) {
	withTempPath(t)
	cert, key, err := MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	l, err := NewLoader(func() (string, string) { return cert, key })
	if err != nil {
		t.Fatal(err)
	}
	first, _ := l.GetCertificate(nil)

	os.WriteFile(cert, []byte("garbage"), 0o644)
	future := time.Now().Add(time.Minute)
	os.Chtimes(cert, future, future)
	l.lastCheck = time.Time{}

	got, err := l.GetCertificate(nil)
	if err != nil || got != first {
		t.Fatalf("got %p err %v, want previous cert %p", got, err, first)
	}
}

func TestLoaderFollowsPathChange(t *testing.T) {
	withTempPath(t)
	cert, key, err := MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	paths := [2]string{cert, key}
	l, err := NewLoader(func() (string, string) { return paths[0], paths[1] })
	if err != nil {
		t.Fatal(err)
	}

	userDir := t.TempDir()
	certPEM, keyPEM, _ := generateSelfSignedCert(nil)
	paths = [2]string{filepath.Join(userDir, "c.pem"), filepath.Join(userDir, "k.pem")}
	os.WriteFile(paths[0], certPEM, 0o644)
	os.WriteFile(paths[1], keyPEM, 0o600)
	l.lastCheck = time.Time{}

	got, _ := l.GetCertificate(nil)
	want := leaf(t, paths[0], paths[1])
	if got.Leaf.SerialNumber.Cmp(want.SerialNumber) != 0 {
		t.Fatal("loader did not switch to the new cert path")
	}
}

// writeUserCert writes a CA-signed pair (like a Let's Encrypt cert) and returns its paths.
func writeUserCert(t *testing.T, certFile, keyFile string) {
	t.Helper()
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Example CA"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(90 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, ca, ca, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	ca, _ = x509.ParseCertificate(caDER)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "tv.example.com"},
		DNSNames:     []string{"tv.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(90 * 24 * time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, leafTmpl, ca, &key.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, _ := x509.MarshalECPrivateKey(key)
	chain := append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})...)
	os.WriteFile(certFile, chain, 0o644)
	os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600)
}

func TestUserCertAtDefaultPathIsNeverReplaced(t *testing.T) {
	dir := withTempPath(t)
	cert, key := filepath.Join(dir, certFileName), filepath.Join(dir, keyFileName)
	writeUserCert(t, cert, key)
	before, _ := os.ReadFile(cert)

	if IsGenerated(cert, key) {
		t.Fatal("user cert at the default path treated as generated")
	}
	// hostname and IPs are missing from the user cert: must not trigger a "renewal"
	gotCert, gotKey, changed, err := EnsureCert(cert, key, []string{"192.168.1.10"})
	if err != nil || changed || gotCert != cert || gotKey != key {
		t.Fatalf("EnsureCert = %q %q changed=%v err=%v", gotCert, gotKey, changed, err)
	}
	if after, _ := os.ReadFile(cert); !bytes.Equal(before, after) {
		t.Fatal("user cert at the default path was overwritten")
	}
}

func TestGarbageAtDefaultPathIsLeftAlone(t *testing.T) {
	dir := withTempPath(t)
	cert, key := filepath.Join(dir, certFileName), filepath.Join(dir, keyFileName)
	os.WriteFile(cert, []byte("not a cert"), 0o644)
	os.WriteFile(key, []byte("not a key"), 0o600)

	_, _, _, err := EnsureCert(cert, key, nil)
	if err == nil || !strings.Contains(err.Error(), "delete both files") {
		t.Fatalf("err = %v, want error with delete hint", err)
	}
	if b, _ := os.ReadFile(cert); string(b) != "not a cert" {
		t.Fatal("unknown file at the default path was overwritten")
	}
}

func TestEnsureCertRegeneratesExpiredGeneratedCert(t *testing.T) {
	withTempPath(t)
	old := selfSignedValidity
	selfSignedValidity = 30 * time.Minute // NotBefore is backdated 1h, so this is already expired
	cert, key, err := MakeCertKeyFiles(nil)
	selfSignedValidity = old
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyCertKeyFiles(cert, key); err == nil {
		t.Fatal("test setup: cert should be expired")
	}

	if _, _, _, err := EnsureCert(cert, key, nil); err != nil {
		t.Fatal(err)
	}
	if err := VerifyCertKeyFiles(cert, key); err != nil {
		t.Fatalf("regenerated pair invalid: %v", err)
	}
}

func TestEnsureCertRegeneratesMissingGeneratedFiles(t *testing.T) {
	withTempPath(t)
	cert, key, err := MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}

	// our cert without its key (e.g. interrupted write): regenerated
	os.Remove(key)
	if _, _, _, err := EnsureCert(cert, key, nil); err != nil {
		t.Fatal(err)
	}
	if err := VerifyCertKeyFiles(cert, key); err != nil {
		t.Fatalf("regenerated pair invalid: %v", err)
	}

	// both deleted by the user to force a new one: regenerated
	os.Remove(cert)
	os.Remove(key)
	if _, _, _, err := EnsureCert(cert, key, nil); err != nil {
		t.Fatal(err)
	}
	if err := VerifyCertKeyFiles(cert, key); err != nil {
		t.Fatalf("regenerated pair invalid: %v", err)
	}
}

func TestLoneKeyAtDefaultPathIsKept(t *testing.T) {
	dir := withTempPath(t)
	key := filepath.Join(dir, keyFileName)
	os.WriteFile(key, []byte("someone's key"), 0o600)

	if _, _, _, err := EnsureCert("", "", nil); err == nil {
		t.Fatal("expected error with a lone key at the default path")
	}
	if b, _ := os.ReadFile(key); string(b) != "someone's key" {
		t.Fatal("lone key was overwritten")
	}
}

func TestEnsureCertRenewsWhenIPChangesBesideStableInterface(t *testing.T) {
	withTempPath(t)
	// e.g. a Tailscale IP that never changes plus a DHCP LAN address
	cert, key, err := MakeCertKeyFiles([]string{"100.115.156.3", "192.168.0.169"})
	if err != nil {
		t.Fatal(err)
	}

	if _, _, _, err := EnsureCert(cert, key, []string{"100.115.156.3", "192.168.0.200"}); err != nil {
		t.Fatal(err)
	}
	c := leaf(t, cert, key)
	for _, ip := range []string{"100.115.156.3", "192.168.0.200", "192.168.0.169"} {
		if err := c.VerifyHostname(ip); err != nil {
			t.Errorf("%s not in SANs after renewal: %v", ip, err)
		}
	}
}

func TestIsGeneratedRecognisesLegacyCert(t *testing.T) {
	dir := withTempPath(t)
	// shape of certs from older versions: O=TorrServer, no CN, P-384
	priv, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(7),
		Subject:               pkix.Name{Organization: []string{"TorrServer"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	cert, key := filepath.Join(dir, certFileName), filepath.Join(dir, keyFileName)
	os.WriteFile(cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644)
	if !IsGenerated(cert, key) {
		t.Fatal("legacy TorrServer cert not recognised")
	}
}

func TestFirstStartAdoptsUserCertAtDefaultPath(t *testing.T) {
	dir := withTempPath(t)
	cert, key := filepath.Join(dir, certFileName), filepath.Join(dir, keyFileName)
	writeUserCert(t, cert, key)
	before, _ := os.ReadFile(cert)

	// no paths configured yet (fresh settings)
	gotCert, gotKey, changed, err := EnsureCert("", "", []string{"192.168.1.10"})
	if err != nil {
		t.Fatal(err)
	}
	if !changed || !samePath(gotCert, cert) || !samePath(gotKey, key) {
		t.Fatalf("got %q %q changed=%v, want the existing default-path pair", gotCert, gotKey, changed)
	}
	if after, _ := os.ReadFile(cert); !bytes.Equal(before, after) {
		t.Fatal("user cert at the default path was overwritten on first start")
	}
}

func TestFirstStartKeepsInvalidUserFiles(t *testing.T) {
	dir := withTempPath(t)
	cert, key := filepath.Join(dir, certFileName), filepath.Join(dir, keyFileName)
	os.WriteFile(cert, []byte("not a cert"), 0o644)
	os.WriteFile(key, []byte("not a key"), 0o600)

	if _, _, _, err := EnsureCert("", "", nil); err == nil {
		t.Fatal("expected error for invalid files at the default path")
	}
	if b, _ := os.ReadFile(cert); string(b) != "not a cert" {
		t.Fatal("invalid user file was overwritten on first start")
	}
}

func TestMakeCertKeyFilesRefusesToOverwriteUserCert(t *testing.T) {
	dir := withTempPath(t)
	cert, key := filepath.Join(dir, certFileName), filepath.Join(dir, keyFileName)
	writeUserCert(t, cert, key)
	before, _ := os.ReadFile(cert)

	if _, _, err := MakeCertKeyFiles(nil); err == nil {
		t.Fatal("MakeCertKeyFiles overwrote a user cert")
	}
	if after, _ := os.ReadFile(cert); !bytes.Equal(before, after) {
		t.Fatal("user cert changed")
	}
}
