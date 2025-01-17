package settings

import (
	"encoding/json"
	"server/log"
	"server/models"
	"server/settings/sqlite_models"

	"golang.org/x/exp/maps"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const VIEWED_XPATH = "Viewed"
const TORRENTS_XPATH = "Torrents"

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

func (s *SqliteDB) Get(xPath, torrentHash string) []byte {
	switch xPath {
	case VIEWED_XPATH:
		return s.GetViewed(torrentHash)
	case TORRENTS_XPATH:
		return s.GetTorrents(torrentHash)
	default:
		log.TLogln("Unknown xpath:", xPath)
		return []byte{}
	}
}

func (s *SqliteDB) GetViewed(torrentHash string) []byte {
	var viewedFiles []sqlite_models.SQLTorrentFile
	err := s.db.Where("viewed = true AND sql_torrent_id = (SELECT id FROM torrents WHERE hash = ?)", torrentHash).
		Find(&viewedFiles).Error
	if err != nil {
		log.TLogln(err)
		return []byte{}
	}

	var viewedTorrentFiles = make(map[int]struct{})
	for _, viewedFile := range viewedFiles {
		viewedTorrentFiles[viewedFile.FileIndex] = struct{}{}
	}

	value, err := json.Marshal(viewedTorrentFiles)
	if err != nil {
		log.TLogln(err)
		return []byte{}
	}

	return value
}

func (s *SqliteDB) GetTorrents(torrentHash string) []byte {
	var torrent sqlite_models.SQLTorrent
	if err := s.db.Where(&sqlite_models.SQLTorrent{Hash: torrentHash}).First(&torrent).Error; err != nil {
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
	switch xPath {
	case VIEWED_XPATH:
		s.SetViewed(name, value)
	case TORRENTS_XPATH:
		s.SetTorrents(name, value)
	default:
		log.TLogln("Unknown xpath:", xPath)
		return
	}
}

func (s *SqliteDB) SetViewed(torrentHash string, value []byte) {
	jsonData := map[string]interface{}{}
	if err := json.Unmarshal(value, &jsonData); err != nil {
		log.TLogln("SetViewed error", err)
		return
	}

	for _, fileIndex := range maps.Keys(jsonData) {
		result := s.db.Model(&sqlite_models.SQLTorrentFile{}).
			Where("file_index = ? AND sql_torrent_id = (SELECT id FROM torrents WHERE hash = ?)", fileIndex, torrentHash).
			Update("viewed", true)

		if result.Error != nil {
			log.TLogln(result.Error)
			return
		}
	}
}

func (s *SqliteDB) SetTorrents(name string, value []byte) {
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
		}).Create(&sqlite_models.SQLTorrentFile{Path: tsFile.Path, Length: tsFile.Length, SQLTorrentID: torrent.ID, Viewed: false, FileIndex: tsFile.Id}).Error; err != nil {
			log.TLogln(err)
			return
		}
	}
}

func (s *SqliteDB) List(xPath string) []string {
	switch xPath {
	case VIEWED_XPATH:
		return s.ListViewed()
	case TORRENTS_XPATH:
		return s.ListTorrents()
	default:
		log.TLogln("Unknown xpath:", xPath)
		return []string{}
	}
}

func (s *SqliteDB) ListViewed() []string {
	log.TLogln("LISTING VIEWED")
	return []string{}
}

func (s *SqliteDB) ListTorrents() []string {
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
	switch xPath {
	case VIEWED_XPATH:
		s.RemViewed(name)
	case TORRENTS_XPATH:
		s.RemTorrent(name)
	default:
		log.TLogln("Unknown xpath:", xPath)
		return
	}
}

func (s *SqliteDB) RemViewed(name string) {
	log.TLogln("REM VIEWED")
}

func (s *SqliteDB) RemTorrent(name string) {
	if err := s.db.Clauses(clause.Returning{}).Where(&sqlite_models.SQLTorrent{Hash: name}).Delete(&sqlite_models.SQLTorrent{}).Error; err != nil {
		log.TLogln(err)
		return
	}
}
