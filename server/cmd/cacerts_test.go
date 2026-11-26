package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// useCAPaths points the CA search at dir/sys and dir/opt for one test.
func useCAPaths(t *testing.T) (sys, opt string) {
	t.Helper()
	dir := t.TempDir()
	sys, opt = filepath.Join(dir, "sys"), filepath.Join(dir, "opt")
	oldFiles, oldDirs, oldEFiles, oldEDirs, oldOS := systemCAFiles, systemCADirs, entwareCAFiles, entwareCADirs, caRootsOS
	t.Cleanup(func() {
		systemCAFiles, systemCADirs, entwareCAFiles, entwareCADirs, caRootsOS = oldFiles, oldDirs, oldEFiles, oldEDirs, oldOS
	})
	systemCAFiles = []string{filepath.Join(sys, "ca-certificates.crt")}
	systemCADirs = []string{filepath.Join(sys, "certs")}
	entwareCAFiles = []string{filepath.Join(opt, "ca-certificates.crt")}
	entwareCADirs = []string{filepath.Join(opt, "certs")}
	caRootsOS = "linux"
	t.Setenv("SSL_CERT_FILE", "")
	t.Setenv("SSL_CERT_DIR", "")
	return sys, opt
}

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCARootsUsesEntwareWhenSystemHasNone(t *testing.T) {
	_, opt := useCAPaths(t)
	writeFile(t, filepath.Join(opt, "ca-certificates.crt"), "pem")
	writeFile(t, filepath.Join(opt, "certs", "a.pem"), "pem")

	msg := configureCARoots()
	if got := os.Getenv("SSL_CERT_FILE"); got != filepath.Join(opt, "ca-certificates.crt") {
		t.Errorf("SSL_CERT_FILE = %q", got)
	}
	if got := os.Getenv("SSL_CERT_DIR"); got != filepath.Join(opt, "certs") {
		t.Errorf("SSL_CERT_DIR = %q", got)
	}
	if !strings.Contains(msg, "SSL_CERT_FILE=") {
		t.Errorf("message = %q", msg)
	}
}

func TestCARootsKeepsSystemRoots(t *testing.T) {
	for name, setup := range map[string]func(sys string){
		"file": func(sys string) { writeFile(t, filepath.Join(sys, "ca-certificates.crt"), "pem") },
		"dir":  func(sys string) { writeFile(t, filepath.Join(sys, "certs", "a.pem"), "pem") },
	} {
		t.Run(name, func(t *testing.T) {
			sys, opt := useCAPaths(t)
			setup(sys)
			writeFile(t, filepath.Join(opt, "ca-certificates.crt"), "pem")
			if msg := configureCARoots(); msg != "" || os.Getenv("SSL_CERT_FILE") != "" {
				t.Errorf("system roots present: message %q, SSL_CERT_FILE %q", msg, os.Getenv("SSL_CERT_FILE"))
			}
		})
	}
}

func TestCARootsIgnoresEmptySystemLocations(t *testing.T) {
	sys, opt := useCAPaths(t)
	writeFile(t, filepath.Join(sys, "ca-certificates.crt"), "")
	if err := os.MkdirAll(filepath.Join(sys, "certs"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(opt, "ca-certificates.crt"), "pem")
	configureCARoots()
	if got := os.Getenv("SSL_CERT_FILE"); got != filepath.Join(opt, "ca-certificates.crt") {
		t.Errorf("SSL_CERT_FILE = %q, want the Entware file", got)
	}
}

func TestCARootsRespectsUserEnv(t *testing.T) {
	_, opt := useCAPaths(t)
	writeFile(t, filepath.Join(opt, "ca-certificates.crt"), "pem")
	t.Setenv("SSL_CERT_DIR", "/custom")
	if msg := configureCARoots(); msg != "" || os.Getenv("SSL_CERT_FILE") != "" {
		t.Errorf("user env set: message %q, SSL_CERT_FILE %q", msg, os.Getenv("SSL_CERT_FILE"))
	}
}

func TestCARootsWarnsWhenNoneFound(t *testing.T) {
	useCAPaths(t)
	if msg := configureCARoots(); !strings.HasPrefix(msg, "Warning:") {
		t.Errorf("message = %q, want a warning", msg)
	}
}

func TestCARootsNoopOutsideLinux(t *testing.T) {
	_, opt := useCAPaths(t)
	writeFile(t, filepath.Join(opt, "ca-certificates.crt"), "pem")
	caRootsOS = "android"
	if msg := configureCARoots(); msg != "" || os.Getenv("SSL_CERT_FILE") != "" {
		t.Errorf("android: message %q, SSL_CERT_FILE %q", msg, os.Getenv("SSL_CERT_FILE"))
	}
}
