package main

import (
	"os"
	"path/filepath"
	"runtime"
)

// Go reads trusted root CAs for outgoing HTTPS (posters, TMDB, Torznab, tracker lists)
// from fixed system locations. Entware, used on routers such as Keenetic, installs them
// under /opt/etc/ssl instead, so without SSL_CERT_FILE/SSL_CERT_DIR every outgoing HTTPS
// request fails certificate verification.

// systemCAFiles and systemCADirs mirror crypto/x509's Linux search paths.
var (
	systemCAFiles = []string{
		"/etc/ssl/certs/ca-certificates.crt",
		"/etc/pki/tls/certs/ca-bundle.crt",
		"/etc/ssl/ca-bundle.pem",
		"/etc/pki/tls/cacert.pem",
		"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem",
		"/etc/ssl/cert.pem",
	}
	systemCADirs = []string{"/etc/ssl/certs", "/etc/pki/tls/certs"}

	entwareCAFiles = []string{"/opt/etc/ssl/certs/ca-certificates.crt", "/opt/etc/ssl/cert.pem"}
	entwareCADirs  = []string{"/opt/etc/ssl/certs"}

	caRootsOS = runtime.GOOS
)

// configureCARoots points crypto/x509 at Entware's root CAs when the system locations have
// none and SSL_CERT_FILE/SSL_CERT_DIR aren't set. It must run before the first TLS
// connection, as Go loads the system roots once. It returns a message to log, or "".
func configureCARoots() string {
	if caRootsOS != "linux" || os.Getenv("SSL_CERT_FILE") != "" || os.Getenv("SSL_CERT_DIR") != "" {
		return ""
	}
	if hasCARoots(systemCAFiles, systemCADirs) {
		return ""
	}
	file, dir := firstFile(entwareCAFiles), firstNonEmptyDir(entwareCADirs)
	if file == "" && dir == "" {
		return "Warning: no trusted root CA certificates found, outgoing HTTPS requests will fail. " +
			"Install ca-certificates or set SSL_CERT_FILE/SSL_CERT_DIR"
	}
	msg := "Root CA certificates:"
	if file != "" {
		os.Setenv("SSL_CERT_FILE", file)
		msg += " SSL_CERT_FILE=" + file
	}
	if dir != "" {
		os.Setenv("SSL_CERT_DIR", dir)
		msg += " SSL_CERT_DIR=" + dir
	}
	return msg
}

func hasCARoots(files, dirs []string) bool {
	return firstFile(files) != "" || firstNonEmptyDir(dirs) != ""
}

func firstFile(paths []string) string {
	for _, p := range paths {
		if fi, err := os.Stat(p); err == nil && fi.Mode().IsRegular() && fi.Size() > 0 {
			return p
		}
	}
	return ""
}

func firstNonEmptyDir(paths []string) string {
	for _, p := range paths {
		if entries, err := os.ReadDir(p); err == nil {
			for _, e := range entries {
				if fi, err := os.Stat(filepath.Join(p, e.Name())); err == nil && fi.Mode().IsRegular() {
					return p
				}
			}
		}
	}
	return ""
}
