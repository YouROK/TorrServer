package user

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"silo/internal/bus"
	"silo/internal/config"
	"silo/internal/log"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrUserIsBanned     = errors.New("user is banned")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrOwnerProtected   = errors.New("cannot modify or delete the owner")
)

type Service struct {
	store  *Store
	cfg    *config.Config
	bus    *bus.Client
	ownerU *User // Виртуальный профиль Owner для режима без пароля
}

func NewService(store *Store, cfg *config.Config) *Service {
	s := &Service{
		store: store,
		cfg:   cfg,
		bus:   bus.Get("user_service"),
		ownerU: &User{
			ID:       "owner",
			Username: "owner",
			Rank:     RankOwner,
			IsBanned: false,
		},
	}

	// Инициализируем Owner в базе данных, если его еще нет
	owner, err := store.GetUserByID("owner")
	if errors.Is(err, ErrUserNotFound) {
		token, _ := s.generateUniqueToken()
		owner = &User{
			ID:        "owner",
			Username:  "owner",
			Rank:      RankOwner,
			APIToken:  token,
			IsBanned:  false,
			CreatedAt: time.Now(),
		}
		if err := store.SaveUser(owner); err != nil {
			log.Errorf("[User] Failed to save Owner profile to database: %v", err)
		} else {
			log.Infof("[User] Initialized Owner profile (API token: %s)", token)
		}
	}
	s.ownerU = owner

	return s
}

// IsAuthRequired возвращает true, если задан пароль Owner
func (s *Service) IsAuthRequired() bool {
	return s.cfg.Auth.OwnerPassword != ""
}

// ============================================================================
// АУТЕНТИФИКАЦИЯ (Проверка токена для плееров и веб-морды)
// ============================================================================

// Authenticate вызывается при каждом HTTP-запросе
func (s *Service) Authenticate(token string) (*User, error) {
	// 1. Если пароль не задан в config.yaml — пускаем всех как Owner!
	if !s.IsAuthRequired() {
		return s.ownerU, nil
	}

	// 2. Если пароль задан, но токен пустой — отказ
	if token == "" {
		return nil, ErrTokenNotFound
	}

	// 3. Ищем пользователя по токену в базе за O(1)
	u, err := s.store.GetUserByToken(token)
	if err != nil {
		return nil, err
	}

	if u.IsBanned {
		return nil, ErrUserIsBanned
	}

	return u, nil
}

// Login аутентифицирует по логину и паролю, возвращая токен
func (s *Service) Login(username, password string) (*User, string, error) {
	// Вход под Owner
	if username == "owner" {
		if s.IsAuthRequired() && password != s.cfg.Auth.OwnerPassword {
			return nil, "", ErrInvalidPassword
		}
		return s.ownerU, s.ownerU.APIToken, nil
	}

	// Вход для остальных пользователей
	u, err := s.store.GetUserByUsername(username)
	if err != nil {
		return nil, "", err
	}

	if u.IsBanned {
		return nil, "", ErrUserIsBanned
	}

	// Проверяем bcrypt хэш пароля
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidPassword
	}

	return u, u.APIToken, nil
}

// ============================================================================
// УПРАВЛЕНИЕ ПОЛЬЗОВАТЕЛЯМИ (С проверкой числовых рангов)
// ============================================================================

