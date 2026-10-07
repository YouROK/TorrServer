package web

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"

	"server/log"
	"server/settings"
	"server/web/auth"
	"server/web/sslcerts"
)

// sslMu serialises certificate changes made through the API.
var sslMu sync.Mutex

type sslStatus struct {
	// Enabled is true when TorrServer was started with --ssl. Certificate changes are
	// stored either way and served without a restart while HTTPS runs.
	Enabled     bool   `json:"enabled"`
	Port        string `json:"port,omitempty"`
	HTTPPort    string `json:"http_port,omitempty"`
	HTTPEnabled bool   `json:"http_enabled"`
	ForceHTTPS  bool   `json:"force_https"`
	HTTPMedia   bool   `json:"http_media"`
	ReadOnly    bool   `json:"read_only"`
	// CertFromFlags is true when --sslcert/--sslkey set the paths: they are applied again
	// on every start, so the certificate can't be changed here.
	CertFromFlags bool          `json:"cert_from_flags"`
	Cert          sslcerts.Info `json:"cert"`
}

func setupSSLRoutes(route gin.IRouter) {
	g := route.Group("/ssl", auth.CheckAuth())
	g.GET("/status", sslStatusHandler)
	g.GET("/cert", sslCertDownload)
	g.POST("/upload", sslUpload)
	g.POST("/selfsigned", sslUseSelfSigned)
	g.POST("/regenerate", sslRegenerate)
}

func currentSSLStatus() sslStatus {
	st := sslStatus{
		Enabled:       settings.Ssl,
		HTTPPort:      settings.Port,
		HTTPEnabled:   settings.HTTPEnabled(),
		ReadOnly:      settings.ReadOnly,
		CertFromFlags: certFromFlags(),
		Cert:          sslcerts.Inspect(sslCertPaths()),
	}
	if settings.Ssl {
		st.Port = settings.SslPort
		if settings.Args != nil {
			st.ForceHTTPS = settings.Args.ForceHTTPS
			st.HTTPMedia = settings.Args.ForceHTTPS && settings.Args.HTTPMedia
		}
	}
	return st
}

// sslStatusHandler godoc
//
//	@Summary		HTTPS status
//	@Description	HTTPS mode, ports and the configured certificate (subject, SANs, issuer, validity, source). Never returns key material.
//
//	@Tags			API
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	sslStatus
//	@Router			/ssl/status [get]
func sslStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, currentSSLStatus())
}

// sslCertDownload godoc
//
//	@Summary		Download the HTTPS certificate
//	@Description	The configured certificate (chain) in PEM, e.g. to trust the self-signed one on a device. The private key is never served.
//
//	@Tags			API
//	@Produce		application/x-x509-ca-cert
//	@Security		BasicAuth
//	@Success		200	{file}		file
//	@Failure		404	{object}	map[string]string
//	@Router			/ssl/cert [get]
func sslCertDownload(c *gin.Context) {
	certFile, keyFile := sslCertPaths()
	if certFile == "" || keyFile == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no certificate configured"})
		return
	}
	name := filepath.Base(certFile)
	if sslcerts.IsGenerated(certFile, keyFile) {
		name = "torrserver.crt"
	}
	c.Header("Content-Type", "application/x-x509-ca-cert")
	c.FileAttachment(certFile, name)
}

// sslUpload godoc
//
//	@Summary		Upload an HTTPS certificate
//	@Description	Stores a PEM certificate (chain) and its unencrypted private key in <config>/ssl/ and uses them. The pair must match and be currently valid. Served without a restart when HTTPS is running.
//
//	@Tags			API
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BasicAuth
//	@Param			cert	formData	file	true	"Certificate (chain), PEM"
//	@Param			key		formData	file	true	"Private key, PEM"
//	@Success		200		{object}	sslStatus
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Router			/ssl/upload [post]
func sslUpload(c *gin.Context) {
	if denyCertChange(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2*sslcerts.MaxPEMSize+64<<10)
	certPEM, err := formFile(c, "cert")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	keyPEM, err := formFile(c, "key")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sslMu.Lock()
	defer sslMu.Unlock()
	cert, key, err := sslcerts.SaveUploaded(certPEM, keyPEM)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.TLogln("Uploaded HTTPS certificate saved:", cert)
	setSSLCertPaths(cert, key)
	c.JSON(http.StatusOK, currentSSLStatus())
}

func formFile(c *gin.Context, field string) ([]byte, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		return nil, errors.New(field + ": file is required")
	}
	if fh.Size > sslcerts.MaxPEMSize {
		return nil, errors.New(field + ": file is too large")
	}
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, sslcerts.MaxPEMSize))
}

// sslUseSelfSigned godoc
//
//	@Summary		Use the self-signed HTTPS certificate
//	@Description	Switches to TorrServer's self-signed certificate (reused, or generated if missing) and deletes an uploaded one. Certificate files given by path are left on disk.
//
//	@Tags			API
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	sslStatus
//	@Failure		403	{object}	map[string]string
//	@Failure		409	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/ssl/selfsigned [post]
func sslUseSelfSigned(c *gin.Context) {
	if denyCertChange(c) {
		return
	}
	sslMu.Lock()
	defer sslMu.Unlock()
	// reuse the existing self-signed pair: devices may already trust it
	c0, k0 := sslcerts.SelfSignedPaths()
	cert, key, _, err := sslcerts.EnsureCert(c0, k0, certIPs())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	wasUploaded := sslcerts.IsUploaded(sslCertPaths())
	setSSLCertPaths(cert, key)
	if wasUploaded {
		if err := sslcerts.RemoveUploaded(); err != nil {
			log.TLogln("Error removing uploaded certificate:", err)
		}
	}
	c.JSON(http.StatusOK, currentSSLStatus())
}

// sslRegenerate godoc
//
//	@Summary		Regenerate the self-signed HTTPS certificate
//	@Description	Creates a new self-signed certificate and key for the current local IPs and hostname. Only when the self-signed certificate is in use.
//
//	@Tags			API
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	sslStatus
//	@Failure		403	{object}	map[string]string
//	@Failure		409	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/ssl/regenerate [post]
func sslRegenerate(c *gin.Context) {
	if denyCertChange(c) {
		return
	}
	sslMu.Lock()
	defer sslMu.Unlock()
	if !sslcerts.IsGenerated(sslCertPaths()) {
		c.JSON(http.StatusConflict, gin.H{"error": "not using the self-signed certificate"})
		return
	}
	cert, key, err := sslcerts.MakeCertKeyFiles(certIPs())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	setSSLCertPaths(cert, key)
	c.JSON(http.StatusOK, currentSSLStatus())
}

func certFromFlags() bool {
	return settings.Args != nil && settings.Args.SslCert != "" && settings.Args.SslKey != ""
}

// denyCertChange rejects certificate changes that can't be saved or would be undone on
// the next start.
func denyCertChange(c *gin.Context) bool {
	switch {
	case settings.ReadOnly:
		c.JSON(http.StatusForbidden, gin.H{"error": "Read-only mode"})
	case certFromFlags():
		c.JSON(http.StatusConflict, gin.H{"error": "the certificate is set by --sslcert/--sslkey"})
	default:
		return false
	}
	return true
}

// setSSLCertPaths saves new cert paths; the Loader serves them within a few seconds.
func setSSLCertPaths(cert, key string) {
	cur, curKey := sslCertPaths()
	if cur == cert && curKey == key {
		return
	}
	sets := *settings.BTsets
	sets.SslCert, sets.SslKey = cert, key
	settings.SetBTSets(&sets)
}
