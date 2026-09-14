package api

import (
	"strings"
	"testing"

	"server/torr"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

func listedTorrent(hex, title string, infoBytes []byte, data string) *torr.Torrent {
	return &torr.Torrent{
		Title: title,
		Data:  data,
		TorrentSpec: &torrent.TorrentSpec{
			InfoHash:  metainfo.NewHashFromHex(hex),
			InfoBytes: infoBytes,
		},
	}
}

func mustInfoBytes(t *testing.T, name string, files []metainfo.FileInfo) []byte {
	t.Helper()
	buf, err := bencode.Marshal(&metainfo.Info{
		Name:        name,
		Files:       files,
		PieceLength: 256 * 1024,
		Pieces:      make([]byte, 20),
	})
	if err != nil {
		t.Fatal(err)
	}
	return buf
}

func TestStatusForMergedM3U(t *testing.T) {
	hash := strings.Repeat("a", 40)
	dataJSON := `{"TorrServer":{"Files":[{"id":3,"path":"from-data.mkv","length":20}]}}`
	specBytes := mustInfoBytes(t, "from-spec", []metainfo.FileInfo{
		{Length: 100, Path: []string{"from-spec.mp4"}},
	})

	t.Run("InfoBytes wins over Data", func(t *testing.T) {
		st := statusForMergedM3U(listedTorrent(hash, "Title", specBytes, dataJSON))
		if st == nil || len(st.FileStats) != 1 {
			t.Fatalf("got %#v", st)
		}
		if st.FileStats[0].Path != "from-spec.mp4" || st.FileStats[0].Id != 1 {
			t.Fatalf("spec files: %#v", st.FileStats[0])
		}
	})

	t.Run("Data files when InfoBytes empty", func(t *testing.T) {
		st := statusForMergedM3U(listedTorrent(hash, "Title", nil, dataJSON))
		if st == nil || len(st.FileStats) != 1 {
			t.Fatalf("got %#v", st)
		}
		if st.FileStats[0].Path != "from-data.mkv" || st.FileStats[0].Id != 3 {
			t.Fatalf("data files: %#v", st.FileStats[0])
		}
		if st.Hash != hash {
			t.Fatalf("hash %q", st.Hash)
		}
	})

	t.Run("neither", func(t *testing.T) {
		tr := listedTorrent(hash, "Idle", nil, "")
		if st := statusForMergedM3U(tr); st != nil {
			t.Fatalf("got %#v", st)
		}
		entry := nestedPlaylistEntry(tr, "http://h")
		if !strings.Contains(entry, `type="playlist"`) || !strings.Contains(entry, hash) {
			t.Fatalf("nested entry: %q", entry)
		}
	})

	t.Run("junk Data", func(t *testing.T) {
		tr := listedTorrent(hash, "Idle", nil, "{not json")
		if st := statusForMergedM3U(tr); st != nil {
			t.Fatalf("got %#v", st)
		}
		entry := nestedPlaylistEntry(tr, "http://h")
		if !strings.Contains(entry, `type="playlist"`) || !strings.Contains(entry, hash) {
			t.Fatalf("nested entry: %q", entry)
		}
	})
}
