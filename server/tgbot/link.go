package tgbot

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func callbackLink(c tele.Context, data string) error {
	uid := c.Sender().ID
	index := 1
	hash := data
	if idx := strings.Index(data, "|"); idx >= 0 && idx+1 < len(data) {
		if i, err := strconv.Atoi(data[idx+1:]); err == nil && i > 0 {
			index = i
			hash = data[:idx]
		}
	}
	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "torrent_not_found")})
	}
	if !strings.Contains(data, "|") && t.WaitInfo() {
		st := t.Status()
		if st != nil && len(st.FileStats) > 1 {
			maxFiles := 5
			if len(st.FileStats) < maxFiles {
				maxFiles = len(st.FileStats)
			}
			var rows [][]tele.InlineButton
			for i := 0; i < maxFiles; i++ {
				f := st.FileStats[i]
				btn := tele.InlineButton{Text: fmt.Sprintf("#%d", f.Id), Unique: "flink", Data: hash + "|" + strconv.Itoa(f.Id)}
				rows = append(rows, []tele.InlineButton{btn})
			}
			kbd := &tele.ReplyMarkup{InlineKeyboard: rows}
			_ = c.Respond(&tele.CallbackResponse{})
			return c.Send("🔗 "+tr(uid, "btn_link")+":", kbd)
		}
	}
	host := getHost()
	url := fmt.Sprintf("%s/play/%s/%d", host, hash, index)
	_ = c.Respond(&tele.CallbackResponse{})
	return c.Send(fmt.Sprintf(tr(uid, "link_play"), url))
}

func cmdLink(c tele.Context) error {
	args := c.Args()
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	hash := resolveHash(c, arg)
	if hash == "" {
		return c.Send(tr(c.Sender().ID, "link_usage"))
	}

	index := 1
	if len(args) > 1 {
		if i, err := strconv.Atoi(args[1]); err == nil && i > 0 {
			index = i
		}
	}

	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Send(tr(c.Sender().ID, "torrent_not_found") + ":\n<code>" + hash + "</code>")
	}

	host := getHost()
	url := fmt.Sprintf("%s/play/%s/%d", host, hash, index)
	return c.Send(fmt.Sprintf(tr(c.Sender().ID, "link_play"), url))
}
