package urlproxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/YouROK/tunsgo/p2p/utils"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (p *UrlProxy) GinHandler(c *gin.Context) {
	selfID := p.host.ID().String()
	if c.GetHeader("X-P2P-Server-ID") == selfID {
		c.JSON(http.StatusLoopDetected, gin.H{"error": "Infinite loop detected"})
		return
	}

	link := strings.TrimPrefix(c.Param("url"), "/")
	u, err := url.Parse(link)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}

	if utils.MatchHost(p.opts.Hosts, u.Hostname()) == true {
		//Local request
		req, err := http.NewRequest(c.Request.Method, link, c.Request.Body)
		if err == nil {
			for k, vv := range c.Request.Header {
				for _, v := range vv {
					req.Header.Add(k, v)
				}
			}
			req.Header.Set("X-P2P-Server-ID", selfID)
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				for k, vv := range resp.Header {
					for _, v := range vv {
						c.Header(k, v)
					}
				}
				c.Status(resp.StatusCode)
				io.Copy(c.Writer, resp.Body)
				resp.Body.Close()
				return
			}
		}
	}

	candidates := p.getCandidateProxies(u.Host)
	if len(candidates) == 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "no proxy nodes available"})
		return
	}

	for _, pID := range candidates {
		p.peers[pID].LastResp = time.Now()

		ctx := context.WithValue(c.Request.Context(), TargetPeerKey, pID)

		req, err := http.NewRequestWithContext(ctx, c.Request.Method, link, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for k, vv := range c.Request.Header {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}

		req.Header.Set("X-P2P-Server-ID", selfID)
		resp, err := p.httpClient.Do(req)
		if err != nil {
			p.host.ConnManager().UpsertTag(pID, "tuns-node", func(current int) int {
				if current > 10 {
					return current - 10
				}
				return 0
			})
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			continue
		}
		log.Printf("[REQ] Request to %s link: %s", pID.String(), link)

		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Header(k, v)
			}
		}
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
		resp.Body.Close()
		p.host.ConnManager().UpsertTag(pID, "tuns-node", func(current int) int {
			return 100
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "No nodes available"})
	return
}

func (p *UrlProxy) getCandidateProxies(targetHost string) []peer.ID {
	type candidate struct {
		id   peer.ID
		last time.Time
	}
	var list []candidate

	for pID, info := range p.peers {
		if utils.MatchHost(info.Hosts, targetHost) {
			list = append(list, candidate{pID, info.LastResp})
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].last.Before(list[j].last)
	})

	result := make([]peer.ID, len(list))
	for i, c := range list {
		result[i] = c.id
	}
	return result
}

func isLocalSelf(u *url.URL, selfPort int) bool {
	host := u.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" {
		if u.Port() == fmt.Sprintf("%d", selfPort) || u.Port() == "" && selfPort == 80 {
			return true
		}
	}
	return false
}
