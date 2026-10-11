package web

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sort"
	"time"

	gstreamer "server/gstreamer/bridge"
	"server/netbind"

	"server/torrfs/fuse"
	"server/torrfs/webdav"

	"server/rutor"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location/v2"
	"github.com/gin-gonic/gin"
	"github.com/wlynxg/anet"

	"server/bonjour"
	"server/dlna"
	"server/settings"
	"server/web/msx"

	"server/log"
	"server/mcp"
	"server/torr"
	"server/version"
	"server/web/api"
	"server/web/auth"
	"server/web/pages"
	"server/web/sslcerts"
	"server/web/waf"
)

var (
	BTS      = torr.NewBTS()
	waitChan = make(chan error)
	// stopRenew stops the self-signed cert renewal loop.
	stopRenew chan struct{}
)

//	@title			Swagger Torrserver API
//	@version		{version.Version}
//	@description	Torrent streaming server.

//	@license.name	GPL 3.0

//	@BasePath	/

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func Start() error {
	log.TLogln("Start TorrServer " + version.Version + " torrent " + version.GetTorrentVersion())
	ips := GetLocalIps()
	if len(ips) > 0 {
		log.TLogln("Local IPs:", ips)
	}
	if settings.Ssl {
		cert, key, changed, err := sslcerts.EnsureCert(settings.BTsets.SslCert, settings.BTsets.SslKey, certIPs())
		if err != nil {
			log.TLogln("Cannot start HTTPS (fix --sslcert/--sslkey or clear them in settings to use a self-signed cert):", err)
			return err
		}
		if changed {
			settings.BTsets.SslCert, settings.BTsets.SslKey = cert, key
			log.TLogln("Saving path to ssl cert and key in db", cert, key)
			settings.SetBTSets(settings.BTsets)
		}
	}
	err := BTS.Connect()
	if err != nil {
		log.TLogln("BTS.Connect() error!", err)
		return err
	}
	rutor.Start()

	gin.SetMode(gin.ReleaseMode)

	// corsCfg := cors.DefaultConfig()
	// corsCfg.AllowAllOrigins = true
	// corsCfg.AllowHeaders = []string{"*"}
	// corsCfg.AllowMethods = []string{"*"}
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowAllOrigins = true
	corsCfg.AllowPrivateNetwork = true
	corsCfg.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS", "DELETE"}
	corsCfg.AllowHeaders = []string{
		"Origin", "Content-Length", "Content-Type", "X-Requested-With", "Accept", "Authorization",
		"Mcp-Protocol-Version", "Mcp-Session-Id", "Last-Event-ID", "Mcp-Method", "Mcp-Name",
	}

	route := gin.New()
	route.Use(log.WebLogger(), waf.WAF(), gin.Recovery(), cors.New(corsCfg), location.Default())
	auth.SetupAuth(route)

	route.GET("/echo", echo)

	api.SetupRoute(route)
	setupSSLRoutes(route)
	mcp.Mount(route.Group("/", auth.CheckAuth()))
	gstreamer.SetupRoute(route)
	msx.SetupRoute(route)
	pages.SetupRoute(route)
	if settings.Args.WebDAV {
		webdav.MountWebDAV(route)
	}

	if settings.BTsets.EnableDLNA {
		dlna.Start()
	}
	if settings.BTsets.EnableBonjour {
		bonjour.Start()
	}

	// Auto-mount FUSE filesystem if enabled
	fuse.FuseAutoMount()

	route.GET("/swagger/*any", swaggerHandler())

	if err := startServers(route); err != nil {
		shutdownServers()
		return err
	}
	return nil
}

