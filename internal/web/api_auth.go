package web

import (
	"net/http"

	"silo/internal/log"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	u, token, err := s.userSvc.Login(req.Username, req.Password)
	if err != nil {
		log.Warnf("[Web] Failed login attempt for user '%s': %v", req.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// Устанавливаем Cookie на 30 дней
	c.SetCookie(CookieTokenName, token, 3600*24*30, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"user": u, "token": token})
}

func (s *Server) handleLogout(c *gin.Context) {
	// Удаляем Cookie
	c.SetCookie(CookieTokenName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

func (s *Server) handleGetMe(c *gin.Context) {
	val, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}
	currentUser := val.(*user.User)
	c.JSON(http.StatusOK, currentUser)
}
