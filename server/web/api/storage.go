package api

import (
	"net/http"

	sets "server/settings"

	"github.com/gin-gonic/gin"
)

// GetStorageSettings godoc
// @Summary Get storage configuration settings
// @Description Retrieves the current storage preferences for settings and viewed history
// @Tags API
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Storage preferences"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /storage/settings [get]
func GetStorageSettings(c *gin.Context) {
	prefs := sets.GetStoragePreferences()
	c.JSON(http.StatusOK, prefs)
}

// UpdateStorageSettings godoc
// @Summary Update storage configuration settings
// @Description Updates the storage preferences for settings and viewed history. Requires application restart for changes to take effect.
// @Tags API
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Security ApiKeyAuth
// @Param request body map[string]interface{} true "Storage preferences to update"
// @Param settings formData string false "Settings storage type" Enums(json,bbolt)
// @Param viewed formData string false "Viewed history storage type" Enums(json,bbolt)
// @Success 200 {object} map[string]string "Update successful"
// @Failure 400 {object} map[string]string "Invalid input data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Read-only mode"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /storage/settings [post]
func UpdateStorageSettings(c *gin.Context) {
	if sets.ReadOnly {
		c.JSON(http.StatusForbidden, gin.H{"error": "Read-only mode"})
		return
	}

	var prefs map[string]interface{}

	// Check Content-Type to handle both JSON and form data
	contentType := c.GetHeader("Content-Type")

	if contentType == "application/x-www-form-urlencoded" {
		// Handle form data
		settings := c.PostForm("settings")
		viewed := c.PostForm("viewed")

		prefs = make(map[string]interface{})
		if settings != "" {
			prefs["settings"] = settings
		}
		if viewed != "" {
			prefs["viewed"] = viewed
		}
	} else {
		// Handle JSON (default)
		if err := c.ShouldBindJSON(&prefs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Validate preferences - only validate if provided
	if settingsPref, ok := prefs["settings"].(string); ok && settingsPref != "" {
		if settingsPref != "json" && settingsPref != "bbolt" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings storage value"})
			return
		}
	}

	if viewedPref, ok := prefs["viewed"].(string); ok && viewedPref != "" {
		if viewedPref != "json" && viewedPref != "bbolt" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid viewed storage value"})
			return
		}
	}

	// Check if we have at least one value to update
	if len(prefs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No preferences provided"})
		return
	}

	if err := sets.SetStoragePreferences(prefs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
