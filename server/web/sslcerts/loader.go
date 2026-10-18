package sslcerts

import (
	"crypto/tls"
	"os"
	"sync"
	"time"

	"server/log"
)

// reloadCheckInterval limits how often the cert files are stat'ed during handshakes.
const reloadCheckInterval = 5 * time.Second

// Loader serves the current certificate to the TLS server and reloads it when the
// configured paths or the files' modification times change. On a failed reload the
// previous certificate keeps being served.
type Loader struct {
	paths func() (certFile, keyFile string)

	mu                sync.Mutex
	cert              *tls.Certificate
	certFile, keyFile string
	certMod, keyMod   time.Time
	lastCheck         time.Time
}

// NewLoader loads the initial certificate from paths, which is re-read on every reload check
// so path changes made in settings apply without a restart.
func NewLoader(paths func() (certFile, keyFile string)) (*Loader, error) {
	certFile, keyFile := paths()
	cert, err := loadPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &Loader{
		paths:     paths,
		cert:      cert,
		certFile:  certFile,
		keyFile:   keyFile,
		certMod:   modTime(certFile),
		keyMod:    modTime(keyFile),
		lastCheck: time.Now(),
	}, nil
}

// GetCertificate implements tls.Config.GetCertificate.
func (l *Loader) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if time.Since(l.lastCheck) >= reloadCheckInterval {
		l.lastCheck = time.Now()
		l.reloadLocked()
	}
	return l.cert, nil
}

func (l *Loader) reloadLocked() {
	certFile, keyFile := l.paths()
	if certFile == "" || keyFile == "" {
		return
	}
	certMod, keyMod := modTime(certFile), modTime(keyFile)
	if certFile == l.certFile && keyFile == l.keyFile && certMod.Equal(l.certMod) && keyMod.Equal(l.keyMod) {
		return
	}
	// remember the attempt so a broken file is reported once, not on every handshake
	l.certFile, l.keyFile, l.certMod, l.keyMod = certFile, keyFile, certMod, keyMod

	cert, err := loadPair(certFile, keyFile)
	if err != nil {
		log.TLogln("TLS certificate reload failed, keeping previous certificate:", err)
		return
	}
	l.cert = cert
	log.TLogln("TLS certificate reloaded:", certFile, "expires:", cert.Leaf.NotAfter.Format(time.RFC3339))
}

func modTime(path string) time.Time {
	st, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return st.ModTime()
}

// RenewLoop regenerates the self-signed certificate every interval when it is close to
// expiry or no longer covers the current local IPs, until stop is closed.
// User-supplied certificates are left alone; the Loader picks up their renewals.
func RenewLoop(stop <-chan struct{}, interval time.Duration, paths func() (certFile, keyFile string), ips func() []string) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
		}
		certFile, keyFile := paths()
		if certFile == "" || keyFile == "" || !IsGenerated(certFile, keyFile) {
			continue
		}
		if _, _, _, err := EnsureCert(certFile, keyFile, ips()); err != nil {
			log.TLogln("Self-signed certificate renewal failed:", err)
		}
	}
}
