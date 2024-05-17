package msx

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"server/settings"
	"server/torr"
	"server/utils"
	"server/version"
	"server/web/auth"

	"github.com/gin-gonic/gin"
)

const files, param = "files", "menu:request:interaction:init@{PREFIX}{SERVER}/msx/plugin.html"

var parameter = param

func SetupRoute(r gin.IRouter) {
	authorized := r.Group("/", auth.CheckAuth())
	authorized.GET("/msx/inf", func(c *gin.Context) {
		if p, o := c.GetQuery("parameter"); o {
			if p == "" {
				p = param
			}
			parameter = p
		}
		r := map[string]any{"version": version.Version, "search": settings.BTsets.EnableRutorSearch}
		if f, e := os.Stat(files); e == nil {
			r[files] = !f.IsDir()
		} else if !os.IsNotExist(e) {
			r[files] = e.Error()
		}
		c.JSON(200, r)
	})
	authorized.GET("/msx/trn", func(c *gin.Context) {
		if h := c.Query("indb"); h != "" {
			var r bool
			for _, t := range settings.ListTorrent() {
				if r = t.InfoHash.HexString() == h; r {
					break
				}
			}
			c.JSON(200, r)
		} else if h = c.Query("hash"); h != "" {
			st, sc := trn(h)
			if sc != "" {
				sc = "{col:" + sc + "}"
			}
			msx(c, map[string]string{"action": "player:label:position:{VALUE}{tb}{tb}" + sc + st})
		} else {
			c.AbortWithStatus(http.StatusBadRequest)
		}
	})
	authorized.POST("/msx/trn", func(c *gin.Context) {
		var j struct{ Data string }
		if e := json.NewDecoder(c.Request.Body).Decode(&j); e != nil {
			msx(c, e)
		} else {
			st, sc := trn(j.Data[strings.LastIndexByte(j.Data, ':')+1:])
			msx(c, map[string]any{"stamp": st, "stampColor": sc, "live": map[string]any{
				"type": "airtime", "duration": 1000, "over": map[string]any{
					"action": "execute:" + utils.GetScheme(c) + "://" + c.Request.Host + c.Request.URL.Path,
					"data":   j.Data,
				},
			}})
		}
	})
	authorized.Any("/msx/proxy", func(c *gin.Context) {
		proxy(c, c.Query("url"), c.QueryArray("header")...)
	})
	authorized.GET("/msx/start.json", func(c *gin.Context) {
		c.JSON(200, map[string]any{"name": "TorrServer", "version": version.Version, "parameter": parameter})
	})
	authorized.GET("/msx/:pth", func(c *gin.Context) {
		proxy(c, "https://damiva.github.io"+c.Request.URL.Path)
	})
	authorized.GET("/imdb/:id", func(c *gin.Context) {
		i, l, j := c.Param("id"), "", false
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
	authorized.Static("/files", files)
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
func msx(c *gin.Context, d any) {
	var r struct {
		R struct {
			S int    `json:"status"`
			T string `json:"text"`
			M string `json:"message,omitempty"`
			D any    `json:"data,omitempty"`
		} `json:"response"`
	}
	if e, o := d.(error); o {
		r.R.S = http.StatusBadRequest
		r.R.M = e.Error()
	} else {
		r.R.S = http.StatusOK
		r.R.D = d
	}
	r.R.T = http.StatusText(r.R.S)
	c.JSON(200, &r)
}
func trn(h string) (t, c string) {
	if h := torr.GetTorrent(h); h != nil {
		if h := h.Status(); h != nil && h.Stat < 5 {
			switch h.Stat {
			case 4:
				c = "msx-red"
			case 3:
				c = "msx-green"
			default:
				c = "msx-yellow"
			}
			t = "{ico:north} " + strconv.Itoa(h.ActivePeers) + " / " + strconv.Itoa(h.TotalPeers) + " {ico:south} " + strconv.Itoa(h.ConnectedSeeders)
		}
	}
	return
}
