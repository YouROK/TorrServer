package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	sets "server/settings"
)

// tmdbSettings godoc
//
//	@Summary		Get TMDB settings
//	@Description	Get TMDB API configuration
//
//	@Tags			API
//
//	@Produce		json
//	@Success		200	{object}	sets.TMDBConfig	"TMDB settings"
//	@Router			/tmdb/settings [get]
func tmdbSettings(c *gin.Context) {
	if sets.BTsets == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Settings not initialized"})
		return
	}
	c.JSON(200, sets.BTsets.TMDBSettings)
}
