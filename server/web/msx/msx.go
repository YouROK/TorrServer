package msx

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"server/settings"
	"server/torr"
	"server/utils"
	"server/version"
	"server/web/auth"

	"github.com/gin-gonic/gin"
)

const base, files = "tsmsx.yourok.ru", "media"

var param = "menu:request:interaction:{SERVER}@{PREFIX}" + base + "/start.html"

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

func rsp(c *gin.Context, r *http.Response, e error) {
	if e != nil {
		c.AbortWithError(http.StatusInternalServerError, e)
	} else {
		defer r.Body.Close()
		c.DataFromReader(r.StatusCode, r.ContentLength, r.Header.Get("Content-Type"), r.Body, nil)
	}
}

func SetupRoute(r gin.IRouter) {
	authorized := r.Group("/", auth.CheckAuth())
	// MSX:
	authorized.GET("/msx/", func(c *gin.Context) {
		r, e := http.Get("http://" + base)
		rsp(c, r, e)
	})
	authorized.GET("/msx/start.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{
			"name":      "TorrServer",
			"version":   version.Version,
			"parameter": param,
			"launcher": map[string]any{
				"type":  "start",
				"image": utils.GetScheme(c) + "://" + c.Request.Host + "/logo.png",
			},
		})
	})
	authorized.POST("/msx/start.json", func(c *gin.Context) {
		if e := c.BindJSON(&param); e != nil {
			c.AbortWithError(http.StatusBadRequest, e)
		}
	})
	authorized.GET("/msx/trn", func(c *gin.Context) {
		r := false
		if h := c.Query("hash"); h != "" {
			for _, t := range settings.ListTorrent() {
				if r = (t != nil && t.InfoHash.HexString() == h); r {
					break
				}
			}
		}
		c.JSON(http.StatusOK, r)
	})
	authorized.POST("/msx/trn", func(c *gin.Context) {
		var r struct {
			R struct {
				S int            `json:"status"`
				T string         `json:"text"`
				M string         `json:"message,omitempty"`
				D map[string]any `json:"data,omitempty"`
			} `json:"response"`
		}
		if j := struct{ Data string }{Data: c.Query("hash")}; j.Data != "" {
			st, sc := trn(j.Data)
			if sc != "" {
				sc = "{col:" + sc + "}"
			}
			r.R.S, r.R.D = http.StatusOK, map[string]any{"action": "player:label:position:{VALUE}{tb}{tb}" + sc + st}
		} else if e := c.BindJSON(&j); e != nil {
			r.R.S, r.R.M = http.StatusBadRequest, e.Error()
		} else if j.Data == "" {
			r.R.S, r.R.M = http.StatusBadRequest, "data is not set"
		} else {
			st, sc := trn(j.Data[strings.LastIndexByte(j.Data, ':')+1:])
			r.R.D = map[string]any{"stamp": st, "stampColor": sc}
			if sc != "" {
				r.R.D["live"] = map[string]any{
					"type": "airtime", "duration": 3000, "over": map[string]any{
						"action": "execute:" + utils.GetScheme(c) + "://" + c.Request.Host + c.Request.URL.Path, "data": j.Data,
					},
				}
			}
			r.R.S, r.R.D = http.StatusOK, map[string]any{"action": j.Data, "data": r.R.D}
		}
		r.R.T = http.StatusText(r.R.S)
		c.JSON(http.StatusOK, &r)
	})
	authorized.Any("/msx/proxy", func(c *gin.Context) {
		if u := c.Query("url"); u == "" {
			c.AbortWithStatus(http.StatusBadRequest)
		} else if q, e := http.NewRequest(c.Request.Method, u, c.Request.Body); e != nil {
			c.AbortWithError(http.StatusInternalServerError, e)
		} else {
			for _, v := range c.QueryArray("header") {
				if v := strings.SplitN(v, ":", 2); len(v) == 2 {
					q.Header.Add(v[0], v[1])
				}
			}
			r, e := http.DefaultClient.Do(q)
			rsp(c, r, e)
		}
	})
	authorized.GET("/msx/imdb/:id", func(c *gin.Context) {
		i, j := strings.TrimPrefix(c.Param("id"), "/"), false
		if j = strings.HasSuffix(i, ".json"); !j {
			i += ".json"
		}
		if r, e := http.Get("https://v2.sg.media-imdb.com/suggestion/h/" + i); e != nil || r.StatusCode != http.StatusOK || j {
			rsp(c, r, e)
		} else {
			var j struct {
				D []struct{ I struct{ ImageUrl string } }
			}
			if e = json.NewDecoder(r.Body).Decode(&j); e != nil {
				c.AbortWithError(http.StatusInternalServerError, e)
			} else if len(j.D) == 0 || j.D[0].I.ImageUrl == "" {
				c.Status(http.StatusNotFound)
			} else {
				c.Redirect(http.StatusMovedPermanently, j.D[0].I.ImageUrl)
			}
		}
	})
	// Files:
	authorized.StaticFS("/files/", gin.Dir(filepath.Join(settings.Path, files), true))
	authorized.GET("/files", func(c *gin.Context) {
		if l, e := os.Readlink(filepath.Join(settings.Path, files)); e == nil || os.IsNotExist(e) {
			c.JSON(http.StatusOK, l)
		} else {
			c.JSON(http.StatusInternalServerError, e.Error)
		}
	})
	authorized.POST("/files", func(c *gin.Context) {
		var l string
		if e := c.BindJSON(&l); e != nil {
			c.AbortWithError(http.StatusBadRequest, e)
		} else if e = os.Remove(filepath.Join(settings.Path, files)); e != nil && !os.IsNotExist(e) {
			c.AbortWithError(http.StatusInternalServerError, e)
		} else if l != "" {
			if f, e := os.Stat(l); e != nil {
				c.AbortWithError(http.StatusBadRequest, e)
			} else if !f.IsDir() {
				c.AbortWithError(http.StatusBadRequest, errors.New(l+" is not a directory"))
			} else if e = os.Symlink(l, filepath.Join(settings.Path, files)); e != nil {
				c.AbortWithError(http.StatusInternalServerError, e)
			}
		}
	})
}
