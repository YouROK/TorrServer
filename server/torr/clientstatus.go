package torr

import (
	"bytes"
	"fmt"
)

// ClientStatusSnapshot is a structured view of what /stat dumps as plaintext.
type ClientStatusSnapshot struct {
	ListenPort    int                  `json:"listen_port"`
	PeerID        string               `json:"peer_id"`
	BannedIPs     int                  `json:"banned_ips"`
	ActiveStreams int32                `json:"active_streams"`
	TorrentCount  int                  `json:"torrent_count"`
	TotalSize     int64                `json:"total_size"`
	LoadedSize    int64                `json:"loaded_size"`
	ActivePeers   int                  `json:"active_peers"`
	TotalPeers    int                  `json:"total_peers"`
	Seeders       int                  `json:"connected_seeders"`
	BytesRead     int64                `json:"bytes_read"`
	BytesWritten  int64                `json:"bytes_written"`
	DownloadSpeed float64              `json:"download_speed"`
	UploadSpeed   float64              `json:"upload_speed"`
	Torrents      []TorrentStatusBrief `json:"torrents"`
	RawStat       string               `json:"raw_stat"`
}

// TorrentStatusBrief is a compact per-torrent row for the Server Status UI.
type TorrentStatusBrief struct {
	Hash             string  `json:"hash"`
	Title            string  `json:"title"`
	Name             string  `json:"name,omitempty"`
	Stat             int     `json:"stat"`
	StatString       string  `json:"stat_string"`
	TorrentSize      int64   `json:"torrent_size"`
	LoadedSize       int64   `json:"loaded_size"`
	DownloadSpeed    float64 `json:"download_speed"`
	UploadSpeed      float64 `json:"upload_speed"`
	ActivePeers      int     `json:"active_peers"`
	TotalPeers       int     `json:"total_peers"`
	ConnectedSeeders int     `json:"connected_seeders"`
	PreloadedBytes   int64   `json:"preloaded_bytes"`
	PreloadSize      int64   `json:"preload_size"`
}

// SnapshotClientStatus builds JSON-friendly stats for the web UI (+ full /stat text).
func SnapshotClientStatus() ClientStatusSnapshot {
	snap := ClientStatusSnapshot{
		Torrents:      make([]TorrentStatusBrief, 0),
		ActiveStreams: GetActiveStreams(),
	}

	if bts != nil && bts.client != nil {
		snap.ListenPort = bts.client.LocalPort()
		snap.PeerID = fmt.Sprintf("%+q", bts.client.PeerID())
		snap.BannedIPs = len(bts.client.BadPeerIPs())
	}

	var buf bytes.Buffer
	WriteStatus(&buf)
	raw := buf.String()
	const maxRaw = 200_000
	if len(raw) > maxRaw {
		raw = raw[:maxRaw] + "\n…\n"
	}
	snap.RawStat = raw

	for _, t := range ListTorrent() {
		st := t.Status()
		if st == nil {
			continue
		}
		title := st.Title
		if title == "" {
			title = st.Name
		}
		brief := TorrentStatusBrief{
			Hash:             st.Hash,
			Title:            title,
			Name:             st.Name,
			Stat:             int(st.Stat),
			StatString:       st.StatString,
			TorrentSize:      st.TorrentSize,
			LoadedSize:       st.LoadedSize,
			DownloadSpeed:    st.DownloadSpeed,
			UploadSpeed:      st.UploadSpeed,
			ActivePeers:      st.ActivePeers,
			TotalPeers:       st.TotalPeers,
			ConnectedSeeders: st.ConnectedSeeders,
			PreloadedBytes:   st.PreloadedBytes,
			PreloadSize:      st.PreloadSize,
		}
		snap.Torrents = append(snap.Torrents, brief)
		snap.TotalSize += st.TorrentSize
		snap.LoadedSize += st.LoadedSize
		snap.ActivePeers += st.ActivePeers
		snap.TotalPeers += st.TotalPeers
		snap.Seeders += st.ConnectedSeeders
		snap.BytesRead += st.BytesRead
		snap.BytesWritten += st.BytesWritten
		snap.DownloadSpeed += st.DownloadSpeed
		snap.UploadSpeed += st.UploadSpeed
	}
	snap.TorrentCount = len(snap.Torrents)
	return snap
}
