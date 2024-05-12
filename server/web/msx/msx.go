package msx

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"server/settings"
	"server/torr"
	"server/version"
	"server/web/auth"

	"github.com/gin-gonic/gin"
)

const base, fls = "https://damiva.github.io/msx", "files"

func SetupRoute(r gin.IRouter) {
	authorized := r.Group("/", auth.CheckAuth())
	authorized.Any("/msx", func(c *gin.Context) {
		if l := c.Query("url"); l != "" {
			proxy(c, l, c.QueryArray("header")...)
		} else if l = c.Query("indb"); l != "" {
			var r bool
			for _, t := range settings.ListTorrent() {
				if r = t.InfoHash.HexString() == l; r {
					break
				}
			}
			c.JSON(200, r)
		} else if c.Request.Method == "POST" {
			serve(c)
		} else {
			proxy(c, base+"/ts.html")
		}
	})
	authorized.GET("/msx/*pth", func(c *gin.Context) {
		p := c.Param("pth")
		if _, n := path.Split(p); n == "" {
			files(c, filepath.Join(fls, filepath.Clean(p)))
		} else if n = strings.ToLower(path.Ext(n)); n == "" {
			c.AbortWithStatus(http.StatusNotFound)
		} else if n == ".html" || n == ".js" || n == ".json" {
			proxy(c, base+p)
		} else {
			c.File(filepath.Join(fls, filepath.Clean(p)))
		}
	})
	authorized.GET("/imdb/:id", func(c *gin.Context) {
		const x = ".json"
		i, l := c.Param("id"), ""
		j := strings.HasSuffix(i, x)
		if i = strings.TrimSuffix(i, x); i != "" {
			if r, e := http.Get("https://v2.sg.media-imdb.com/suggestion/h/" + i + x); e == nil {
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

func serve(c *gin.Context) {
	var j struct {
		Data struct {
			Update string
			Info   struct{ Content struct{ Flag string } }
		}
	}
	if e := c.Bind(&j); e != nil {
		c.AbortWithError(http.StatusBadRequest, e)
	} else if j.Data.Update == "" && j.Data.Info.Content.Flag == "" {
		r := map[string]any{"version": version.Version, "search": settings.BTsets.EnableRutorSearch}
		if l, e := os.Readlink(fls); e == nil {
			r["files"] = l
		} else if !os.IsNotExist(e) {
			r["error"] = e.Error()
		}
		c.JSON(200, r)
	} else {
		var r map[string]any
		h, sc, st := j.Data.Info.Content.Flag, "", ""
		if h == "" {
			h = j.Data.Update[strings.LastIndexByte(j.Data.Update, ':')+1:]
		}
		if t := torr.GetTorrent(h); t != nil {
			if t := t.Status(); t != nil && t.Stat < 5 {
				switch t.Stat {
				case 4:
					sc = "msx-red"
				case 3:
					sc = "msx-green"
				default:
					sc = "msx-yellow"
				}
				st = "{ico:north} " + strconv.Itoa(t.ActivePeers) + " / " + strconv.Itoa(t.TotalPeers) + " {ico:south} " + strconv.Itoa(t.ConnectedSeeders)
			}
		}
		if j.Data.Update != "" {
			r = map[string]any{"action": "update:" + j.Data.Update, "data": map[string]string{"stamp": st, "stampColor": sc}}
		} else {
			if sc != "" {
				sc = "{tb}{tb}{col:" + sc + "}"
			}
			r = map[string]any{"action": "player:label:position:{LABEL}" + sc + st}
		}
		c.JSON(200, map[string]any{"response": map[string]any{"status": http.StatusOK, "data": r}})
	}
}

func files(c *gin.Context, p string) {
	if d, e := os.ReadDir(p); e == nil {
		var ds, fs []map[string]any
		u := c.Request.URL.EscapedPath()
		for _, f := range d {
			if n := f.Name(); f.IsDir() {
				ds = append(ds, map[string]any{"id": u + url.PathEscape(n) + "/", "path": n})
			} else if f, e := f.Info(); e == nil {
				fs = append(fs, map[string]any{"id": u + url.PathEscape(n), "path": n, "length": f.Size()})
			}
		}
		c.JSON(200, map[string]any{"title": filepath.Base(strings.TrimSuffix(p, "/")), "path": u, "files": append(ds, fs...)})
	} else if os.IsNotExist(e) {
		c.AbortWithError(http.StatusNotFound, e)
	} else {
		c.AbortWithError(http.StatusInternalServerError, e)
	}
}
