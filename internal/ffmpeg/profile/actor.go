package profile

import "errors"

// Actor описывает права того, от чьего имени выполняется операция с профилями.
// Модуль профилей не зависит от пакета пользователей: роли вычисляет вызывающий код,
// а сюда передаются уже готовые признаки прав.
type Actor struct {
	ID    string
	Rank  int  // Ранг для сравнения с RankRequired
	Admin bool // Может создавать и менять свои профили
	Owner bool // Может управлять любыми профилями и назначать профиль по умолчанию
}

// CanCreate сообщает, может ли пользователь создавать профили.
func (a Actor) CanCreate() bool {
	return a.Admin || a.Owner
}

// CanManage сообщает, может ли пользователь изменять указанный профиль.
func (a Actor) CanManage(p *Profile) bool {
	if p == nil {
		return false
	}
	if a.Owner {
		return true
	}
	return a.Admin && p.OwnerID == a.ID
}

// ValidateForActor проверяет профиль и права на его параметры.
// Нормализация выполняется до проверки: иначе незаполненные поля,
// у которых есть значение по умолчанию, считаются ошибкой.
func ValidateForActor(actor Actor, p *Profile) error {
	if p == nil {
		return errors.New("profile is required")
	}
	p.Normalize()

	if err := p.Validate(); err != nil {
		return err
	}
	if p.RankRequired > actor.Rank {
		return ErrPermission
	}
	return nil
}
