package api

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/rutor/models"
	sets "server/settings"
	"server/torznab"
)

// torznabSearch godoc
//
//	@Summary		Makes a torznab search
//	@Description	Makes a torznab search.
//
//	@Tags			API
//
//	@Param			query	query	string	true	"Torznab query"
//
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{array}	models.TorrentDetails	"Torznab torrent search result(s)"
//	@Router			/torznab/search [get]
func torznabSearch(c *gin.Context) {
	if !sets.BTsets.EnableTorznabSearch {
		c.JSON(http.StatusBadRequest, gin.H{"error": "torznab disabled"})
		return
	}
	query := c.Query("query")
	indexStr := c.DefaultQuery("index", "-1")
	index := -1
	if i, err := strconv.Atoi(indexStr); err == nil {
		index = i
	}

	cat := c.Query("cat")
	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o > 0 {
		offset = o
	}
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 {
		limit = l
	}
	if limit > 250 {
		limit = 250
	}

	query, _ = url.QueryUnescape(query)
	list := torznab.Search(c.Request.Context(), query, index, cat, offset, limit)
	if list == nil {
		list = []*models.TorrentDetails{}
	}
	c.JSON(200, list)
}

type torznabCapsCategory struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	Subcategories []torznabCapsCategory `json:"subcategories,omitempty"`
}

type torznabCapsResponse struct {
	Limits struct {
		Max     int `json:"max"`
		Default int `json:"default"`
	} `json:"limits"`
	Categories []torznabCapsCategory `json:"categories"`
}

func torznabCapsCategoryFromModel(cat torznab.CapsCategory) torznabCapsCategory {
	out := torznabCapsCategory{
		ID:   cat.ID,
		Name: cat.Name,
	}
	for _, sub := range cat.Subcats {
		out.Subcategories = append(out.Subcategories, torznabCapsCategoryFromModel(sub))
	}
	return out
}

func torznabCaps(c *gin.Context) {
	if !sets.BTsets.EnableTorznabSearch {
		c.JSON(http.StatusBadRequest, gin.H{"error": "torznab disabled"})
		return
	}

	indexStr := c.Query("index")
	if indexStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index required"})
		return
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(sets.BTsets.TorznabUrls) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
		return
	}

	config := sets.BTsets.TorznabUrls[index]
	if config.Host == "" || config.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "indexer not configured"})
		return
	}

	caps, err := torznab.FetchCaps(c.Request.Context(), config.Host, config.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := torznabCapsResponse{}
	resp.Limits.Max = caps.Limits.Max
	resp.Limits.Default = caps.Limits.Default
	for _, cat := range caps.Categories {
		resp.Categories = append(resp.Categories, torznabCapsCategoryFromModel(cat))
	}
	c.JSON(200, resp)
}

type torznabTestReq struct {
	Host string `json:"host"`
	Key  string `json:"key"`
}

func torznabTest(c *gin.Context) {
	var req torznabTestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := torznab.Test(c.Request.Context(), req.Host, req.Key); err != nil {
		c.JSON(200, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}
