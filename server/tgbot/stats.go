package tgbot

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func cmdStats(c tele.Context) error {
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}

	var totalSize, loadedSize int64
	var totalPeers, activePeers, seeders int
	for _, t := range torrents {
		st := t.Status()
		if st != nil {
			totalSize += st.TorrentSize
			loadedSize += st.LoadedSize
			totalPeers += st.TotalPeers
			activePeers += st.ActivePeers
			seeders += st.ConnectedSeeders
		} else {
			totalSize += t.Size
		}
	}

	streams := torr.GetActiveStreams()

	uid := c.Sender().ID
	var sb strings.Builder
	sb.WriteString("📊 <b>" + tr(uid, "stats_title") + "</b>\n\n")
	fmt.Fprintf(&sb, "%s: %d\n", tr(uid, "stats_torrents"), len(torrents))
	fmt.Fprintf(&sb, "%s: %s\n", tr(uid, "stats_total_size"), humanize.IBytes(uint64(totalSize)))
	progress := 0.0
	if totalSize > 0 {
		progress = float64(loadedSize) / float64(totalSize) * 100
	}
	fmt.Fprintf(&sb, "%s: %s (%.1f%%)\n",
		tr(uid, "stats_loaded"), humanize.IBytes(uint64(loadedSize)), progress)
	fmt.Fprintf(&sb, "%s: %d %s, %d %s\n",
		tr(uid, "stats_peers"), activePeers, tr(uid, "stats_active"), seeders, tr(uid, "stats_seeds"))
	fmt.Fprintf(&sb, "%s: %d\n", tr(uid, "stats_streams"), streams)
	return c.Send(sb.String())
}
