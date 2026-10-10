package web

import (
	"errors"
	"net/http"

	"silo/internal/ffmpeg/profile"
	"silo/internal/log"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// handleListProfiles отдаёт профили, доступные текущему пользователю.
func (s *Server) handleListProfiles(c *gin.Context) {
	actor := profileActor(c)

	if s.profiles == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "profiles are not available"})
		return
	}

	list, err := s.profiles.List(actor)
	if err != nil {
		writeProfileError(c, err)
		return
	}
	if list == nil {
		list = []*profile.Profile{}
	}

	// Владелец дополнительно видит полный список для администрирования
	if actor.Owner {
		if all, err := s.profiles.ListAll(actor); err == nil {
			c.JSON(http.StatusOK, gin.H{"profiles": list, "all": all})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"profiles": list})
}

// handleGetProfile отдаёт один профиль.
func (s *Server) handleGetProfile(c *gin.Context) {
	actor := profileActor(c)

	if s.profiles == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "profiles are not available"})
		return
	}

	p, err := s.profiles.Get(actor, c.Param("id"))
	if err != nil {
		writeProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

// handleCreateProfile создаёт профиль транскодирования.
func (s *Server) handleCreateProfile(c *gin.Context) {
	actor := profileActor(c)

	if s.profiles == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "profiles are not available"})
		return
	}

	var req profile.Profile
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile format"})
		return
	}

	if err := profile.ValidateForActor(actor, &req); err != nil {
		writeProfileError(c, err)
		return
	}

	created, err := s.profiles.Create(actor, &req)
	if err != nil {
		writeProfileError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

// handleUpdateProfile изменяет профиль.
func (s *Server) handleUpdateProfile(c *gin.Context) {
	actor := profileActor(c)

	if s.profiles == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "profiles are not available"})
		return
	}

	var req profile.Profile
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile format"})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.profiles.Update(actor, c.Param("id"), &req)
	if err != nil {
		writeProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// handleDeleteProfile удаляет профиль.
func (s *Server) handleDeleteProfile(c *gin.Context) {
	actor := profileActor(c)

	if s.profiles == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "profiles are not available"})
		return
	}

	if err := s.profiles.Delete(actor, c.Param("id")); err != nil {
		writeProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// handleSetDefaultProfile назначает профиль используемым по умолчанию.
func (s *Server) handleSetDefaultProfile(c *gin.Context) {
	actor := profileActor(c)

	if s.profiles == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "profiles are not available"})
		return
	}

	p, err := s.profiles.SetDefault(actor, c.Param("id"))
	if err != nil {
		writeProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

// writeProfileError переводит ошибку профилей в HTTP-ответ.
func writeProfileError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, profile.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
	case errors.Is(err, profile.ErrPermission):
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
	case errors.Is(err, profile.ErrNameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "profile name is already used"})
	case errors.Is(err, profile.ErrBuiltin):
		c.JSON(http.StatusBadRequest, gin.H{"error": "builtin profile cannot be changed"})
	default:
		log.Errorf("[Web] Profile request failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// profileActor собирает права пользователя для операций с профилями.
func profileActor(c *gin.Context) profile.Actor {
	val, _ := c.Get("user")
	u, _ := val.(*user.User)
	if u == nil {
		return profile.Actor{}
	}

	return profile.Actor{
		ID:    u.ID,
		Rank:  int(u.Rank),
		Admin: u.Rank >= user.RankAdmin,
		Owner: u.Rank >= user.RankOwner,
	}
}

// handleSetPreferredProfile сохраняет профиль, выбранный пользователем.
func (s *Server) handleSetPreferredProfile(c *gin.Context) {
	val, _ := c.Get("user")
	actor := val.(*user.User)

	var req struct {
		ProfileID string `json:"profile_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Выбранный профиль должен быть доступен пользователю
	if req.ProfileID != "" && req.ProfileID != profile.DefaultID {
		available, err := s.profiles.List(profileActor(c))
		if err != nil {
			writeProfileError(c, err)
			return
		}

		var found bool
		for _, p := range available {
			if p.ID == req.ProfileID {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusForbidden, gin.H{"error": "profile is not available"})
			return
		}
	}

	if err := s.userSvc.SetTranscodeProfile(actor, req.ProfileID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "saved", "profile_id": req.ProfileID})
}
