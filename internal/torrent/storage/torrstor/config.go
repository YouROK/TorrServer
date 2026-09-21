package torrstor

// Config содержит параметры работы кэша и хранилища
type Config struct {
	Capacity          int64  `json:"capacity"`             // Размер кэша в байтах
	UseDisk           bool   `json:"use_disk"`             // Использовать диск вместо RAM
	TorrentsSavePath  string `json:"torrents_save_path"`   // Путь для сохранения кусков на диске
	RemoveCacheOnDrop bool   `json:"remove_cache_on_drop"` // Удалять файлы с диска при закрытии торрента
	ConnectionsLimit  int    `json:"connections_limit"`    // Лимит соединений (для расчета приоритетов кусков)
	ReaderReadAHead   int    `json:"reader_read_ahead"`    // Процент упреждающего чтения (по умолчанию 95)
}

func DefaultConfig() *Config {
	return &Config{
		Capacity:          64 * 1024 * 1024, // 64 MB
		UseDisk:           false,
		TorrentsSavePath:  "torrents",
		RemoveCacheOnDrop: true,
		ConnectionsLimit:  32,
		ReaderReadAHead:   95,
	}
}
