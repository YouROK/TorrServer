package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/rutor"
	"server/rutor/models"
	sets "server/settings"
)

// rutorSearch godoc
//
//	@Summary		Makes a rutor search
//	@Description	Makes a rutor search.
//
//	@Tags			API
//
//	@Param			query	query	string	true	"Rutor query"
//
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{array}	models.TorrentDetails	"Rutor torrent search result(s)"
//	@Router			/search [get]
func rutorSearch(c *gin.Context) {
	if !sets.BTsets.EnableRutorSearch {
		c.JSON(http.StatusBadRequest, []string{})
		return
	}
	query := c.Query("query")
	query = unescapeQuery(query)
	list := rutor.Search(query)
	if list == nil {
		list = []*models.TorrentDetails{}
	}
	c.JSON(200, list)
}
