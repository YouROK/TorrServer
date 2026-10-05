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
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"server/log"
	"server/settings"
)

const (
	certFileName = "server.pem"
	keyFileName  = "server.key"
)

// renewBefore is how long before expiry the self-signed cert is regenerated.
const renewBefore = 30 * 24 * time.Hour

// selfSignedValidity is the lifetime of generated certs (a var so tests can shorten it).
var selfSignedValidity = 365 * 24 * time.Hour

// EnsureCert returns usable cert and key file paths for the HTTPS server.
//
// If no paths are configured, a self-signed pair is generated. The self-signed pair
// is regenerated when it is invalid, close to expiry, or missing a current local IP
// or hostname (previously covered IPs are kept). A user-supplied cert, including one
// copied to the default location, is never replaced; the error is returned instead.
// changed reports whether the returned paths differ from the input and should be saved.
func EnsureCert(certFile, keyFile string, ips []string) (cert, key string, changed bool, err error) {
	if certFile == "" || keyFile == "" {
		c, k := generatedPaths()
		if !ownedPair(c, k) {
			// the user placed a certificate at the default location: use it, never overwrite it
			if verr := VerifyCertKeyFiles(c, k); verr != nil {
				return "", "", false, fmt.Errorf("certificate at %q / %q is not TorrServer's and is invalid, "+
					"so it is left untouched (fix it, or delete both files to have a new one generated): %w", c, k, verr)
			}
			log.TLogln("Using existing certificate at the default location:", c)
			return c, k, true, nil
		}
		cert, key, err = MakeCertKeyFiles(ips)
		return cert, key, err == nil, err
	}
	generated := IsGenerated(certFile, keyFile)
	pair, verr := loadPair(certFile, keyFile)
	switch {
	case verr != nil && !generated:
		err = fmt.Errorf("invalid ssl cert %q / key %q: %w", certFile, keyFile, verr)
		if c, k := generatedPaths(); samePath(certFile, c) && samePath(keyFile, k) {
			err = fmt.Errorf("%w (not a TorrServer self-signed certificate, so it is left untouched; "+
				"delete both files to have a new one generated)", err)
		}
		return "", "", false, err
	case verr != nil:
		log.TLogln("Self-signed certificate is invalid, regenerating:", verr)
	case !generated:
		return certFile, keyFile, false, nil
	default:
		restrictKeyPerms(keyFile)
		reason := renewalReason(pair.Leaf, ips)
		if reason == "" {
			return certFile, keyFile, false, nil
		}
		log.TLogln("Renewing self-signed certificate:", reason)
		// keep previously covered IPs so flapping interfaces (VPN, Docker) don't cause churn
		ips = mergeIPs(pair.Leaf.IPAddresses, ips)
	}
	cert, key, err = MakeCertKeyFiles(ips)
	return cert, key, err == nil && (cert != certFile || key != keyFile), err
}

// renewalReason returns why a generated leaf should be regenerated, or "" if it is fine.
func renewalReason(leaf *x509.Certificate, ips []string) string {
	if time.Until(leaf.NotAfter) < renewBefore {
		return "expires on " + leaf.NotAfter.Format(time.RFC3339)
	}
	for _, s := range ips {
		if ip := net.ParseIP(s); ip != nil && !containsIP(leaf.IPAddresses, ip) {
			return "missing IP " + s
		}
	}
	for _, name := range localDNSNames() {
		if !slices.Contains(leaf.DNSNames, name) {
			return "missing hostname " + name
		}
	}
	return ""
}

func containsIP(list []net.IP, ip net.IP) bool {
	return slices.ContainsFunc(list, ip.Equal)
}

func mergeIPs(old []net.IP, current []string) []string {
	out := slices.Clone(current)
	for _, ip := range old {
		out = append(out, ip.String())
	}
	return out
}

func localDNSNames() []string {
	names := []string{"localhost"}
	if host, err := os.Hostname(); err == nil && host != "" {
		host = strings.TrimSuffix(host, ".local")
		names = append(names, host, host+".local")
	}
	return names
}

// IsGenerated reports whether the paths point to the self-signed pair managed by TorrServer:
// the default location, holding a certificate TorrServer generated (or none yet).
// Ownership is decided by content, not by name, so a user's own certificate copied to
// the default location is never regenerated or treated as self-signed.
func IsGenerated(certFile, keyFile string) bool {
	c, k := generatedPaths()
	return samePath(certFile, c) && samePath(keyFile, k) && ownedPair(certFile, keyFile)
}

