package tgbot

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/rutor"
	"server/rutor/models"
	sets "server/settings"
	"server/torznab"
)

func cmdSearch(c tele.Context) error {
	if sets.BTsets == nil || (!sets.BTsets.EnableRutorSearch && !sets.BTsets.EnableTorznabSearch) {
		return c.Send(tr(c.Sender().ID, "search_disabled_rutor"))
	}

	args := c.Args()
	if len(args) == 0 {
		return c.Send(tr(c.Sender().ID, "search_usage"))
	}
	query := strings.Join(args, " ")
	uid := c.Sender().ID
	statusMsg, err := c.Bot().Send(c.Sender(), tr(uid, "searching"))
	if err != nil {
		return err
	}
	go func() {
		var list []*models.TorrentDetails
		if sets.BTsets != nil && sets.BTsets.EnableRutorSearch {
			list = append(list, rutor.Search(query)...)
		}
		if sets.BTsets != nil && sets.BTsets.EnableTorznabSearch {
			list = append(list, torznab.Search(query, -1)...)
		}
		source := "RuTor+Torznab"
		sendSearchResultsAsync(c.Bot(), c.Sender(), statusMsg, uid, query, list, source)
	}()
	return nil
}

func cmdSearchRutor(c tele.Context) error {
	if sets.BTsets == nil || !sets.BTsets.EnableRutorSearch {
		return c.Send(tr(c.Sender().ID, "search_disabled_rutor"))
	}

	args := c.Args()
	if len(args) == 0 {
		return c.Send(tr(c.Sender().ID, "rutor_usage"))
	}
	query := strings.Join(args, " ")
	uid := c.Sender().ID
	statusMsg, err := c.Bot().Send(c.Sender(), tr(uid, "searching"))
	if err != nil {
		return err
	}
	go func() {
		list := rutor.Search(query)
		sendSearchResultsAsync(c.Bot(), c.Sender(), statusMsg, uid, query, list, "RuTor")
	}()
	return nil
}

func cmdTorznab(c tele.Context) error {
	if sets.BTsets == nil || !sets.BTsets.EnableTorznabSearch {
		return c.Send(tr(c.Sender().ID, "search_disabled_torznab"))
	}

	args := c.Args()
	if len(args) == 0 {
		return c.Send(tr(c.Sender().ID, "torznab_usage"))
	}
	query := strings.Join(args, " ")
	index := -1
	if len(args) > 1 {
		if i, err := strconv.Atoi(args[len(args)-1]); err == nil && i >= 0 && i < 100 {
			index = i
			query = strings.Join(args[:len(args)-1], " ")
		}
	}
	uid := c.Sender().ID
	statusMsg, err := c.Bot().Send(c.Sender(), tr(uid, "searching"))
	if err != nil {
		return err
	}
	go func() {
		list := torznab.Search(query, index)
		sendSearchResultsAsync(c.Bot(), c.Sender(), statusMsg, uid, query, list, "Torznab")
	}()
	return nil
}

func sendSearchResultsAsync(api tele.API, recipient tele.Recipient, statusMsg *tele.Message, userID int64, query string, list []*models.TorrentDetails, source string) {
	if len(list) == 0 {
		_, _ = api.Edit(statusMsg, fmt.Sprintf(tr(userID, "search_not_found"), query, source))
		return
	}
	_ = api.Delete(statusMsg)
	_ = sendSearchResultsToRecipient(api, recipient, userID, 0, list, source)
}

func sendSearchResultsToRecipient(api tele.API, recipient tele.Recipient, userID int64, offset int, list []*models.TorrentDetails, source string) error {
	const pageSize = 10
	if offset == 0 {
		storeSearchResults(userID, list)
	}
	start := offset
	end := offset + pageSize
	if end > len(list) {
		end = len(list)
	}
	page := list[start:end]

	for i, item := range page {
		idx := offset + i
		link := item.Magnet
		if link == "" {
			link = item.Link
		}
		if link == "" {
			continue
		}
		size := item.Size
		if size == "" {
			size = "?"
		}
		txt := fmt.Sprintf("%d. <b>%s</b> (%s) S:%d P:%d", idx+1, escapeHtml(item.Title), size, item.Seed, item.Peer)
		btnAdd := tele.InlineButton{Text: tr(userID, "btn_add"), Unique: "fadd", Data: strconv.Itoa(idx)}
		kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnAdd}}}
		_, _ = api.Send(recipient, txt, kbd)
	}

	if end < len(list) {
		btnMore := tele.InlineButton{Text: "🔍 " + tr(userID, "search_more"), Unique: "fmore", Data: strconv.Itoa(end)}
		kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnMore}}}
		_, _ = api.Send(recipient, fmt.Sprintf(tr(userID, "search_more_hint"), end, len(list)), kbd)
	}
	return nil
}

func callbackSearchMore(c tele.Context, offsetStr string) error {
	uid := c.Sender().ID
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "error")})
	}
	slice, total := getSearchResultsSlice(uid, offset, 10)
	if len(slice) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "search_expired")})
	}
	_ = c.Respond(&tele.CallbackResponse{})
	if c.Callback().Message != nil {
		_ = c.Bot().Delete(c.Callback().Message)
	}
	return sendSearchResultsPage(c.Bot(), c.Sender(), uid, offset, slice, total)
}

func sendSearchResultsPage(api tele.API, recipient tele.Recipient, userID int64, offset int, page []*models.TorrentDetails, total int) error {
	for i, item := range page {
		idx := offset + i
		link := item.Magnet
		if link == "" {
			link = item.Link
		}
		if link == "" {
			continue
		}
		size := item.Size
		if size == "" {
			size = "?"
		}
		txt := fmt.Sprintf("%d. <b>%s</b> (%s) S:%d P:%d", idx+1, escapeHtml(item.Title), size, item.Seed, item.Peer)
		btnAdd := tele.InlineButton{Text: tr(userID, "btn_add"), Unique: "fadd", Data: strconv.Itoa(idx)}
		kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnAdd}}}
		_, _ = api.Send(recipient, txt, kbd)
	}
	nextOffset := offset + len(page)
	if nextOffset < total {
		btnMore := tele.InlineButton{Text: "🔍 " + tr(userID, "search_more"), Unique: "fmore", Data: strconv.Itoa(nextOffset)}
		kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnMore}}}
		_, _ = api.Send(recipient, fmt.Sprintf(tr(userID, "search_more_hint"), nextOffset, total), kbd)
	}
	return nil
}
