package bonjour

import (
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/grandcat/zeroconf"
	"github.com/wlynxg/anet"

	"server/log"
	"server/settings"
	"server/version"
)

const (
	serviceTorrServer = "_torrserver._tcp"
	serviceHTTP       = "_http._tcp"
	serviceHTTPS      = "_https._tcp"
)

var (
	mu      sync.Mutex
	servers []*zeroconf.Server
)

// Start advertises TorrServer over mDNS (_torrserver, _http, and _https when TLS is on).
func Start() {
	mu.Lock()
	defer mu.Unlock()
	stopLocked()

	port, err := strconv.Atoi(settings.Port)
	if err != nil || port <= 0 {
		log.TLogln("Bonjour: invalid web port", settings.Port)
		return
	}

	ifaces, ips := advertiseAddrs()
	if len(ips) == 0 {
		log.TLogln("Bonjour: no suitable interface addresses")
		return
	}

	host := mdnsHostname()
	name := instanceName()
	txt := baseTXT()

	register(name, serviceTorrServer, port, host, ips, txt, ifaces)
	register(name, serviceHTTP, port, host, ips, txt, ifaces)

	if settings.Ssl {
		if sslPort, err := strconv.Atoi(settings.SslPort); err == nil && sslPort > 0 {
			register(name, serviceHTTPS, sslPort, host, ips, []string{
				"version=" + version.Version,
				"path=/",
			}, ifaces)
		}
	}
}

// Stop withdraws mDNS advertisements.
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	stopLocked()
}

func stopLocked() {
	for _, s := range servers {
		s.Shutdown()
	}
	if len(servers) > 0 {
		log.TLogln("Bonjour: stopped")
	}
	servers = nil
}

func baseTXT() []string {
	txt := []string{
		"version=" + version.Version,
		"path=/",
	}
	if settings.Ssl && settings.SslPort != "" {
		txt = append(txt, "https="+settings.SslPort)
	}
	return txt
}

func register(name, service string, port int, host string, ips []string, txt []string, ifaces []net.Interface) {
	srv, err := zeroconf.RegisterProxy(name, service, "local.", port, host, ips, txt, ifaces)
	if err != nil {
		log.TLogln("Bonjour: register", service, "failed:", err)
		return
	}
	servers = append(servers, srv)
	log.TLogln("Bonjour: advertising", name, "as", service, "on", host+":"+strconv.Itoa(port), "("+strings.Join(ips, ",")+")")
}

func instanceName() string {
	if settings.BTsets != nil && settings.BTsets.FriendlyName != "" {
		return sanitizeInstance(settings.BTsets.FriendlyName)
	}
	return "TorrServer"
}

func sanitizeInstance(name string) string {
	name = stripLocalSuffix(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return "TorrServer"
	}
	if len(name) > 63 {
		return strings.TrimSpace(name[:63])
	}
	return name
}

func mdnsHostname() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return "torrserver"
	}
	host = stripLocalSuffix(host)
	host = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			return r
		}
		return '-'
	}, host)
	host = strings.Trim(host, "-")
	if host == "" {
		return "torrserver"
	}
	return host
}

// stripLocalSuffix removes a trailing ".local" so macOS dns-sd can browse the name.
func stripLocalSuffix(s string) string {
	s = strings.TrimSuffix(s, ".")
	if len(s) >= 6 && strings.EqualFold(s[len(s)-6:], ".local") {
		s = s[:len(s)-6]
	}
	return strings.TrimSuffix(s, ".")
}

func advertiseAddrs() ([]net.Interface, []string) {
	ifaces, err := anet.Interfaces()
	if err != nil {
		log.TLogln("Bonjour: list interfaces:", err)
		return nil, nil
	}

	var out []net.Interface
	var ips []string
	seen := make(map[string]bool)

	for _, i := range ifaces {
		if !usableIface(i) {
			continue
		}
		v4, v6 := addrsForIface(i)
		if len(v4) == 0 && len(v6) == 0 {
			continue
		}
		out = append(out, i)
		for _, ip := range append(v4, v6...) {
			s := ip.String()
			if seen[s] {
				continue
			}
			seen[s] = true
			ips = append(ips, s)
		}
	}
	return out, ips
}

func usableIface(i net.Interface) bool {
	if runtime.GOOS != "windows" && (i.Flags&net.FlagLoopback != 0 || i.Flags&net.FlagUp == 0 || i.Flags&net.FlagMulticast == 0) {
		return false
	}
	name := strings.ToLower(i.Name)
	for _, p := range []string{"utun", "awdl", "llw", "ap", "bridge", "anpi", "gif", "stf", "vmnet", "veth", "docker", "br-", "cni", "flannel"} {
		if strings.HasPrefix(name, p) {
			return false
		}
	}
	return true
}

func addrsForIface(i net.Interface) (v4, v6 []net.IP) {
	addrs, err := anet.InterfaceAddrsByInterface(&i)
	if err != nil {
		return nil, nil
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}
		if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			continue
		}
		if ip4 := ip.To4(); ip4 != nil {
			v4 = append(v4, ip4)
		} else if ip.To16() != nil {
			v6 = append(v6, ip)
		}
	}
	return v4, v6
}
