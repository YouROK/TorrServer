package state

import "testing"

func TestTorrentStatString(t *testing.T) {
	cases := []struct {
		stat TorrentStat
		want string
	}{
		{TorrentAdded, "Torrent added"},
		{TorrentGettingInfo, "Torrent getting info"},
		{TorrentPreload, "Torrent preload"},
		{TorrentWorking, "Torrent working"},
		{TorrentClosed, "Torrent closed"},
		{TorrentInDB, "Torrent in db"},
		{TorrentStat(99), "Torrent unknown status"},
		{TorrentStat(-1), "Torrent unknown status"},
	}
	for _, tc := range cases {
		if got := tc.stat.String(); got != tc.want {
			t.Errorf("%d.String() = %q, want %q", tc.stat, got, tc.want)
		}
	}
}
