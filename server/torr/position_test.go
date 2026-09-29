package torr

import (
	"testing"

	sets "server/settings"
)

// A player that opens an episode at its start to read the header, on a torrent too slow to hand
// it over quickly, looks for a while like a session that began at zero. Seen on 2026-09-28: the
// 1914s the episode had been left at were replaced by 0s, and the player never got as far as
// seeking to them.
func TestAPositionAtTheStartDoesNotReplaceAFurtherOne(t *testing.T) {
	stored := []*sets.Viewed{
		{Hash: "h", FileIndex: 2, TimeCode: 1180, Duration: 2930},
		{Hash: "h", FileIndex: 3, TimeCode: 1914, Duration: 2680},
	}
	for _, tc := range []struct {
		name     string
		fileID   int
		timecode float64
		want     bool
	}{
		{"заголовок поверх 1914с", 3, 0, false},
		{"первые секунды поверх 1914с", 3, 14, false},
		{"пересмотр с начала, уже сыграно 15с", 3, 15, true},
		{"дальше записанного", 3, 1950, true},
		{"назад, но не к началу — перемотка", 3, 600, true},
		{"новая серия, записи нет", 4, 0, true},
		{"у соседней серии своя запись", 2, 0, false},
	} {
		if got := mayReplace(stored, tc.fileID, tc.timecode); got != tc.want {
			t.Errorf("%s: %v, ждали %v", tc.name, got, tc.want)
		}
	}
}
