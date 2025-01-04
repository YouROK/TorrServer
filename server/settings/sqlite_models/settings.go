package sqlite_models

import "gorm.io/gorm"

type SQLSetting struct {
	gorm.Model
	Raw []byte
}

func (SQLSetting) TableName() string {
	return "settings"
}
