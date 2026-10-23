package torr

// CacheRow is a cheap per-torrent cache snapshot (atomics only, no GetState).
type CacheRow struct {
	Hash     string `json:"hash"`
	Filled   int64  `json:"filled"`
	Capacity int64  `json:"capacity"`
	Pieces   int    `json:"pieces"`
	Closed   bool   `json:"closed"`
}

// SnapshotCaches returns per-torrent cache rows and aggregates.
func SnapshotCaches() (rows []CacheRow, filled, capacity int64) {
	list := ListTorrent()
	rows = make([]CacheRow, 0, len(list))
	for _, t := range list {
		if t == nil {
			continue
		}
		c := t.GetCache()
		if c == nil {
			continue
		}
		f := c.FilledBytes()
		capn := c.GetCapacity()
		rows = append(rows, CacheRow{
			Hash:     t.Hash().HexString(),
			Filled:   f,
			Capacity: capn,
			Pieces:   c.PieceCount(),
			Closed:   c.IsClosed(),
		})
		filled += f
		capacity += capn
	}
	return rows, filled, capacity
}
