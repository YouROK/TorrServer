package waf

import (
	"net"
	"net/http"
	"strings"

	"server/log"

	"github.com/gin-gonic/gin"
)

// WAF returns Gin middleware that enforces IP white/blacklist and referer denylist.
// Lists are loaded at first use via Load() and can be hot-reloaded through the store.
// Built-in default referer blocks always run first and are never bypassed by the IP whitelist.
func WAF() gin.HandlerFunc {
	Load()

	return func(c *gin.Context) {
		snap := GetSnapshot()

		// Built-in + configured referer denylist is checked before the IP ACL.
		if host, blocked := isBlockedReferer(c.GetHeader("Referer"), c.GetHeader("Origin"), snap.Referers); blocked {
			log.WebLogln("Block referer:", host)
			ban(c)
			return
		}

		if !snap.IPEnabled {
			c.Next()
			return
		}

		ip := clientIP(c)
		if ip == nil {
			log.WebLogln("Block ip, unable to determine client address")
			ban(c)
			return
		}
		minifyIP(&ip)

		if snap.WhiteIP.NumRanges() > 0 {
			if _, ok := snap.WhiteIP.Lookup(ip); !ok {
				log.WebLogln("Block ip, not in white list", ip.String())
				ban(c)
				return
			}
		}
		if snap.BlackIP.NumRanges() > 0 {
			if r, ok := snap.BlackIP.Lookup(ip); ok {
				log.WebLogln("Block ip, in black list:", ip.String(), "in range", r.Description, ":", r.First, "-", r.Last)
				ban(c)
				return
			}
		}
		c.Next()
	}
}

func ban(c *gin.Context) {
	c.String(http.StatusForbidden, "Banned")
	c.Abort()
}

func clientIP(c *gin.Context) net.IP {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		// RemoteAddr without port (rare)
		host = strings.Trim(c.Request.RemoteAddr, "[]")
	}
	return net.ParseIP(host)
}
