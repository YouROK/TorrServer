package torrstor

// Config содержит параметры работы кэша и хранилища
type Config struct {
	Capacity          int64  `json:"capacity"`             // Размер кэша в байтах
	NextAheadMB       int64  `json:"next_ahead_mb"`        // Ближняя зона загрузки после куска ридера
	ReadaheadMB       int64  `json:"readahead_mb"`         // Дальняя зона загрузки после ближней
	ReserveMB         int64  `json:"reserve_mb"`           // Запас вне зоны, 0 - посчитать автоматически
	UseDisk           bool   `json:"use_disk"`             // Использовать диск вместо RAM
	TorrentsSavePath  string `json:"torrents_save_path"`   // Путь для сохранения кусков на диске
	RemoveCacheOnDrop bool   `json:"remove_cache_on_drop"` // Удалять файлы с диска при закрытии торрента
	ConnectionsLimit  int    `json:"connections_limit"`    // Лимит соединений клиента
}

func DefaultConfig() *Config {
	return &Config{
		Capacity:          64 * 1024 * 1024, // 64 MB
		NextAheadMB:       4,
		ReadaheadMB:       16,
		ReserveMB:         0,
		UseDisk:           false,
		TorrentsSavePath:  "torrents",
		RemoveCacheOnDrop: true,
		ConnectionsLimit:  32,
	}
}

// ReserveBytes возвращает запас буфера, который остается вне зоны скачивания.
// Запас покрывает куски в полете и недокачанный край зоны, поэтому не растет вместе с кэшем.
func (c *Config) ReserveBytes() int64 {
	if c.ReserveMB > 0 {
		return c.ReserveMB << 20
	}

	reserve := c.Capacity / 4
	if reserve > 16<<20 {
		reserve = 16 << 20
	}
	return reserve
}

// ZoneBytes возвращает полную зону скачивания торрента, которая делится между ридерами.
// capacity задается вызывающим, потому что кэш может иметь собственный размер.
func (c *Config) ZoneBytes(capacity int64) int64 {
	if capacity <= 0 {
		capacity = c.Capacity
	}
	zone := capacity - c.ReserveBytes()
	if zone <= 0 {
		return capacity
	}
	return zone
}

// NextBytes возвращает ближнюю зону загрузки в байтах.
func (c *Config) NextBytes() int64 {
	if c.NextAheadMB <= 0 {
		return 0
	}
	return c.NextAheadMB << 20
}

// ReadaheadBytes возвращает дальнюю зону загрузки в байтах.
func (c *Config) ReadaheadBytes() int64 {
	if c.ReadaheadMB <= 0 {
		return 0
	}
	return c.ReadaheadMB << 20
}

// Normalize приводит конфиг к рабочим значениям после загрузки из базы.
// Старые конфиги не знают про зоны загрузки и приходят с нулями, из-за чего
// движок не получает ярусы приоритетов и не знает границ зоны.
func (c *Config) Normalize() {
	def := DefaultConfig()

	if c.Capacity <= 0 {
		c.Capacity = def.Capacity
	}
	if c.NextAheadMB <= 0 {
		c.NextAheadMB = def.NextAheadMB
	}
	if c.ReadaheadMB <= 0 {
		c.ReadaheadMB = def.ReadaheadMB
	}
	if c.ReserveMB < 0 {
		c.ReserveMB = 0
	}
	if c.ConnectionsLimit <= 0 {
		c.ConnectionsLimit = def.ConnectionsLimit
	}
	if c.UseDisk && c.TorrentsSavePath == "" {
		c.UseDisk = false
	}
}
