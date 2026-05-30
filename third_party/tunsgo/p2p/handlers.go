package p2p

import "github.com/gin-gonic/gin"

// GinHandler godoc
//
//	@Summary		Universal P2P Proxy Tunnel
//	@Description	Proxies any HTTP request (GET, POST, PUT, DELETE, etc.) through the P2P network.
//	@Description	Uses a smart routing logic: local fetch if the host is allowed locally,
//	@Description	otherwise tunnels the raw traffic to a candidate P2P node.
//	@Tags			Proxy
//	@Accept			*/*
//	@Produce		*/*
//	@Param			url	path		string	true	"Full target URL"
//	@Success		200	{string}	string	"Proxied response"
//	@Failure		429	{json}		string	"Remote node slots busy"
//	@Failure		502	{json}		string	"No proxy nodes available"
//	@Router			/proxy/{url} [get]
//	@Router			/proxy/{url} [post]
//	@Router			/proxy/{url} [put]
//	@Router			/proxy/{url} [delete]
func (s *P2PServer) GinHandler(c *gin.Context) {
	s.urlprx.GinHandler(c)
}
