package web

import (
	"net/http"
	"time"

	"silo/internal/log"
	"silo/internal/user"
	"silo/internal/version"

	"github.com/gin-gonic/gin"
)

func (s *Server) handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleGetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version.Version,
		"torrent": version.GetTorrentVersion(),
	})
}

func (s *Server) handleGetLogs(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 50 {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	logs := log.GetRecentLogs()
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// handleGetStats отдает живые метрики движка и процесса для дашборда
func (s *Server) handleGetStats(c *gin.Context) {
	stats := s.torrentMgr.Stats()

	c.JSON(http.StatusOK, gin.H{
		"stats":      stats,
		"uptime_sec": int64(time.Since(s.startedAt).Seconds()),
		"version":    version.Version,
		"torrent":    version.GetTorrentVersion(),
	})
}

// handleShutdown gracefully завершает процесс сервера (только Owner).
// Ответ уходит клиенту раньше, чем процесс начнет остановку.
func (s *Server) handleShutdown(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	if s.shutdownFn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "shutdown hook not initialized"})
		return
	}

	log.Warnf("[Web] Shutdown requested by user %s", currentUser.Username)
	c.JSON(http.StatusOK, gin.H{"status": "shutting down"})

	// Даём ответу дойти до клиента, затем запускаем graceful shutdown
	go func() {
		time.Sleep(500 * time.Millisecond)
		s.shutdownFn()
	}()
}
