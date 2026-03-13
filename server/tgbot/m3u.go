package tgbot

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func callbackM3u(c tele.Context, hash string) error {
	uid := c.Sender().ID
	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "torrent_not_found")})
	}
	host := getHost()
	url := fmt.Sprintf("%s/playlist?hash=%s", host, hash)
	_ = c.Respond(&tele.CallbackResponse{})
	return c.Send(fmt.Sprintf(tr(uid, "m3u_playlist"), url))
}

func cmdM3u(c tele.Context) error {
	args := c.Args()
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	hash := resolveHash(c, arg)
	if hash == "" {
		return c.Send(tr(c.Sender().ID, "m3u_usage"))
	}

	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Send(tr(c.Sender().ID, "torrent_not_found") + ":\n<code>" + hash + "</code>")
	}

	host := getHost()
	url := fmt.Sprintf("%s/playlist?hash=%s", host, hash)
	if len(args) > 1 && strings.ToLower(args[1]) == "fromlast" {
		url += "&fromlast=1"
	}
	return c.Send(fmt.Sprintf(tr(c.Sender().ID, "m3u_playlist"), url))
}

func cmdM3uAll(c tele.Context) error {
	host := getHost()
	url := host + "/playlistall/all.m3u"
	return c.Send(fmt.Sprintf(tr(c.Sender().ID, "m3u_all"), url))
}
