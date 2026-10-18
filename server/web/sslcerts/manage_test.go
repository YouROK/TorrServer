package sslcerts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveUploadedAndInspect(t *testing.T) {
	dir := withTempPath(t)
	src := t.TempDir()
	srcCert, srcKey := filepath.Join(src, "c.pem"), filepath.Join(src, "k.pem")
	writeUserCert(t, srcCert, srcKey)
	certPEM, _ := os.ReadFile(srcCert)
	keyPEM, _ := os.ReadFile(srcKey)

	cert, key, err := SaveUploaded(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	if !samePath(cert, filepath.Join(dir, uploadDir, uploadCertName)) || !IsUploaded(cert, key) {
		t.Fatalf("unexpected paths %q %q", cert, key)
	}
	if st, _ := os.Stat(key); st.Mode().Perm() != 0o600 {
		t.Errorf("key mode = %v, want 0600", st.Mode().Perm())
	}

	info := Inspect(cert, key)
	if info.Source != SourceUploaded || info.Error != "" {
		t.Fatalf("info = %+v", info)
	}
	if len(info.DNSNames) != 1 || info.DNSNames[0] != "tv.example.com" || info.Trusted {
		t.Fatalf("info = %+v", info)
	}

	if err := RemoveUploaded(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cert); !os.IsNotExist(err) {
		t.Fatal("uploaded cert not removed")
	}
}

func TestSaveUploadedRejectsBadInput(t *testing.T) {
	withTempPath(t)
	src := t.TempDir()
	c1, k1 := filepath.Join(src, "c1"), filepath.Join(src, "k1")
	c2, k2 := filepath.Join(src, "c2"), filepath.Join(src, "k2")
	writeUserCert(t, c1, k1)
	writeUserCert(t, c2, k2)
	cert1, _ := os.ReadFile(c1)
	key2, _ := os.ReadFile(k2)

	for name, tc := range map[string]struct {
		cert, key []byte
		want      string
	}{
		"empty":     {nil, key2, "required"},
		"mismatch":  {cert1, key2, "valid pair"},
		"garbage":   {cert1, []byte("hello"), "not PEM"},
		"encrypted": {cert1, []byte("-----BEGIN RSA PRIVATE KEY-----\nProc-Type: 4,ENCRYPTED\nDEK-Info: AES-128-CBC,00\n\nAAAA\n-----END RSA PRIVATE KEY-----\n"), "password"},
	} {
		if _, _, err := SaveUploaded(tc.cert, tc.key); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}

func TestInspectSources(t *testing.T) {
	withTempPath(t)
	if got := Inspect("", "").Source; got != SourceNone {
		t.Errorf("empty: %s", got)
	}
	cert, key, err := MakeCertKeyFiles([]string{"192.168.1.10"})
	if err != nil {
		t.Fatal(err)
	}
	if info := Inspect(cert, key); info.Source != SourceSelfSigned || info.Error != "" || info.Trusted {
		t.Errorf("self-signed: %+v", info)
	}
	if info := Inspect("/nonexistent/c.pem", "/nonexistent/k.pem"); info.Source != SourceUser || info.Error == "" {
		t.Errorf("missing user cert: %+v", info)
	}
}
