package web

import (
	"net/http"

	"silo/internal/log"
	"silo/internal/torrent"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// handleGetTorrentConfig отдает конфиг движка, хранящийся в базе данных
func (s *Server) handleGetTorrentConfig(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	// Настройки движка меняет только Owner
	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	cfg, err := s.torrentMgr.GetConfig()
	if err != nil {
		log.Errorf("[Web] Failed to read torrent config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read torrent config"})
		return
	}

	c.JSON(http.StatusOK, cfg)
}

// handleSaveTorrentConfig сохраняет конфиг движка в базу данных
func (s *Server) handleSaveTorrentConfig(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	var cfg torrent.Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config format"})
		return
	}

	if err := s.torrentMgr.SaveConfig(&cfg); err != nil {
		log.Errorf("[Web] Failed to save torrent config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save torrent config"})
		return
	}

	log.Infof("[Web] Torrent engine config saved by user %s", currentUser.Username)
	c.JSON(http.StatusOK, gin.H{"status": "saved", "restart_required": true})
}

// handleSetBlocklist обновляет динамический блок-лист IP
func (s *Server) handleSetBlocklist(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := s.torrentMgr.SetBlocklistText(req.Text); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "blocklist updated"})
}

// handleSetTrackers обновляет политику трекеров
func (s *Server) handleSetTrackers(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	var req struct {
		Mode     string   `json:"mode"`
		Trackers []string `json:"trackers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	s.torrentMgr.SetTrackerPolicy(torrent.TrackerMode(req.Mode), req.Trackers)
	c.JSON(http.StatusOK, gin.H{"status": "trackers updated"})
}
