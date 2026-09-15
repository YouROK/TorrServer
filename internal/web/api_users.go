package web

import (
	"errors"
	"net/http"

	"silo/internal/log"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleListUsers(c *gin.Context) {
	val, _ := c.Get("user")
	actor := val.(*user.User)

	// Только Admin (50) и Owner (100) могут смотреть список всех пользователей
	if actor.Rank < 50 {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	users, err := s.userSvc.ListUsers()
	if err != nil {
		log.Errorf("[Web] Failed to list users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	// Скрываем хэши паролей из ответа API
	for _, u := range users {
		u.PasswordHash = ""
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (s *Server) handleCreateUser(c *gin.Context) {
	val, _ := c.Get("user")
	actor := val.(*user.User)

	var req struct {
		Username string        `json:"username" binding:"required"`
		Password string        `json:"password" binding:"required"`
		Rank     user.RoleRank `json:"rank"`
		Limits   user.Limits   `json:"limits"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	newUser, err := s.userSvc.CreateUser(actor, req.Username, req.Password, req.Rank, req.Limits)
	if err != nil {
		log.Warnf("[Web] Failed to create user '%s': %v", req.Username, err)
		handleUserError(c, err)
		return
	}

	newUser.PasswordHash = ""
	c.JSON(http.StatusCreated, newUser)
}

func (s *Server) handleUpdateUser(c *gin.Context) {
	val, _ := c.Get("user")
	actor := val.(*user.User)
	targetID := c.Param("id")

	var req struct {
		Password *string `json:"password"`
		Banned   *bool   `json:"banned"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if req.Banned != nil {
		if err := s.userSvc.BanUser(actor, targetID, *req.Banned); err != nil {
			handleUserError(c, err)
			return
		}
	}

	// TODO: Добавить смену пароля, если req.Password != nil

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) handleDeleteUser(c *gin.Context) {
	val, _ := c.Get("user")
	actor := val.(*user.User)
	targetID := c.Param("id")

	if err := s.userSvc.DeleteUser(actor, targetID); err != nil {
		handleUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (s *Server) handleRegenerateToken(c *gin.Context) {
	val, _ := c.Get("user")
	actor := val.(*user.User)
	targetID := c.Param("id")

	newToken, err := s.userSvc.RegenerateToken(actor, targetID)
	if err != nil {
		handleUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

// handleUserError маппит внутренние ошибки user.Service на HTTP-статусы
func handleUserError(c *gin.Context, err error) {
	if errors.Is(err, user.ErrPermissionDenied) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
	} else if errors.Is(err, user.ErrOwnerProtected) {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot modify or delete the owner"})
	} else if errors.Is(err, user.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// handleSetRank изменяет ранг пользователя
func (s *Server) handleSetRank(c *gin.Context) {
	val, _ := c.Get("user")
	actor := val.(*user.User)
	targetID := c.Param("id")

	var req struct {
		Rank user.RoleRank `json:"rank"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := s.userSvc.SetRank(actor, targetID, req.Rank); err != nil {
		handleUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "rank updated"})
}

// handleAdminListTorrents показывает торренты конкретного пользователя
func (s *Server) handleAdminListTorrents(c *gin.Context) {
	val, _ := c.Get("user")
	actor := val.(*user.User)
	targetID := c.Param("id")

	list, err := s.userSvc.AdminListTorrents(actor, targetID)
	if err != nil {
		handleUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"torrents": list})
}
