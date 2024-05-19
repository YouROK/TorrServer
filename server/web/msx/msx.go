package msx

import (
	"encoding/json"
	"errors"
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
	authorized.Any("/msx", mng)
	authorized.GET("/msx/*pth", func(c *gin.Context) { proxy(c, "https://damiva.github.io/msx"+c.Param("pth")) })

	authorized.GET("/files", fls)
	authorized.StaticFS("/files/", gin.Dir(files, true))

	authorized.GET("/imdb/:id", imdb)
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
func mng(c *gin.Context) {
	if p := c.Query("url"); p != "" {
		proxy(c, p, c.QueryArray("header")...)
	} else if c.Request.Method == "POST" {
		trn(c)
	} else if p = c.Query("indb"); p != "" {
		var r bool
		for _, t := range settings.ListTorrent() {
			if r = (t != nil && t.InfoHash.HexString() == p); r {
				break
			}
		}
		c.JSON(200, r)
	} else {
		if p, o := c.GetQuery("parameter"); o {
			if p == "" {
				parameter = param
			} else {
				parameter = p
			}
		}
		c.JSON(200, map[string]any{"version": version.Version, "search": settings.BTsets.EnableRutorSearch, "parameter": parameter})
	}
}
func trn(c *gin.Context) {
	var (
		h, a string
		q    struct{ Data any }
		r    struct {
			R struct {
				S int    `json:"status"`
				T string `json:"text"`
				M string `json:"message,omitempty"`
				D any    `json:"data,omitempty"`
			} `json:"response"`
		}
	)
	if e := json.NewDecoder(c.Request.Body).Decode(&q); e != nil {
		r.R.M = e.Error()
	} else if s, o := q.Data.(string); o {
		a, h = s, a[strings.LastIndexByte(s, ':')+1:]
	} else if s, o := q.Data.(map[string]any); o {
		if s, o := s["info"].(map[string]any); o {
			if s, o := s["content"].(map[string]any); o {
				h, _ = s["flag"].(string)
			}
		}
	}
	if h != "" {
		var st, sc string
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
		if a != "" {
			r.R.D = map[string]any{"action": a, "data": map[string]any{
				"stamp": st, "stampColor": sc, "live": map[string]any{
					"type": "airtime", "duration": 1000, "over": map[string]any{
						"action": "execute:" + utils.GetScheme(c) + c.Request.Host + c.Request.URL.Path, "data": a,
					},
				},
			}}
		} else {
			if sc != "" {
				sc = "{col:" + sc + "}"
			}
			r.R.D = map[string]string{"action": "player:label:position:{VALUE}{tb}{tb}" + sc + st}
		}
	} else if r.R.M == "" {
		r.R.M = "wrong data struct"
	}
	if r.R.D == nil {
		r.R.S = http.StatusBadRequest
	} else {
		r.R.S = http.StatusOK
	}
	r.R.T = http.StatusText(r.R.S)
	c.JSON(200, &r)
}
func fls(c *gin.Context) {
	var e error
	p, o := c.GetQuery("path")
	if o {
		if e = os.Remove(files); e != nil && os.IsNotExist(e) {
			e = nil
		}
		if e == nil && p != "" {
			var f os.FileInfo
			if f, e = os.Stat(p); e == nil {
				if f.IsDir() {
					e = os.Symlink(p, files)
				} else {
					e = errors.New(p + " is not a directory")
				}
			}
		}
	} else if p, e = os.Readlink(files); e != nil && os.IsNotExist(e) {
		e = nil
	}
	if e == nil {
		c.JSON(200, p)
	} else {
		c.AbortWithError(http.StatusInternalServerError, e)
	}
}
func imdb(c *gin.Context) {
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
}
