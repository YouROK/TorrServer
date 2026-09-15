package user

import "time"

// RoleRank — числовой ранг прав доступа (10..100)
type RoleRank int

const (
	RankOwner RoleRank = 100 // Владелец: пароль из config.yaml
	RankAdmin RoleRank = 50  // Админы: 50 .. 99
	RankUser  RoleRank = 10  // Пользователи: 10 .. 49
	RankGuest RoleRank = 1   // Гость: 1 .. 9 (только просмотр)
)

func (r RoleRank) IsOwner() bool { return r >= 100 }
func (r RoleRank) IsAdmin() bool { return r >= 50 && r < 100 }
func (r RoleRank) IsUser() bool  { return r >= 10 && r < 50 }

type Limits struct {
	MaxTorrents    int `json:"max_torrents"`     // 0 = без лимита
	DownloadRateKB int `json:"download_rate_kb"` // 0 = без ограничений скорости
}

type User struct {
	ID           string    `json:"id"`            // Уникальный ID (например, "u_admin")
	Username     string    `json:"username"`      // Уникальный логин
	PasswordHash string    `json:"password_hash"` // bcrypt хэш
	Rank         RoleRank  `json:"rank"`          // Числовой ранг
	APIToken     string    `json:"api_token"`     // Постоянный токен для плеера и API
	IsBanned     bool      `json:"is_banned"`     // Флаг блокировки
	Limits       Limits    `json:"limits"`        // Персональные квоты
	CreatedAt    time.Time `json:"created_at"`
}

// CanManage проверяет, может ли текущий пользователь управлять целевым (строго больше по рангу)
func (u *User) CanManage(target *User) bool {
	if u.IsBanned {
		return false
	}
	return u.Rank > target.Rank
}

// CanAssignRank проверяет, может ли пользователь назначить такой ранг
func (u *User) CanAssignRank(newRank RoleRank) bool {
	return u.Rank > newRank
}

// UserTorrent — карточка связи пользователя с конкретным торрентом
type UserTorrent struct {
	UserID      string    `json:"user_id"`
	TorrentHash string    `json:"torrent_hash"` // Хэш раздачи
	Title       string    `json:"title"`        // Название
	Poster      string    `json:"poster"`
	Category    string    `json:"category"`
	AddedAt     time.Time `json:"added_at"`     // Дата добавления пользователем
	ViewedFiles []int     `json:"viewed_files"` // Индексы просмотренных файлов: [0, 1, 3]
}

// MarkFileViewed помечает файл как просмотренный
func (ut *UserTorrent) MarkFileViewed(fileIdx int) {
	for _, idx := range ut.ViewedFiles {
		if idx == fileIdx {
			return
		}
	}
	ut.ViewedFiles = append(ut.ViewedFiles, fileIdx)
}

// UnmarkFileViewed снимает отметку о просмотре
func (ut *UserTorrent) UnmarkFileViewed(fileIdx int) {
	var filtered []int
	for _, idx := range ut.ViewedFiles {
		if idx != fileIdx {
			filtered = append(filtered, idx)
		}
	}
	ut.ViewedFiles = filtered
}

// IsFileViewed проверяет, просмотрен ли файл
func (ut *UserTorrent) IsFileViewed(fileIdx int) bool {
	for _, idx := range ut.ViewedFiles {
		if idx == fileIdx {
			return true
		}
	}
	return false
}