// ownedPair reports whether TorrServer may (re)write the pair: the cert is a TorrServer
// self-signed one, or neither file exists yet. A lone key without a cert is kept, as it
// may be the user's. Generated pairs are written cert first, so an interrupted write
// never leaves such a lone key behind.
func ownedPair(certFile, keyFile string) bool {
	if _, err := os.Stat(certFile); errors.Is(err, fs.ErrNotExist) {
		_, err := os.Stat(keyFile)
		return errors.Is(err, fs.ErrNotExist)
	}
	return ownedCert(certFile)
}

type ownedEntry struct {
	mod   time.Time
	size  int64
	owned bool
}

var (
	ownedMu    sync.Mutex
	ownedCache = map[string]ownedEntry{}
)

// ownedCert reports whether certFile is missing or holds a TorrServer self-signed
// certificate. Results are cached by modification time and size, as it runs per request.
func ownedCert(certFile string) bool {
	st, err := os.Stat(certFile)
	if errors.Is(err, fs.ErrNotExist) {
		return true
	}
	if err != nil {
		return false
	}
	ownedMu.Lock()
	defer ownedMu.Unlock()
	if e, ok := ownedCache[certFile]; ok && e.mod.Equal(st.ModTime()) && e.size == st.Size() {
		return e.owned
	}
	owned := isTorrServerSelfSigned(readLeaf(certFile))
	ownedCache[certFile] = ownedEntry{mod: st.ModTime(), size: st.Size(), owned: owned}
	return owned
}

func readLeaf(certFile string) *x509.Certificate {
	data, err := os.ReadFile(certFile)
	if err != nil {
		return nil
	}
	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			return nil
		}
		if block.Type == "CERTIFICATE" {
			leaf, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil
			}
			return leaf
		}
	}
}

// isTorrServerSelfSigned matches certificates made by generateSelfSignedCert, including
// those from older versions: issuer == subject, organization TorrServer, signed by its own key.
func isTorrServerSelfSigned(leaf *x509.Certificate) bool {
	return leaf != nil &&
		bytes.Equal(leaf.RawIssuer, leaf.RawSubject) &&
		slices.Equal(leaf.Subject.Organization, []string{"TorrServer"}) &&
		leaf.CheckSignature(leaf.SignatureAlgorithm, leaf.RawTBSCertificate, leaf.Signature) == nil
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
	notAfter := notBefore.Add(selfSignedValidity)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	netIps := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	for _, ip := range ips {
		if parsed := net.ParseIP(ip); parsed != nil && !containsIP(netIps, parsed) {
			netIps = append(netIps, parsed)
		}
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
		DNSNames:              localDNSNames(),
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
	if !ownedPair(certPath, keyPath) {
		return "", "", fmt.Errorf("refusing to overwrite %q / %q: not a TorrServer self-signed certificate", certPath, keyPath)
	}
	// cert first: a crash in between leaves our cert without a key, which is regenerated
	// next time, instead of a lone key that looks like the user's
	if err = writeFileAtomic(certPath, certPEM, 0o644); err != nil {
		return "", "", fmt.Errorf("write certificate: %w", err)
	}
	if err = writeFileAtomic(keyPath, privPEM, 0o600); err != nil {
		return "", "", fmt.Errorf("write private key: %w", err)
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

// loadPair loads the cert and key, checks they match and the leaf is currently valid.
func loadPair(certFile, keyFile string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	if len(cert.Certificate) == 0 {
		return nil, errors.New("no certificate found")
	}
	if cert.Leaf == nil {
		if cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
			return nil, err
		}
	}
	now := time.Now()
	if now.Before(cert.Leaf.NotBefore) {
		return nil, fmt.Errorf("certificate is not valid until %s", cert.Leaf.NotBefore.Format(time.RFC3339))
	}
	if now.After(cert.Leaf.NotAfter) {
		return nil, fmt.Errorf("certificate has expired on %s", cert.Leaf.NotAfter.Format(time.RFC3339))
	}
	return &cert, nil
}

// VerifyCertKeyFiles checks that the cert and key load, match and the leaf is currently valid,
// and logs what will be served.
func VerifyCertKeyFiles(certFile, keyFile string) error {
	cert, err := loadPair(certFile, keyFile)
	if err != nil {
		return err
	}
	leaf := cert.Leaf
	sans := slices.Clone(leaf.DNSNames)
	for _, ip := range leaf.IPAddresses {
		sans = append(sans, ip.String())
	}
	log.TLogln("Certificate valid:", leaf.Subject.String(), "SANs:", sans, "expires:", leaf.NotAfter.Format(time.RFC3339))
	return nil
}
