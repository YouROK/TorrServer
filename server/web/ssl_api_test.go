package web

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"server/settings"
	"server/web/sslcerts"
)

// sslTestRouter opens a settings DB in a temp dir once per package (it can't be
// reopened after CloseDB) and starts each test without a configured certificate.
func sslTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	sslTestDBOnce.Do(func() {
		dir, err := os.MkdirTemp("", "ts-ssl-api")
		if err == nil {
			sslTestDBDir, settings.Path = dir, dir
			err = settings.InitSets(false, false)
		}
		sslTestDBErr = err
	})
	if sslTestDBErr != nil {
		t.Fatal(sslTestDBErr)
	}
	setSSLCertPaths("", "")
	settings.Ssl = true
	t.Cleanup(func() { settings.ReadOnly, settings.Ssl = false, false })
	gin.SetMode(gin.TestMode)
	r := gin.New()
	setupSSLRoutes(r)
	return r
}

var (
	sslTestDBOnce sync.Once
	sslTestDBErr  error
	sslTestDBDir  string
)

func TestMain(m *testing.M) {
	code := m.Run()
	if sslTestDBDir != "" {
		settings.CloseDB()
		os.RemoveAll(sslTestDBDir)
	}
	os.Exit(code)
}

func sslDo(t *testing.T, r http.Handler, method, path string, body *bytes.Buffer, contentType string) (int, sslStatus, string) {
	t.Helper()
	if body == nil {
		body = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var st sslStatus
	_ = json.Unmarshal(w.Body.Bytes(), &st)
	return w.Code, st, w.Body.String()
}

func pemPair(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(7),
		Subject:      pkix.Name{CommonName: "tv.example.com"},
		DNSNames:     []string{"tv.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, _ := x509.MarshalECPrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
}

func TestSSLAPIUploadAndSwitchBack(t *testing.T) {
	r := sslTestRouter(t)

	code, st, _ := sslDo(t, r, http.MethodGet, "/ssl/status", nil, "")
	if code != http.StatusOK || st.Cert.Source != sslcerts.SourceNone {
		t.Fatalf("initial status %d %+v", code, st)
	}
	if code, st, _ = sslDo(t, r, http.MethodPost, "/ssl/selfsigned", nil, ""); code != http.StatusOK || st.Cert.Source != sslcerts.SourceSelfSigned {
		t.Fatalf("first self-signed %d %+v", code, st)
	}

	certPEM, keyPEM := pemPair(t)
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, _ := mw.CreateFormFile("cert", "cert.pem")
	fw.Write(certPEM)
	fw, _ = mw.CreateFormFile("key", "key.pem")
	fw.Write(keyPEM)
	mw.Close()
	code, st, raw := sslDo(t, r, http.MethodPost, "/ssl/upload", body, mw.FormDataContentType())
	if code != http.StatusOK || st.Cert.Source != sslcerts.SourceUploaded || st.Cert.Subject != "CN=tv.example.com" {
		t.Fatalf("upload %d %s", code, raw)
	}
	if !sslcerts.IsUploaded(settings.BTsets.SslCert, settings.BTsets.SslKey) {
		t.Fatal("uploaded paths not saved in settings")
	}

	// regenerating is only for the self-signed certificate
	if code, _, _ := sslDo(t, r, http.MethodPost, "/ssl/regenerate", nil, ""); code != http.StatusConflict {
		t.Fatalf("regenerate with uploaded cert: %d", code)
	}

	code, _, raw = sslDo(t, r, http.MethodGet, "/ssl/cert", nil, "")
	if code != http.StatusOK || !strings.Contains(raw, "BEGIN CERTIFICATE") || strings.Contains(raw, "PRIVATE KEY") {
		t.Fatalf("download %d %q", code, raw)
	}

	uploadedCert := settings.BTsets.SslCert
	selfCert, _ := sslcerts.SelfSignedPaths()
	selfBefore, _ := os.ReadFile(selfCert)
	code, st, raw = sslDo(t, r, http.MethodPost, "/ssl/selfsigned", nil, "")
	if code != http.StatusOK || st.Cert.Source != sslcerts.SourceSelfSigned {
		t.Fatalf("selfsigned %d %s", code, raw)
	}
	if info := sslcerts.Inspect(uploadedCert, uploadedCert); info.Error == "" {
		t.Fatal("uploaded certificate not removed")
	}
	if selfAfter, _ := os.ReadFile(selfCert); len(selfBefore) == 0 || !bytes.Equal(selfBefore, selfAfter) {
		t.Fatal("existing self-signed certificate was replaced instead of reused")
	}

	before, _ := os.ReadFile(st.Cert.CertFile)
	code, st, raw = sslDo(t, r, http.MethodPost, "/ssl/regenerate", nil, "")
	after, _ := os.ReadFile(st.Cert.CertFile)
	if code != http.StatusOK || st.Cert.Source != sslcerts.SourceSelfSigned || bytes.Equal(before, after) {
		t.Fatalf("regenerate %d %s", code, raw)
	}
}

func TestSSLAPIRejectsBadUploadAndReadOnly(t *testing.T) {
	r := sslTestRouter(t)
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, _ := mw.CreateFormFile("cert", "cert.pem")
	fw.Write([]byte("not a cert"))
	mw.Close()
	if code, _, _ := sslDo(t, r, http.MethodPost, "/ssl/upload", body, mw.FormDataContentType()); code != http.StatusBadRequest {
		t.Fatalf("missing key: %d", code)
	}

	settings.ReadOnly = true
	if code, _, _ := sslDo(t, r, http.MethodPost, "/ssl/selfsigned", nil, ""); code != http.StatusForbidden {
		t.Fatalf("read-only: %d", code)
	}
	settings.ReadOnly = false

	oldArgs := settings.Args
	t.Cleanup(func() { settings.Args = oldArgs })
	settings.Args = &settings.ExecArgs{SslCert: "/etc/c.pem", SslKey: "/etc/k.pem"}
	code, st, _ := sslDo(t, r, http.MethodPost, "/ssl/selfsigned", nil, "")
	if code != http.StatusConflict {
		t.Fatalf("cert from flags: %d", code)
	}
	if _, st, _ = sslDo(t, r, http.MethodGet, "/ssl/status", nil, ""); !st.CertFromFlags {
		t.Fatal("cert_from_flags not reported")
	}
}

func TestSSLAPIPaths(t *testing.T) {
	r := sslTestRouter(t)
	dir := t.TempDir()
	certPEM, keyPEM := pemPair(t)
	certFile, keyFile := filepath.Join(dir, "fullchain.pem"), filepath.Join(dir, "privkey.pem")
	os.WriteFile(certFile, certPEM, 0o644)
	os.WriteFile(keyFile, keyPEM, 0o600)

	// start from an uploaded certificate: switching to paths deletes its copy
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, _ := mw.CreateFormFile("cert", "cert.pem")
	fw.Write(certPEM)
	fw, _ = mw.CreateFormFile("key", "key.pem")
	fw.Write(keyPEM)
	mw.Close()
	if code, _, raw := sslDo(t, r, http.MethodPost, "/ssl/upload", body, mw.FormDataContentType()); code != http.StatusOK {
		t.Fatalf("upload %d %s", code, raw)
	}
	uploaded := settings.BTsets.SslCert

	req := func(cert, key string) *bytes.Buffer {
		b, _ := json.Marshal(map[string]string{"cert": cert, "key": key})
		return bytes.NewBuffer(b)
	}
	if code, _, _ := sslDo(t, r, http.MethodPost, "/ssl/paths", req(certFile, filepath.Join(dir, "missing.key")), "application/json"); code != http.StatusBadRequest {
		t.Fatalf("missing key: %d", code)
	}
	code, st, raw := sslDo(t, r, http.MethodPost, "/ssl/paths", req(certFile, keyFile), "application/json")
	if code != http.StatusOK || st.Cert.Source != sslcerts.SourceUser || st.Cert.CertFile != certFile {
		t.Fatalf("paths %d %s", code, raw)
	}
	if _, err := os.Stat(uploaded); !os.IsNotExist(err) {
		t.Fatal("uploaded copy not removed after switching to paths")
	}
}

func TestSSLAPIRequiresHTTPS(t *testing.T) {
	r := sslTestRouter(t)
	settings.Ssl = false
	code, st, _ := sslDo(t, r, http.MethodGet, "/ssl/status", nil, "")
	if code != http.StatusOK || st.Enabled {
		t.Fatalf("status %d %+v", code, st)
	}
	for _, path := range []string{"/ssl/selfsigned", "/ssl/regenerate", "/ssl/upload", "/ssl/paths"} {
		if code, _, _ := sslDo(t, r, http.MethodPost, path, nil, ""); code != http.StatusConflict {
			t.Errorf("%s without --ssl: %d", path, code)
		}
	}
	if code, _, _ := sslDo(t, r, http.MethodGet, "/ssl/cert", nil, ""); code != http.StatusNotFound {
		t.Errorf("download without --ssl: %d", code)
	}
}
