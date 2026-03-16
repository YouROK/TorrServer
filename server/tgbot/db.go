package tgbot

import (
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	"server/log"
	sets "server/settings"
)

const dbPageSize = 10

func cmdDb(c tele.Context) error {
	return sendDbPage(c, 0)
}

func sendDbPage(c tele.Context, page int) error {
	uid := c.Sender().ID
	dbList := sets.ListTorrent()
	if len(dbList) == 0 {
		return c.Send(tr(uid, "db_empty"))
	}

	totalPages := (len(dbList) + dbPageSize - 1) / dbPageSize
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	start := page * dbPageSize
	end := start + dbPageSize
	if end > len(dbList) {
		end = len(dbList)
	}
	pageList := dbList[start:end]

	var sb strings.Builder
	sb.WriteString("📁 <b>" + tr(uid, "db_title") + "</b> (" + strconv.Itoa(len(dbList)) + ")\n\n")
	for i, t := range pageList {
		hash := t.InfoHash.HexString()
		sb.WriteString(strconv.Itoa(start+i+1) + ". <b>" + escapeHtml(t.Title) + "</b>")
		if t.Size > 0 {
			sb.WriteString(" <i>" + humanize.IBytes(uint64(t.Size)) + "</i>")
		}
		sb.WriteString("\n<code>" + hash + "</code>\n\n")
	}
	msg := strings.TrimSuffix(sb.String(), "\n\n")

	navRow := []tele.InlineButton{}
	if totalPages > 1 {
		if page > 0 {
			navRow = append(navRow, tele.InlineButton{Text: "◀️", Unique: "fdb", Data: strconv.Itoa(page - 1)})
		}
		navRow = append(navRow, tele.InlineButton{Text: strconv.Itoa(page+1) + "/" + strconv.Itoa(totalPages), Unique: "fnop", Data: ""})
		if page < totalPages-1 {
			navRow = append(navRow, tele.InlineButton{Text: "▶️", Unique: "fdb", Data: strconv.Itoa(page + 1)})
		}
	}
	navRow = append(navRow, tele.InlineButton{Text: "🔄", Unique: "fdbrefresh", Data: strconv.Itoa(page)})

	kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{navRow}}
	if err := c.Send(msg, kbd); err != nil {
		log.TLogln("tg db send err", err)
		return err
	}
	return nil
}

func callbackDbPage(c tele.Context, data string) error {
	page := 0
	if data != "" {
		if p, err := strconv.Atoi(data); err == nil {
			page = p
		}
	}
	_ = c.Respond(&tele.CallbackResponse{})
	if c.Callback().Message != nil {
		_ = c.Bot().Delete(c.Callback().Message)
	}
	return sendDbPage(c, page)
}

func callbackDbRefresh(c tele.Context, data string) error {
	page := 0
	if data != "" {
		if p, err := strconv.Atoi(data); err == nil {
			page = p
		}
	}
	_ = c.Respond(&tele.CallbackResponse{Text: "🔄"})
	if c.Callback().Message != nil {
		_ = c.Bot().Delete(c.Callback().Message)
	}
	return sendDbPage(c, page)
}
