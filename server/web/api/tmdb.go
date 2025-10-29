package api

import (
	"net/http"
	
	"server/log"
	set "server/settings"
	"server/tmdb"
	
	"github.com/gin-gonic/gin"
)

type tmdbSearchRequest struct {
	Query    string `json:"query"`
	Language string `json:"language,omitempty"`
	Year     string `json:"year,omitempty"`
	Type     string `json:"type,omitempty"` // "movie", "tv", or "auto"
}

type tmdbSearchResponse struct {
	Success bool          `json:"success"`
	Data    *tmdb.Media   `json:"data,omitempty"`
	Posters []string      `json:"posters,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// tmdbSearch godoc
//
//	@Summary		Search TMDB for metadata
//	@Description	Search movie/TV show metadata from TMDB
//	@Tags			API
//	@Param			request	body	tmdbSearchRequest	true	"Search request"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	tmdbSearchResponse
//	@Router			/tmdb/search [post]
func tmdbSearch(c *gin.Context) {
	var req tmdbSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tmdbSearchResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Check if TMDB API key is configured
	if set.BTsets.TMDBApiKey == "" {
		c.JSON(http.StatusOK, tmdbSearchResponse{
			Success: false,
			Error:   "TMDB API key not configured",
		})
		return
	}
	
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, tmdbSearchResponse{
			Success: false,
			Error:   "Query is required",
		})
		return
	}
	
	client := tmdb.NewClient(set.BTsets.TMDBApiKey)
	
	var media *tmdb.Media
	var err error
	
	// Determine search type
	switch req.Type {
	case "movie":
		media, err = client.SearchMovie(req.Query, req.Year)
	case "tv":
		media, err = client.SearchTV(req.Query, req.Year)
	case "multi":
		// Combined search (movies + TV shows)
		media, err = client.SearchMulti(req.Query)
	default:
		// Auto-detect from torrent name
		media, err = client.SearchAuto(req.Query)
	}
	
	if err != nil {
		log.TLogln("TMDB search error:", err)
		c.JSON(http.StatusOK, tmdbSearchResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	// Generate poster URLs (multiple sizes)
	posters := []string{}
	if media.PosterPath != "" {
		posters = append(posters, media.GetPosterURL())
		// Add different sizes
		posters = append(posters, "https://image.tmdb.org/t/p/w300"+media.PosterPath)
		posters = append(posters, "https://image.tmdb.org/t/p/w780"+media.PosterPath)
	}
	
	c.JSON(http.StatusOK, tmdbSearchResponse{
		Success: true,
		Data:    media,
		Posters: posters,
	})
}
