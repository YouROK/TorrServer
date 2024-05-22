package msx

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"server/settings"
	"server/torr"
	"server/utils"
	"server/version"
	"server/web/auth"

	"github.com/gin-gonic/gin"
)

const (
	files = `files`
	base  = `https://damiva.github.io/msx`
	htmlB = `<!DOCTYPE html><html><head><title>TorrServer MSX Plugin</title><meta charset="UTF-8" /><meta name="author" content="damiva" /><script type="text/javascript" src="http://msx.benzac.de/js/tvx-plugin.min.js"></script><script type="text/javascript" src="`
	htmlE = `.js"></script></head><body></body></html>`
)

var start = struct {
	N string `json:"name"`
	V string `json:"version"`
	P string `json:"parameter"`
}{"TorrServer", version.Version, "menu:request:interaction:init@{PREFIX}{SERVER}/msx/ts"}

func SetupRoute(r gin.IRouter) {
	authorized := r.Group("/", auth.CheckAuth())
	authorized.GET("/msx/start.json", func(c *gin.Context) {
		c.JSON(200, &start)
	})
	authorized.POST("/msx/start,json", func(c *gin.Context) {
		if e := c.Bind(&start); e != nil {
			c.AbortWithError(http.StatusBadRequest, e)
		}
	})
	authorized.Any("/msx/proxy", func(c *gin.Context) {
		proxy(c, c.Query("url"), c.QueryArray("header")...)
	})
	authorized.GET("/msx/trn", func(c *gin.Context) {
		if h := c.Query("hash"); h != "" {
			var r bool
			for _, t := range settings.ListTorrent() {
				if r = (t != nil && t.InfoHash.HexString() == h); r {
					break
				}
			}
			c.JSON(200, r)
		} else {
			c.AbortWithStatus(http.StatusBadRequest)
		}
	})
	authorized.POST("/msx/trn", func(c *gin.Context) {
		if j := struct{ Data string }{Data: c.Query("hash")}; j.Data != "" {
			st, sc := trn(j.Data)
			if sc != "" {
				sc = "{col:" + sc + "}"
			}
			msx(c, map[string]string{"action": "player:label:position:{VALUE}{tb}{tb}" + sc + st})
		} else if e := c.Bind(&j); e != nil {
			msx(c, e)
		} else if j.Data == "" {
			msx(c, errors.New("data is not set"))
		} else {
			st, sc := trn(j.Data[strings.LastIndexByte(j.Data, ':')+1:])
			msx(c, map[string]any{"stamp": st, "stampColor": sc, "live": map[string]any{
				"stamp": st, "stampColor": sc, "live": map[string]any{
					"type": "airtime", "duration": 1000, "over": map[string]any{
						"action": "execute:" + utils.GetScheme(c) + c.Request.Host + c.Request.URL.Path, "data": j.Data,
					},
				},
			}})
		}
	})
	authorized.GET("/msx/*pth", func(c *gin.Context) {
		if p := c.Param("pth"); !strings.HasSuffix(p, "/") && path.Ext(p) == "" {
			c.Data(200, "text/html;charset=UTF-8", append(append(append([]byte(htmlB), base...), p...), htmlE...))
		} else {
			proxy(c, base+p)
		}
	})

	authorized.GET("/files", func(c *gin.Context) {
		if l, e := os.Readlink(files); e == nil || os.IsNotExist(e) {
			c.JSON(200, l)
		} else {
			c.AbortWithError(http.StatusInternalServerError, e)
		}
	})
	authorized.POST("/files", func(c *gin.Context) {
		var l string
		if e := c.Bind(&l); e != nil {
			c.AbortWithError(http.StatusBadRequest, e)
		} else if e = os.Remove(files); e != nil && !os.IsNotExist(e) {
			c.AbortWithError(http.StatusInternalServerError, e)
		} else if l != "" {
			if e = os.Symlink(l, files); e != nil {
				c.AbortWithError(http.StatusInternalServerError, e)
			}
		}
	})
	authorized.StaticFS("/files/", gin.Dir(files, true))

	authorized.GET("/imdb/:id", func(c *gin.Context) {
		i, l, j := strings.TrimPrefix(c.Param("id"), "/"), "", false
		if j = strings.HasSuffix(i, ".json"); !j {
			i += ".json"
		}
		if r, e := http.Get("https://v2.sg.media-imdb.com/suggestion/h/" + i); e == nil {
			if r.StatusCode == http.StatusOK {
				var j struct {
					D []struct{ I struct{ ImageUrl string } }
				}
				if e = json.NewDecoder(r.Body).Decode(&j); e == nil && len(j.D) > 0 {
					l = j.D[0].I.ImageUrl
				}
			}
			r.Body.Close()
		}
		if j {
			c.JSON(200, l)
		} else if l == "" {
			c.Status(http.StatusNotFound)
		} else {
			c.Redirect(http.StatusMovedPermanently, l)
		}
	})
}
func proxy(c *gin.Context, u string, h ...string) {
	if u == "" {
		c.AbortWithStatus(http.StatusBadRequest)
	} else if q, e := http.NewRequest(c.Request.Method, u, c.Request.Body); e != nil {
		c.AbortWithError(http.StatusInternalServerError, e)
	} else {
		for _, v := range h {
			if v := strings.SplitN(v, ":", 2); len(v) == 2 {
				q.Header.Add(v[0], v[1])
			}
		}
		if r, e := http.DefaultClient.Do(q); e != nil {
			c.AbortWithError(http.StatusInternalServerError, e)
		} else {
			c.DataFromReader(r.StatusCode, r.ContentLength, r.Header.Get("Content-Type"), r.Body, nil)
			r.Body.Close()
		}
	}
}
func trn(h string) (st, sc string) {
	if h := torr.GetTorrent(h); h != nil {
		if h := h.Status(); h != nil && h.Stat < 5 {
			switch h.Stat {
			case 4:
				sc = "msx-red"
			case 3:
				sc = "msx-green"
			default:
				sc = "msx-yellow"
			}
			st = "{ico:north} " + strconv.Itoa(h.ActivePeers) + " / " + strconv.Itoa(h.TotalPeers) + " {ico:south} " + strconv.Itoa(h.ConnectedSeeders)
		}
	}
	return
}
func msx(c *gin.Context, a any) {
	var r struct {
		R struct {
			S int    `json:"status"`
			T string `json:"text"`
			M string `json:"message,omitempty"`
			D any    `json:"data,omitempty"`
		} `json:"response"`
	}
	if e, o := a.(error); o {
		r.R.S, r.R.M = http.StatusBadRequest, e.Error()
	} else {
		r.R.S, r.R.D = http.StatusOK, a
	}
	r.R.T = http.StatusText(r.R.S)
	c.JSON(200, &r)
}