// startServers binds the HTTP and (with --ssl) HTTPS ports and starts serving h.
// Plain HTTP sent to the HTTPS port is redirected to https://.
func startServers(h http.Handler) error {
	if settings.Ssl {
		loader, err := sslcerts.NewLoader(sslCertPaths)
		if err != nil {
			return fmt.Errorf("load ssl cert: %w", err)
		}
		ln, err := netbind.Listen(settings.IPs, settings.SslPort)
		if err != nil {
			return fmt.Errorf("https listen: %w", err)
		}
		tlsLn, plainLn := splitTLS(ln)

		srv := newServer(h)
		// MinVersion is left to Go's default (TLS 1.2) so GODEBUG=tls10server=1 still works for old TVs
		srv.TLSConfig = &tls.Config{GetCertificate: loader.GetCertificate}
		serve(func() error { return srv.ServeTLS(tlsLn, "", "") })

		redirect := newServer(httpsRedirectHandler())
		serve(func() error { return redirect.Serve(plainLn) })
		logAddrs("Start https server at", settings.SslPort)

		stopRenew = make(chan struct{})
		go sslcerts.RenewLoop(stopRenew, time.Hour, sslCertPaths, certIPs)
	}

	startInternalServer(h)

	if !settings.HTTPEnabled() {
		log.TLogln("HTTPS only: the plain HTTP port", settings.Port, "is not opened")
		warnSelfSignedStrict("--https-only")
		return nil
	}

	ln, err := netbind.Listen(settings.IPs, settings.Port)
	if err != nil {
		return fmt.Errorf("http listen: %w", err)
	}
	if settings.Args != nil && settings.Args.ForceHTTPS && settings.Ssl {
		httpMedia := settings.Args.HTTPMedia
		srv := newServer(forceHTTPSHandler(h, httpMedia))
		serve(func() error { return srv.Serve(ln) })
		if httpMedia {
			logAddrs("Start http server (media only, everything else redirects to https) at", settings.Port)
			log.TLogln("Warning: --http-media serves stream URLs over plain HTTP; they and Basic auth " +
				"credentials travel unencrypted. Don't expose the HTTP port to the internet.")
		} else {
			logAddrs("Start http server (redirect to https) at", settings.Port)
			warnSelfSignedStrict("--force-https")
		}
		return nil
	}
	srv := newServer(h)
	serve(func() error { return srv.Serve(ln) })
	logAddrs("Start http server at", settings.Port)
	return nil
}

// warnSelfSignedStrict warns when media is only served over HTTPS with the self-signed
// certificate, which most players and TVs reject.
func warnSelfSignedStrict(flag string) {
	if sslcerts.IsGenerated(sslCertPaths()) {
		log.TLogln("Warning: " + flag + " with a self-signed certificate: media players and TVs " +
			"usually reject it and won't play. Use a trusted certificate (see README, HTTPS)" +
			" or, on a trusted network, --force-https --http-media.")
	}
}

// startInternalServer serves h over plain HTTP on a random loopback port for
// TorrServer's requests to itself (ffprobe, GStreamer). Unlike the public HTTP port it is
// never redirected by --force-https, and it works when --ip excludes loopback. Proxies
// and port mappings can't target it, as the port changes on every start.
func startInternalServer(h http.Handler) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.TLogln("Internal loopback listener unavailable, self-requests use the public ports:", err)
		return
	}
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	srv := newServer(h)
	serve(func() error { return srv.Serve(ln) })
	settings.InternalPort = port
	log.TLogln("Internal loopback listener at", ln.Addr(), "(TorrServer's own requests only)")
}

// certIPs are the addresses the self-signed cert must cover: the bound IPs when
// --ip is given, all local IPs otherwise.
func certIPs() []string {
	var bound []string
	for _, ip := range settings.IPs {
		if parsed := net.ParseIP(ip); parsed != nil && !parsed.IsUnspecified() {
			bound = append(bound, ip)
		} else {
			return GetLocalIps() // "", 0.0.0.0 or :: bind everything
		}
	}
	if len(bound) == 0 {
		return GetLocalIps()
	}
	return bound
}

func sslCertPaths() (string, string) {
	return settings.BTsets.SslCert, settings.BTsets.SslKey
}

func logAddrs(msg, port string) {
	for _, ip := range netbind.Normalize(settings.IPs) {
		log.TLogln(msg, netbind.Addr(ip, port))
	}
}

func Wait() error {
	return <-waitChan
}

func Stop() {
	// stop accepting requests before tearing down what they depend on
	if stopRenew != nil {
		close(stopRenew)
		stopRenew = nil
	}
	shutdownServers()
	settings.InternalPort = ""
	gstreamer.Stop()
	dlna.Stop()
	bonjour.Stop()
	// Unmount FUSE filesystem if mounted
	fuse.FuseCleanup()
	BTS.Disconnect()
	waitChan <- nil
}

// echo godoc
//
//	@Summary		Tests server status
//	@Description	Tests whether server is alive or not
//
//	@Tags			API
//
//	@Produce		plain
//	@Success		200	{string}	string	"Server version"
//	@Router			/echo [get]
func echo(c *gin.Context) {
	c.String(200, "%v", version.Version)
}

func GetLocalIps() []string {
	ifaces, err := anet.Interfaces()
	if err != nil {
		log.TLogln("Error get local IPs")
		return nil
	}
	var list []string
	for _, i := range ifaces {
		addrs, _ := anet.InterfaceAddrsByInterface(&i)
		if i.Flags&net.FlagUp == net.FlagUp {
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() {
					list = append(list, ip.String())
				}
			}
		}
	}
	sort.Strings(list)
	return list
}