// CreateUser создает пользователя с проверкой прав создателя
func (s *Service) CreateUser(actor *User, username, password string, rank RoleRank, limits Limits) (*User, error) {
	// Проверяем, может ли создатель выдать такой ранг
	if !actor.CanAssignRank(rank) {
		return nil, ErrPermissionDenied
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newToken, err := s.generateUniqueToken()
	if err != nil {
		return nil, err
	}

	u := &User{
		ID:           "", // id создается автоматически через NextSequence в store
		Username:     username,
		PasswordHash: string(hash),
		Rank:         rank,
		APIToken:     newToken,
		IsBanned:     false,
		Limits:       limits,
		CreatedAt:    time.Now(),
	}

	if err := s.store.SaveUser(u); err != nil {
		return nil, err
	}

	log.Infof("[User] Created user '%s' (rank: %d)", u.Username, u.Rank)
	s.bus.Emit("user:created", u)

	return u, nil
}

// GetUserByID находит пользователя по его ID (включая профиль Owner)
func (s *Service) GetUserByID(id string) (*User, error) {
	if id == "owner" || (s.ownerU != nil && id == s.ownerU.ID) {
		return s.ownerU, nil
	}
	return s.store.GetUserByID(id)
}

// ListUsers возвращает список всех пользователей системы
func (s *Service) ListUsers() ([]*User, error) {
	return s.store.ListUsers()
}

// BanUser блокирует пользователя и сбрасывает его видеопотоки
func (s *Service) BanUser(actor *User, targetID string, banned bool) error {
	target, err := s.store.GetUserByID(targetID)
	if err != nil {
		return err
	}

	if target.Rank.IsOwner() {
		return ErrOwnerProtected
	}

	if !actor.CanManage(target) {
		return ErrPermissionDenied
	}

	target.IsBanned = banned
	if err := s.store.SaveUser(target); err != nil {
		return err
	}

	log.Warnf("[User] User '%s' ban status updated: %v", target.Username, banned)

	if banned {
		// Оповещаем стример, чтобы он мгновенно закрыл все видеопотоки этого юзера!
		s.bus.Emit("user:banned", target.ID)
	}

	return nil
}

// RegenerateToken перегенерирует токен доступа
func (s *Service) RegenerateToken(actor *User, targetID string) (string, error) {
	target, err := s.store.GetUserByID(targetID)
	if err != nil {
		return "", err
	}

	// Юзер может сбросить токен сам себе, либо это делает вышестоящий админ
	if actor.ID != target.ID && !actor.CanManage(target) {
		return "", ErrPermissionDenied
	}

	newToken, err := s.generateUniqueToken()
	if err != nil {
		return "", err
	}
	target.APIToken = newToken

	if err := s.store.SaveUser(target); err != nil {
		return "", err
	}

	log.Infof("[User] Generated new API token for user '%s'", target.Username)

	// Оповещаем систему о сбросе токена (старые плееры отключатся)
	s.bus.Emit("user:token:reset", map[string]string{
		"user_id": target.ID,
		"token":   newToken,
	})

	return newToken, nil
}

// DeleteUser удаляет пользователя и вычищает осиротевшие раздачи
func (s *Service) DeleteUser(actor *User, targetID string) error {
	target, err := s.store.GetUserByID(targetID)
	if err != nil {
		return err
	}

	if target.Rank.IsOwner() {
		return ErrOwnerProtected
	}

	if !actor.CanManage(target) {
		return ErrPermissionDenied
	}

	// Удаляем пользователя и получаем список хэшей, которые больше никто не держит
	orphanedHashes, err := s.store.DeleteUser(targetID)
	if err != nil {
		return err
	}

	log.Warnf("[User] Deleted user '%s'. Orphaned torrents: %d", target.Username, len(orphanedHashes))

	// Сигнализируем движку выгрузить эти торренты из RAM памяти
	for _, hash := range orphanedHashes {
		log.Infof("[Torrent] Torrent '%s' has no owners left. Dropping from memory...", hash)
		s.bus.Emit("torrent:drop", hash)
	}

	s.bus.Emit("user:deleted", targetID)
	return nil
}

// ============================================================================
// ПОЛЬЗОВАТЕЛЬСКИЕ ТОРРЕНТЫ (С контролем лимитов)
// ============================================================================

// AddTorrent добавляет торрент в список пользователя
func (s *Service) AddTorrent(u *User, hash, title, poster, category string) error {
	// Проверяем лимиты пользователя (если заданы)
	if u.Limits.MaxTorrents > 0 {
		currentCount, err := s.store.CountUserTorrents(u.ID)
		if err != nil {
			return err
		}
		if currentCount >= u.Limits.MaxTorrents {
			return fmt.Errorf("torrent limit reached for your account (max: %d)", u.Limits.MaxTorrents)
		}
	}

	ut := &UserTorrent{
		UserID:      u.ID,
		TorrentHash: hash,
		Title:       title,
		Poster:      poster,
		Category:    category,
		AddedAt:     time.Now(),
		ViewedFiles: []int{},
	}

	if err := s.store.AddUserTorrent(ut); err != nil {
		return err
	}

	log.Infof("[User:%s] Added torrent '%s' (%s)", u.Username, title, hash)

	// Оповещаем торрент-движок, что торрент добавлен
	s.bus.Emit("user:torrent:added", ut)
	return nil
}

// GetUserTorrent возвращает личную карточку торрента пользователя (с его постером и названием)
func (s *Service) GetUserTorrent(userID, hash string) (*UserTorrent, error) {
	return s.store.GetUserTorrent(userID, hash)
}

// RemoveTorrent удаляет торрент из списка пользователя
func (s *Service) RemoveTorrent(u *User, hash string) error {
	isLastOwner, err := s.store.RemoveUserTorrent(u.ID, hash)
	if err != nil {
		return err
	}

	log.Infof("[User:%s] Removed torrent '%s' from list", u.Username, hash)
	s.bus.Emit("user:torrent:removed", map[string]string{
		"user_id": u.ID,
		"hash":    hash,
	})

	// Если этот торрент не держит больше ни один другой пользователь — выгружаем из RAM
	if isLastOwner {
		log.Infof("[Torrent] Torrent '%s' has no owners left. Dropping from RAM cache...", hash)
		s.bus.Emit("torrent:drop", hash)
	}

	return nil
}

// SetFileViewed помечает файл просмотренным
func (s *Service) SetFileViewed(u *User, hash string, fileIdx int, viewed bool) error {
	if err := s.store.SetFileViewedStatus(u.ID, hash, fileIdx, viewed); err != nil {
		return err
	}

	s.bus.Emit("user:file:viewed", map[string]any{
		"user_id":  u.ID,
		"hash":     hash,
		"file_idx": fileIdx,
		"viewed":   viewed,
	})
	return nil
}

// ListTorrents возвращает список торрентов пользователя
func (s *Service) ListTorrents(u *User) ([]*UserTorrent, error) {
	return s.store.ListUserTorrents(u.ID)
}

// ============================================================================
// Вспомогательная генерация токенов
// ============================================================================

func (s *Service) generateUniqueToken() (string, error) {
	for {
		candidate := fmt.Sprintf("ts_%s", generateRandomString(24))

		_, err := s.store.GetUserByToken(candidate)
		if errors.Is(err, ErrTokenNotFound) {
			return candidate, nil
		}

		if err != nil {
			return "", err
		}
	}
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// SetRank изменяет ранг пользователя
func (s *Service) SetRank(actor *User, targetID string, newRank RoleRank) error {
	target, err := s.store.GetUserByID(targetID)
	if err != nil {
		return err
	}

	if target.Rank.IsOwner() {
		return ErrOwnerProtected
	}

	if !actor.CanAssignRank(newRank) {
		return ErrPermissionDenied
	}

	target.Rank = newRank
	if err := s.store.SaveUser(target); err != nil {
		return err
	}

	log.Infof("[User] Changed rank of '%s' to %d", target.Username, newRank)
	s.bus.Emit("user:rank:changed", map[string]any{
		"user_id":  target.ID,
		"new_rank": newRank,
	})
	return nil
}

// AdminListTorrents возвращает торренты указанного пользователя (для админки)
func (s *Service) AdminListTorrents(actor *User, targetID string) ([]*UserTorrent, error) {
	if actor.Rank < 50 {
		return nil, ErrPermissionDenied
	}

	// Проверяем, что целевой пользователь существует
	if _, err := s.store.GetUserByID(targetID); err != nil {
		return nil, err
	}

	return s.store.ListUserTorrents(targetID)
}

// UpdateTorrentMeta изменяет личную карточку торрента (название, постер, категория)
func (s *Service) UpdateTorrentMeta(u *User, hash, title, poster, category string) error {
	ut, err := s.store.GetUserTorrent(u.ID, hash)
	if err != nil {
		return err
	}

	ut.Title = title
	ut.Poster = poster
	ut.Category = category

	if err := s.store.AddUserTorrent(ut); err != nil {
		return err
	}

	log.Infof("[User:%s] Updated meta for torrent %s", u.Username, hash)
	s.bus.Emit("user:torrent:updated", map[string]string{
		"user_id": u.ID,
		"hash":    hash,
	})
	return nil
}
