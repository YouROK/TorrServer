package torrent

import "testing"

// TestMergeTorrentRecords проверяет дополнение записи и объединение трекеров.
func TestMergeTorrentRecords(t *testing.T) {
	old := &TorrentRecord{
		Hash:      "abc",
		Name:      "Old",
		Size:      10,
		Timestamp: 1,
		MagnetUri: "magnet:old",
		InfoBytes: []byte{1},
		Trackers:  []string{"a", "b"},
		Files:     []*TorrentFileStat{{Path: "old.mkv"}},
	}

	// Пустые поля новой записи не затирают старые
	empty := &TorrentRecord{Hash: "abc", Timestamp: 2, Trackers: []string{"b", "c"}}
	got := MergeTorrentRecords(old, empty)

	if got.Name != "Old" || got.Size != 10 || got.MagnetUri != "magnet:old" {
		t.Errorf("пустые поля затёрли старые: %+v", got)
	}
	if len(got.InfoBytes) != 1 || len(got.Files) != 1 {
		t.Errorf("info bytes или файлы потеряны: %+v", got)
	}
	if got.Timestamp != 2 {
		t.Errorf("timestamp = %d, want 2", got.Timestamp)
	}
	if len(got.Trackers) != 3 || got.Trackers[0] != "a" || got.Trackers[1] != "b" || got.Trackers[2] != "c" {
		t.Errorf("трекеры = %v, want [a b c]", got.Trackers)
	}

	// Непустые поля заменяют старые
	full := &TorrentRecord{
		Hash:      "abc",
		Name:      "New",
		Size:      20,
		MagnetUri: "magnet:new",
		InfoBytes: []byte{2},
		Files:     []*TorrentFileStat{{Path: "new.mkv"}},
	}
	got = MergeTorrentRecords(old, full)

	if got.Name != "New" || got.Size != 20 || got.MagnetUri != "magnet:new" {
		t.Errorf("новые поля не применились: %+v", got)
	}
	if len(got.InfoBytes) != 1 || got.InfoBytes[0] != 2 {
		t.Errorf("info bytes = %v, want [2]", got.InfoBytes)
	}
	if len(got.Files) != 1 || got.Files[0].Path != "new.mkv" {
		t.Errorf("файлы = %+v, want new.mkv", got.Files)
	}
}
