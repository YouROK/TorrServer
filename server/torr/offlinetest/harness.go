// Offline anacrolix/torrent client and synthetic multi-file metainfo for tests.
// No DHT, trackers, or public network: a loopback seeder holds the bytes.
package offlinetest

import (
	"bytes"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"

	"server/ffprobe"
	"server/settings"
)

// Prepare points settings at a temp DB and turns the BitTorrent client off the network.
func Prepare(t testing.TB) {
	t.Helper()
	dir := t.TempDir()
	restoreFF := ffprobe.SetBinPath(filepath.Join(dir, "no-ffprobe"))
	oldPath, oldSets, oldArgs := settings.Path, settings.BTsets, settings.Args
	oldRO, oldAuth, oldWA := settings.ReadOnly, settings.HttpAuth, settings.SearchWA
	oldIP4, oldIP6, oldTor := settings.PubIPv4, settings.PubIPv6, settings.TorAddr
	oldEmb, oldStop := settings.Embedded, settings.EmbeddedStop
	t.Cleanup(func() {
		restoreFF()
		settings.CloseDB()
		settings.Path, settings.BTsets, settings.Args = oldPath, oldSets, oldArgs
		settings.ReadOnly, settings.HttpAuth, settings.SearchWA = oldRO, oldAuth, oldWA
		settings.PubIPv4, settings.PubIPv6, settings.TorAddr = oldIP4, oldIP6, oldTor
		settings.Embedded, settings.EmbeddedStop = oldEmb, oldStop
	})

	settings.Path = dir
	if err := settings.InitSets(false, false); err != nil {
		t.Fatal(err)
	}
	if settings.BTsets == nil {
		settings.BTsets = &settings.BTSets{}
	}
	s := settings.BTsets
	s.CacheSize = 32 << 20
	s.PreloadCache = 100
	s.ReaderReadAHead = 50
	s.ConnectionsLimit = 25
	s.RetrackersMode = 0
	s.TorrentDisconnectTimeout = 120
	s.ResponsiveMode = true
	s.EnableIPv6 = false
	s.DisableDHT = true
	s.DisablePEX = true
	s.DisableUPNP = true
	s.DisableUTP = true
	s.DisableTCP = false
	s.EnableLPD = false
	s.EnableDLNA = false
	s.EnableBonjour = false
	s.EnableRutorSearch = false
	s.EnableTorznabSearch = false
	s.EnableDebug = false
	s.UseDisk = false
	s.DownloadRateLimit = 0
	s.UploadRateLimit = 0
	s.PeersListenPort = 0
	s.ForceEncrypt = false
	settings.PubIPv4 = "8.8.8.8"
	settings.PubIPv6 = ""
	settings.TorAddr = "127.0.0.1:0"
	settings.ReadOnly = false
	settings.HttpAuth = false
	settings.SearchWA = false
	settings.Embedded = false
	if settings.Args == nil {
		settings.Args = &settings.ExecArgs{}
	}
	settings.Args.ProxyURL = ""
	settings.Args.ProxyMode = ""
	settings.Args.Pprof = false
	settings.Args.RDB = false
	settings.MaxSize = 0
}

// Show is a multi-file torrent whose bytes already sit in memory.
type Show struct {
	Meta *metainfo.MetaInfo
	Spec *torrent.TorrentSpec
	Data []byte
}

// WriteSized builds a single-file torrent of length bytes. length must be a multiple of pieceLen.
func WriteSized(t testing.TB, name string, length, pieceLen int64) *Show {
	t.Helper()
	if length <= 0 || pieceLen <= 0 || length%pieceLen != 0 {
		t.Fatalf("bad torrent size %d / %d", length, pieceLen)
	}
	data := bytes.Repeat([]byte("M"), int(length))
	info := metainfo.Info{Name: name, Length: length, PieceLength: pieceLen}
	if err := info.GeneratePieces(func(metainfo.FileInfo) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}); err != nil {
		t.Fatal(err)
	}
	mi := &metainfo.MetaInfo{InfoBytes: bencode.MustMarshal(info)}
	return &Show{Meta: mi, Spec: torrent.TorrentSpecFromMetaInfo(mi), Data: data}
}

