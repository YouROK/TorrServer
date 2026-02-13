package settings

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"server/log"
)

type TorznabConfig struct {
	Host string
	Key  string
	Name string
}

type TMDBConfig struct {
	APIKey     string // TMDB API Key
	APIURL     string // Base API URL (default: https://api.themoviedb.org)
	ImageURL   string // Image URL (default: https://image.tmdb.org)
	ImageURLRu string // Image URL for Russian users (default: https://imagetmdb.com)
}

type BTSets struct {
	// Cache
	CacheSize       int64 // in byte, def 64 MB
	ReaderReadAHead int   // in percent, 5%-100%, [...S__X__E...] [S-E] not clean
	PreloadCache    int   // in percent

	// Disk
	UseDisk           bool
	TorrentsSavePath  string
	RemoveCacheOnDrop bool

	// Torrent
	ForceEncrypt             bool
	RetrackersMode           int  // 0 - don`t add, 1 - add retrackers (def), 2 - remove retrackers 3 - replace retrackers
	TorrentDisconnectTimeout int  // in seconds
	EnableDebug              bool // debug logs

	// DLNA
	EnableDLNA   bool
	FriendlyName string

	// Rutor
	EnableRutorSearch bool

	// Torznab
	EnableTorznabSearch bool
	TorznabUrls         []TorznabConfig

	// TMDB
	TMDBSettings TMDBConfig

	// BT Config
	EnableIPv6        bool
	DisableTCP        bool
	DisableUTP        bool
	DisableUPNP       bool
	DisableDHT        bool
	DisablePEX        bool
	DisableUpload     bool
	DownloadRateLimit int // in kb, 0 - inf
	UploadRateLimit   int // in kb, 0 - inf
	ConnectionsLimit  int
	PeersListenPort   int

	// HTTPS
	SslPort int
	SslCert string
	SslKey  string

	// Reader
	ResponsiveMode bool // enable Responsive reader (don't wait pieceComplete)

	// FS
	ShowFSActiveTorr bool

	// Storage preferences
	StoreSettingsInJson bool
	StoreViewedInJson   bool
}

func (v *BTSets) String() string {
	buf, _ := json.Marshal(v)
	return string(buf)
}

var BTsets *BTSets

func SetBTSets(sets *BTSets) {
	if ReadOnly {
		return
	}
	// failsafe checks (use defaults)
	if sets.CacheSize == 0 {
		sets.CacheSize = 64 * 1024 * 1024
	}
	if sets.ConnectionsLimit == 0 {
		sets.ConnectionsLimit = 25
	}
	if sets.TorrentDisconnectTimeout == 0 {
		sets.TorrentDisconnectTimeout = 30
	}

	if sets.ReaderReadAHead < 5 {
		sets.ReaderReadAHead = 5
	}
	if sets.ReaderReadAHead > 100 {
		sets.ReaderReadAHead = 100
	}

	if sets.PreloadCache < 0 {
		sets.PreloadCache = 0
	}
	if sets.PreloadCache > 100 {
		sets.PreloadCache = 100
	}

	if sets.TorrentsSavePath == "" {
		sets.UseDisk = false
	} else if sets.UseDisk {
		// apply environment overrides before persisting
		applyEnvOverrides(sets)
		BTsets = sets

		go filepath.WalkDir(sets.TorrentsSavePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && strings.ToLower(d.Name()) == ".tsc" {
				BTsets.TorrentsSavePath = path
				log.TLogln("Find directory \"" + BTsets.TorrentsSavePath + "\", use as cache dir")
				return io.EOF
			}
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		})
	}

	BTsets = sets
	buf, err := json.Marshal(BTsets)
	if err != nil {
		log.TLogln("Error marshal btsets", err)
		return
	}
	tdb.Set("Settings", "BitTorr", buf)
}

func SetDefaultConfig() {
	sets := new(BTSets)
	sets.CacheSize = 64 * 1024 * 1024 // 64 MB
	sets.PreloadCache = 50
	sets.ConnectionsLimit = 25
	sets.RetrackersMode = 1
	sets.TorrentDisconnectTimeout = 30
	sets.ReaderReadAHead = 95 // 95%
	sets.ResponsiveMode = true
	sets.ShowFSActiveTorr = true
	sets.StoreSettingsInJson = true
	// Set default TMDB settings
	sets.TMDBSettings = TMDBConfig{
		APIKey:     "",
		APIURL:     "https://api.themoviedb.org",
		ImageURL:   "https://image.tmdb.org",
		ImageURLRu: "https://imagetmdb.com",
	}
	// apply environment overrides so envs can change defaults
	applyEnvOverrides(sets)
	BTsets = sets
	if !ReadOnly {
		buf, err := json.Marshal(BTsets)
		if err != nil {
			log.TLogln("Error marshal btsets", err)
			return
		}
		tdb.Set("Settings", "BitTorr", buf)
	}
}

func loadBTSets() {
	buf := tdb.Get("Settings", "BitTorr")
	if len(buf) > 0 {
		err := json.Unmarshal(buf, &BTsets)
		if err == nil {
			if BTsets.ReaderReadAHead < 5 {
				BTsets.ReaderReadAHead = 5
			}
			// Set default TMDB settings if missing (for existing configs)
			if BTsets.TMDBSettings.APIURL == "" {
				BTsets.TMDBSettings = TMDBConfig{
					APIKey:     "",
					APIURL:     "https://api.themoviedb.org",
					ImageURL:   "https://image.tmdb.org",
					ImageURLRu: "https://imagetmdb.com",
				}
			}
			// apply environment overrides (envs take precedence over stored config)
			applyEnvOverrides(BTsets)
			return
		}
		log.TLogln("Error unmarshal btsets", err)
	}
	// initialize defaults on error
	SetDefaultConfig()
}

