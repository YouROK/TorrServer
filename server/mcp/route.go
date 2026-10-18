package mcp

import (
	"github.com/gin-gonic/gin"
)

// Mount registers the Streamable HTTP MCP endpoint at /mcp.
// Callers should pass a router group that already applies HTTP Basic auth
// (the same CheckAuth used by POST /torrents).
func Mount(route gin.IRouter) {
	h := Handler()
	route.Any("/mcp", gin.WrapH(h))
	route.Any("/mcp/*path", gin.WrapH(h))
	logMCPReady()
}