// WriteShow builds Season 1/episode1.mkv, a subtitle, and a poster with real piece hashes.
func WriteShow(t testing.TB) *Show {
	t.Helper()
	root := filepath.Join(t.TempDir(), "Show")
	files := map[string][]byte{
		"Season 1/episode1.mkv":    bytes.Repeat([]byte("M"), 64*1024),
		"Season 1/episode1.srt":    []byte("1\n00:00:00,000 --> 00:00:01,000\nHi\n"),
		"Season 1/Show.S01E02.mkv": bytes.Repeat([]byte("E"), 1024),
		"poster.jpg":               {0xff, 0xd8, 0xff, 0xd9, 0x00, 0x01},
	}
	for rel, data := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	info := metainfo.Info{PieceLength: 16 * 1024}
	if err := info.BuildFromFilePath(root); err != nil {
		t.Fatal(err)
	}
	var data bytes.Buffer
	for _, fi := range info.UpvertedFiles() {
		b, err := os.ReadFile(filepath.Join(root, filepath.Join(fi.Path...)))
		if err != nil {
			t.Fatal(err)
		}
		data.Write(b)
	}
	mi := &metainfo.MetaInfo{InfoBytes: bencode.MustMarshal(info)}
	return &Show{Meta: mi, Spec: torrent.TorrentSpecFromMetaInfo(mi), Data: data.Bytes()}
}

// Seeder is a loopback client that already has every piece.
type Seeder struct {
	Client  *torrent.Client
	Torrent *torrent.Torrent
}

// Seed starts a client bound to 127.0.0.1:0 and waits until the show is complete.
func Seed(t testing.TB, show *Show) *Seeder {
	t.Helper()
	cfg := torrent.NewDefaultClientConfig()
	cfg.DefaultStorage = memStore{data: show.Data}
	cfg.Seed = true
	cfg.NoDHT = true
	cfg.DisableTrackers = true
	cfg.DisablePEX = true
	cfg.DisableIPv6 = true
	cfg.DisableUTP = true
	cfg.NoDefaultPortForwarding = true
	cfg.ListenPort = 0
	cfg.ListenHost = func(string) string { return "127.0.0.1" }
	cl, err := torrent.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cl.Close() })
	tor, err := cl.AddTorrent(show.Meta)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-tor.GotInfo():
	case <-time.After(10 * time.Second):
		t.Fatal("seeder did not get info")
	}
	tor.DownloadAll()
	deadline := time.Now().Add(15 * time.Second)
	for tor.BytesCompleted() < tor.Length() {
		if time.Now().After(deadline) {
			t.Fatalf("seeder incomplete %d/%d", tor.BytesCompleted(), tor.Length())
		}
		time.Sleep(20 * time.Millisecond)
	}
	return &Seeder{Client: cl, Torrent: tor}
}

// memStore reports every piece complete and serves the torrent bytes.
// The fork dropped the file storage backend, so the seeder cannot use a directory.
type memStore struct{ data []byte }

func (m memStore) OpenTorrent(*metainfo.Info, metainfo.Hash) (storage.TorrentImpl, error) {
	return m, nil
}
func (m memStore) Close() error { return nil }
func (m memStore) Piece(p metainfo.Piece) storage.PieceImpl {
	return memPiece{data: m.data, p: p}
}

type memPiece struct {
	data []byte
	p    metainfo.Piece
}

func (p memPiece) ReadAt(b []byte, off int64) (int, error) {
	start := p.p.Offset() + off
	if start >= int64(len(p.data)) {
		return 0, io.EOF
	}
	n := copy(b, p.data[start:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}
func (p memPiece) WriteAt(b []byte, _ int64) (int, error) { return len(b), nil }
func (p memPiece) MarkComplete() error                    { return nil }
func (p memPiece) MarkNotComplete() error                 { return nil }
func (p memPiece) Completion() storage.Completion {
	return storage.Completion{Complete: true, Ok: true}
}

// Peer is the seeder's TCP address for Torrent.AddPeers.
func (s *Seeder) Peer() torrent.Peer {
	for _, a := range s.Client.ListenAddrs() {
		ta, ok := a.(*net.TCPAddr)
		if !ok || ta == nil || ta.Port == 0 {
			continue
		}
		ip := ta.IP
		if v4 := ip.To4(); v4 != nil {
			ip = v4
		}
		return torrent.Peer{IP: ip, Port: ta.Port}
	}
	return torrent.Peer{}
}
