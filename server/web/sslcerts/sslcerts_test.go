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
