package settings

import (
	"encoding/json"
	"server/log"
	"server/models"
	"server/settings/sqlite_models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SqliteDB struct {
	db *gorm.DB
}

func NewSqliteDB(dbPath string) (TorrServerDB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Create and migrate if needed all models
	err = db.AutoMigrate(
		&sqlite_models.SQLSetting{},
		&sqlite_models.SQLTorrent{},
		&sqlite_models.SQLTorrentFile{},
	)
	if err != nil {
		return nil, err
	}

	return &SqliteDB{db: db}, nil
}

func (s *SqliteDB) CloseDB() {
	// nothing to do
}

func (s *SqliteDB) Get(xPath, wantedName string) []byte {
	var torrent sqlite_models.SQLTorrent
	if err := s.db.Where(&sqlite_models.SQLTorrent{Hash: wantedName}).First(&torrent).Error; err != nil {
		log.TLogln(err)
		return []byte{}
	}

	buf, err := json.Marshal(fromSQLTorrent(torrent))
	if err != nil {
		log.TLogln(err)
		return []byte{}
	}

	return buf
}

func (s *SqliteDB) Set(xPath, name string, value []byte) {
	var torrentDB TorrentDB
	if err := json.Unmarshal(value, &torrentDB); err != nil {
		log.TLogln(err)
		return
	}

	torrent := torrentDB.toSQLTorrent()

	res := s.db.Clauses(clause.OnConflict{
		UpdateAll: true,
		Columns:   []clause.Column{{Name: "hash"}},
	}).Create(&torrent)

	if res.Error != nil {
		log.TLogln(res.Error)
		return
	}

	var tsFiles models.TsFiles
	if err := json.Unmarshal([]byte(torrentDB.Data), &tsFiles); err != nil {
		log.TLogln(err)
		return
	}

	for _, tsFile := range tsFiles.TorrServer.Files {
		if err := s.db.Clauses(clause.OnConflict{
			UpdateAll: true,
			Columns:   []clause.Column{{Name: "Path"}, {Name: "sql_torrent_id"}},
		}).Create(&sqlite_models.SQLTorrentFile{Path: tsFile.Path, Length: tsFile.Length, SQLTorrentID: torrent.ID}).Error; err != nil {
			log.TLogln(err)
			return
		}
	}
}

func (s *SqliteDB) List(xPath string) []string {
	var torrents []sqlite_models.SQLTorrent
	if err := s.db.Find(&torrents).Error; err != nil {
		log.TLogln(err)
		return []string{}
	}

	var hash []string
	for _, torrent := range torrents {
		hash = append(hash, torrent.Hash)
	}

	return hash
}

func (s *SqliteDB) Rem(xPath, name string) {
	if err := s.db.Clauses(clause.Returning{}).Where(&sqlite_models.SQLTorrent{Hash: name}).Delete(&sqlite_models.SQLTorrent{}).Error; err != nil {
		log.TLogln(err)
		return
	}
}
