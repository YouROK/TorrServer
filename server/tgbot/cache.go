package tgbot

import (
	"fmt"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func cmdCache(c tele.Context) error {
	arg := ""
	if args := c.Args(); len(args) > 0 {
		arg = args[0]
	}
	hash := resolveHash(c, arg)
	if hash == "" {
		return c.Send(tr(c.Sender().ID, "cache_usage"))
	}

	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Send(tr(c.Sender().ID, "torrent_not_found") + ":\n<code>" + hash + "</code>")
	}

	st := t.CacheState()
	if st == nil {
		return c.Send(fmt.Sprintf(tr(c.Sender().ID, "cache_unavailable"), hash))
	}

	uid := c.Sender().ID
	txt := "💾 <b>" + escapeHtml(st.Torrent.Title) + "</b>\n\n"
	txt += fmt.Sprintf("%s: %s\n", tr(uid, "cache_capacity"), humanize.IBytes(uint64(st.Capacity)))
	txt += fmt.Sprintf("%s: %s\n", tr(uid, "cache_filled"), humanize.IBytes(uint64(st.Filled)))
	txt += fmt.Sprintf("%s: %d\n", tr(uid, "cache_pieces"), st.PiecesCount)
	txt += fmt.Sprintf("%s: %d\n", tr(uid, "cache_readers"), len(st.Readers))
	txt += fmt.Sprintf("<code>%s</code>", hash)
	return c.Send(txt)
}
