package sqlite_models

import "gorm.io/gorm"

type SQLTorrentFile struct {
	gorm.Model
	SQLTorrentID uint   `gorm:"uniqueIndex:uniqueTorrentFileIndex"`
	Path         string `gorm:"uniqueIndex:uniqueTorrentFileIndex"`
	Length       int64
	FileIndex    int
	Viewed       bool
}

func (SQLTorrentFile) TableName() string {
	return "torrent_files"
}
