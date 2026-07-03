package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	gstreamer "server/gstreamer/bridge"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

var (
	filteredSwaggerOnce sync.Once
	filteredSwaggerDoc  string
	filteredSwaggerErr  error
)

func swaggerHandler() gin.HandlerFunc {
	base := ginSwagger.WrapHandler(swaggerFiles.Handler)
	if gstreamer.BuiltIn() {
		return base
	}

	return func(c *gin.Context) {
		if strings.HasSuffix(c.Request.URL.Path, "doc.json") {
			doc, err := filteredSwaggerJSON()
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(doc))
			return
		}
		base(c)
	}
}

func filteredSwaggerJSON() (string, error) {
	filteredSwaggerOnce.Do(func() {
		raw, err := swag.ReadDoc()
		if err != nil {
			filteredSwaggerErr = err
			return
		}

		var spec map[string]any
		if err := json.Unmarshal([]byte(raw), &spec); err != nil {
			filteredSwaggerErr = err
			return
		}

		paths, _ := spec["paths"].(map[string]any)
		for path := range paths {
			if strings.HasPrefix(path, "/gst/") && path != "/gst/settings" {
				delete(paths, path)
			}
		}

		out, err := json.Marshal(spec)
		if err != nil {
			filteredSwaggerErr = err
			return
		}
		filteredSwaggerDoc = string(out)
	})

	return filteredSwaggerDoc, filteredSwaggerErr
}
