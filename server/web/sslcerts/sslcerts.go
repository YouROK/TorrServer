package sslcerts

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"server/log"
	"server/settings"
)

const (
	certFileName = "server.pem"
	keyFileName  = "server.key"
)

// EnsureCert returns usable cert and key file paths for the HTTPS server.
//
// If no paths are configured, a self-signed pair is generated. If the configured
// pair is invalid, it is regenerated only when it is our own self-signed pair;
// a user-supplied cert is never replaced, the error is returned instead.
// changed reports whether the returned paths differ from the input and should be saved.
func EnsureCert(certFile, keyFile string, ips []string) (cert, key string, changed bool, err error) {
	if certFile == "" || keyFile == "" {
		cert, key, err = MakeCertKeyFiles(ips)
		return cert, key, err == nil, err
	}
	verr := VerifyCertKeyFiles(certFile, keyFile)
	if verr == nil {
		if IsGenerated(certFile, keyFile) {
			restrictKeyPerms(keyFile)
		}
		return certFile, keyFile, false, nil
	}
	if !IsGenerated(certFile, keyFile) {
		return "", "", false, fmt.Errorf("invalid ssl cert %q / key %q: %w", certFile, keyFile, verr)
	}
	log.TLogln("Self-signed certificate is invalid, regenerating:", verr)
	cert, key, err = MakeCertKeyFiles(ips)
	return cert, key, err == nil && (cert != certFile || key != keyFile), err
}

// IsGenerated reports whether the paths point to the self-signed pair managed by TorrServer.
func IsGenerated(certFile, keyFile string) bool {
	c, k := generatedPaths()
	return samePath(certFile, c) && samePath(keyFile, k)
}

func generatedPaths() (string, string) {
	return filepath.Join(settings.Path, certFileName), filepath.Join(settings.Path, keyFileName)
}

func samePath(a, b string) bool {
	aa, err1 := filepath.Abs(a)
	bb, err2 := filepath.Abs(b)
	return err1 == nil && err2 == nil && filepath.Clean(aa) == filepath.Clean(bb)
}

func generateSelfSignedCert(ips []string) ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now().Add(-time.Hour) // tolerate small clock skew on clients
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	netIps := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	for _, ip := range ips {
		if parsed := net.ParseIP(ip); parsed != nil {
			netIps = append(netIps, parsed)
		}
	}

	dnsNames := []string{"localhost"}
	if host, err := os.Hostname(); err == nil && host != "" {
		host = strings.TrimSuffix(host, ".local")
		dnsNames = append(dnsNames, host, host+".local")
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"TorrServer"},
			CommonName:   "TorrServer",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		IPAddresses:           netIps,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	return certPEM, privPEM, nil
}

// MakeCertKeyFiles generates a self-signed cert and key in settings.Path and returns their absolute paths.
func MakeCertKeyFiles(ips []string) (string, string, error) {
	certPEM, privPEM, err := generateSelfSignedCert(ips)
	if err != nil {
		return "", "", fmt.Errorf("generate certificate: %w", err)
	}
	certPath, keyPath := generatedPaths()
	if certPath, err = filepath.Abs(certPath); err != nil {
		return "", "", err
	}
	if keyPath, err = filepath.Abs(keyPath); err != nil {
		return "", "", err
	}
	if err = writeFileAtomic(keyPath, privPEM, 0o600); err != nil {
		return "", "", fmt.Errorf("write private key: %w", err)
	}
	if err = writeFileAtomic(certPath, certPEM, 0o644); err != nil {
		return "", "", fmt.Errorf("write certificate: %w", err)
	}
	log.TLogln("Self-signed certificate and private key generated successfully.")
	return certPath, keyPath, nil
}

// writeFileAtomic writes data to a temp file in the same directory and renames it into place,
// so a crash never leaves a truncated PEM behind.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after successful rename

	if err = tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// restrictKeyPerms tightens permissions of key files generated by older versions (0644).
func restrictKeyPerms(keyFile string) {
	st, err := os.Stat(keyFile)
	if err != nil || st.Mode().Perm()&0o077 == 0 {
		return
	}
	if err := os.Chmod(keyFile, 0o600); err != nil {
		log.TLogln("Error restricting private key permissions:", err)
	}
}

// VerifyCertKeyFiles checks that the cert and key load, match and the leaf is currently valid.
func VerifyCertKeyFiles(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	if len(cert.Certificate) == 0 {
		return errors.New("no certificate found")
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}
	now := time.Now()
	if now.Before(leaf.NotBefore) {
		return fmt.Errorf("certificate is not valid until %s", leaf.NotBefore.Format(time.RFC3339))
	}
	if now.After(leaf.NotAfter) {
		return fmt.Errorf("certificate has expired on %s", leaf.NotAfter.Format(time.RFC3339))
	}
	sans := append([]string{}, leaf.DNSNames...)
	for _, ip := range leaf.IPAddresses {
		sans = append(sans, ip.String())
	}
	log.TLogln("Certificate valid:", leaf.Subject.String(), "SANs:", sans, "expires:", leaf.NotAfter.Format(time.RFC3339))
	return nil
}