// parse boolean-like env values
func parseBoolEnv(v string) (bool, bool) {
	if v == "" {
		return false, false
	}
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	}
	return false, false
}

// applyEnvOverrides reads env vars prefixed with TORRSERVER_ and overrides fields
func applyEnvOverrides(sets *BTSets) {
	if sets == nil {
		return
	}

	if v := os.Getenv("TS_BTSETS_CACHESIZE"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			sets.CacheSize = i
		}
	}
	if v := os.Getenv("TS_BTSETS_READER_READ_AHEAD"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sets.ReaderReadAHead = i
		}
	}
	if v := os.Getenv("TS_BTSETS_PRELOAD_CACHE"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sets.PreloadCache = i
		}
	}

	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_USE_DISK")); ok {
		sets.UseDisk = v
	}
	if v := os.Getenv("TS_BTSETS_TORRENTS_SAVE_PATH"); v != "" {
		sets.TorrentsSavePath = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_REMOVE_CACHE_ON_DROP")); ok {
		sets.RemoveCacheOnDrop = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_FORCE_ENCRYPT")); ok {
		sets.ForceEncrypt = v
	}
	if v := os.Getenv("TS_BTSETS_RETRACKERS_MODE"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sets.RetrackersMode = i
		}
	}
	if v := os.Getenv("TS_BTSETS_TORRENT_DISCONNECT_TIMEOUT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sets.TorrentDisconnectTimeout = i
		}
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_ENABLE_DEBUG")); ok {
		sets.EnableDebug = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_ENABLE_DLNA")); ok {
		sets.EnableDLNA = v
	}
	if v := os.Getenv("TS_BTSETS_FRIENDLY_NAME"); v != "" {
		sets.FriendlyName = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_ENABLE_RUTOR_SEARCH")); ok {
		sets.EnableRutorSearch = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_ENABLE_TORZNAB_SEARCH")); ok {
		sets.EnableTorznabSearch = v
	}
	if v := os.Getenv("TS_BTSETS_TORZNAB_URLS"); v != "" {
		// try JSON first
		var urls []TorznabConfig
		if err := json.Unmarshal([]byte(v), &urls); err == nil {
			sets.TorznabUrls = urls
		} else {
			// fallback: semicolon separated host|key|name entries
			parts := strings.Split(v, ";")
			var list []TorznabConfig
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				fields := strings.Split(p, "|")
				if len(fields) >= 2 {
					cfg := TorznabConfig{Host: strings.TrimSpace(fields[0]), Key: strings.TrimSpace(fields[1])}
					if len(fields) >= 3 {
						cfg.Name = strings.TrimSpace(fields[2])
					}
					list = append(list, cfg)
				}
			}
			if len(list) > 0 {
				sets.TorznabUrls = list
			}
		}
	}

	// TMDB
	if v := os.Getenv("TS_BTSETS_TMDB_APIKEY"); v != "" {
		sets.TMDBSettings.APIKey = v
	}
	if v := os.Getenv("TS_BTSETS_TMDB_APIURL"); v != "" {
		sets.TMDBSettings.APIURL = v
	}
	if v := os.Getenv("TS_BTSETS_TMDB_IMAGEURL"); v != "" {
		sets.TMDBSettings.ImageURL = v
	}
	if v := os.Getenv("TS_BTSETS_TMDB_IMAGEURL_RU"); v != "" {
		sets.TMDBSettings.ImageURLRu = v
	}
	// BT config
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_ENABLE_IPV6")); ok {
		sets.EnableIPv6 = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_DISABLE_TCP")); ok {
		sets.DisableTCP = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_DISABLE_UTP")); ok {
		sets.DisableUTP = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_DISABLE_UPNP")); ok {
		sets.DisableUPNP = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_DISABLE_DHT")); ok {
		sets.DisableDHT = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_DISABLE_PEX")); ok {
		sets.DisablePEX = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_DISABLE_UPLOAD")); ok {
		sets.DisableUpload = v
	}
	if v := os.Getenv("TS_BTSETS_DOWNLOAD_RATE_LIMIT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sets.DownloadRateLimit = i
		}
	}
	if v := os.Getenv("TS_BTSETS_UPLOAD_RATE_LIMIT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sets.UploadRateLimit = i
		}
	}
	if v := os.Getenv("TS_BTSETS_CONNECTIONS_LIMIT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sets.ConnectionsLimit = i
		}
	}
	if v := os.Getenv("TS_BTSETS_PEERS_LISTEN_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sets.PeersListenPort = i
		}
	}
	// HTTPS
	if v := os.Getenv("TS_BTSETS_SSL_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sets.SslPort = i
		}
	}
	if v := os.Getenv("TS_BTSETS_SSL_CERT"); v != "" {
		sets.SslCert = v
	}
	if v := os.Getenv("TS_BTSETS_SSL_KEY"); v != "" {
		sets.SslKey = v
	}
	// Reader
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_RESPONSIVE_MODE")); ok {
		sets.ResponsiveMode = v
	}
	// FS
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_SHOW_FS_ACTIVE_TORR")); ok {
		sets.ShowFSActiveTorr = v
	}
	// Storage preferences
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_STORE_SETTINGS_IN_JSON")); ok {
		sets.StoreSettingsInJson = v
	}
	if v, ok := parseBoolEnv(os.Getenv("TS_BTSETS_STORE_VIEWED_IN_JSON")); ok {
		sets.StoreViewedInJson = v
	}
}
