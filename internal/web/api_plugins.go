package web

import (
	"errors"
	"io"
	"net/http"
	"os"

	"silo/internal/log"
	"silo/internal/plugin"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleListPlugins(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	if s.pluginMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "plugin manager not initialized"})
		return
	}

	plugins := s.pluginMgr.ListPlugins()
	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}

func (s *Server) handleUploadPlugin(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	if s.pluginMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "plugin manager not initialized"})
		return
	}

	file, header, err := c.Request.FormFile("plugin")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file"})
		return
	}
	defer file.Close()

	tmp, err := os.CreateTemp("", "silo-plugin-upload-*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file"})
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	tmp.Close()

	_, update := c.GetQuery("update")

	var installed []string
	if update {
		installed, err = s.pluginMgr.UpdatePlugin(tmpPath)
	} else {
		installed, err = s.pluginMgr.InstallPlugin(tmpPath)
	}

	if err != nil {
		var existsErr *plugin.PluginExistsError
		if errors.As(err, &existsErr) {
			c.JSON(http.StatusConflict, gin.H{
				"error":      existsErr.Error(),
				"code":       "plugin_exists",
				"id":         existsErr.ID,
				"version":    existsErr.Version,
				"is_builtin": existsErr.IsBuiltin,
			})
			return
		}
		log.Errorf("[Web] Failed to install plugin: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Infof("[Web] %d plugin(s) installed from '%s' by user %s (update=%v)",
		len(installed), header.Filename, currentUser.Username, update)
	c.JSON(http.StatusOK, gin.H{
		"status":    "installed",
		"filename":  header.Filename,
		"updated":   update,
		"installed": installed,
	})
}

func (s *Server) handleDeletePlugin(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	if s.pluginMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "plugin manager not initialized"})
		return
	}

	pluginID := c.Param("id")

	if s.pluginMgr.IsBuiltin(pluginID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "cannot delete built-in plugin",
			"code":  "plugin_builtin",
			"id":    pluginID,
		})
		return
	}

	if err := s.pluginMgr.UninstallPlugin(pluginID); err != nil {
		log.Errorf("[Web] Failed to uninstall plugin '%s': %v", pluginID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Infof("[Web] Plugin '%s' uninstalled by user %s", pluginID, currentUser.Username)
	c.JSON(http.StatusOK, gin.H{"status": "uninstalled"})
}

func (s *Server) handleSetPluginOrder(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	if s.pluginMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "plugin manager not initialized"})
		return
	}

	var req struct {
		Order []string `json:"order" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := s.pluginMgr.SetPluginOrder(req.Order); err != nil {
		log.Errorf("[Web] Failed to set plugin order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Infof("[Web] Plugin order updated by user %s", currentUser.Username)
	c.JSON(http.StatusOK, gin.H{"status": "order updated"})
}

// handlePluginInfo отдает полный манифест плагина для модального окна
func (s *Server) handlePluginInfo(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	man, err := s.pluginMgr.GetPluginInfo(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          man.ID,
		"name":        man.Name,
		"version":     man.Version,
		"author":      man.Author,
		"description": man.Description,
		"theme_ui":    man.ThemeUI,
		"entry":       man.Entry,
		"events":      man.Events,
		"routes":      man.Routes,
		"builtin":     s.pluginMgr.IsBuiltin(man.ID),
	})
}

// handlePluginEnable переключает флаг включенности плагина
func (s *Server) handlePluginEnable(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := s.pluginMgr.SetPluginEnabled(c.Param("id"), req.Enabled); err != nil {
		log.Errorf("[Web] Failed to toggle plugin '%s': %v", c.Param("id"), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// handleInstallPluginFromURL устанавливает плагин по прямой ссылке
func (s *Server) handleInstallPluginFromURL(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if currentUser.Rank < 100 {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}

	if s.pluginMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "plugin manager not initialized"})
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	_, update := c.GetQuery("update")

	var installed []string
	var err error
	if update {
		installed, err = s.pluginMgr.UpdateFromURL(req.URL)
	} else {
		installed, err = s.pluginMgr.InstallFromURL(req.URL)
	}

	if err != nil {
		var existsErr *plugin.PluginExistsError
		if errors.As(err, &existsErr) {
			c.JSON(http.StatusConflict, gin.H{
				"error":      existsErr.Error(),
				"code":       "plugin_exists",
				"id":         existsErr.ID,
				"version":    existsErr.Version,
				"is_builtin": existsErr.IsBuiltin,
			})
			return
		}
		log.Errorf("[Web] Failed to install plugin from URL %s: %v", req.URL, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Infof("[Web] %d plugin(s) installed from URL by user %s (update=%v)",
		len(installed), currentUser.Username, update)
	c.JSON(http.StatusOK, gin.H{
		"status":    "installed",
		"updated":   update,
		"installed": installed,
	})
}

// handlePluginMenu отдает пункты меню активных плагинов, отфильтрованные по рангу
func (s *Server) handlePluginMenu(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	if s.pluginMgr == nil {
		c.JSON(http.StatusOK, gin.H{"menu": []any{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"menu": s.pluginMgr.MenuFor(int(currentUser.Rank))})
}
