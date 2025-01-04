package sqlite_models

import (
	"gorm.io/gorm"
)

type SQLTorrent struct {
	gorm.Model
	Hash        string `gorm:"unique"`
	Category    string
	Poster      string
	Size        int64
	Title       string
	DisplayName string
	ChunkSize   int
	Data        string `gorm:"-"`
	Files       []SQLTorrentFile
}

func (t *SQLTorrent) AfterDelete(tx *gorm.DB) error {
	var files = make([]SQLTorrentFile, 0)
	return tx.Where(&SQLTorrentFile{SQLTorrentID: t.ID}).Delete(&files).Error
}

func (SQLTorrent) TableName() string {
	return "torrents"
}
