package tgbot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"server/log"
	"server/rutor/models"
	sets "server/settings"
	"server/torr"

	tele "gopkg.in/telebot.v4"
)

const inlinePageSize = 20

// inlineSwitchQuery is passed to switch_inline_query_current_chat.
// A single space opens the panel more reliably on iOS than an empty string;
// handleInlineQuery trims it to a library listing.
const inlineSwitchQuery = " "

type inlinePick struct {
	private bool
	link    string
	expires time.Time
}

type inlineSearchCache struct {
	query   string
	results []*models.TorrentDetails
	expires time.Time
}

var (
	inlinePickMu   sync.Mutex
	inlinePicks    = make(map[string]*inlinePick)
	inlineSearchMu sync.Mutex
	inlineSearches = make(map[int64]*inlineSearchCache)
)

func inlinePickKey(uid int64, resultID string) string {
	return strconv.FormatInt(uid, 10) + ":" + resultID
}

func storeInlinePick(uid int64, resultID, link string, private bool) {
	inlinePickMu.Lock()
	defer inlinePickMu.Unlock()
	if len(inlinePicks) > 2000 {
		now := time.Now()
		for k, v := range inlinePicks {
			if v == nil || now.After(v.expires) {
				delete(inlinePicks, k)
			}
		}
	}
	inlinePicks[inlinePickKey(uid, resultID)] = &inlinePick{
		private: private,
		link:    link,
		expires: time.Now().Add(10 * time.Minute),
	}
}

func takeInlinePick(uid int64, resultID string) *inlinePick {
	key := inlinePickKey(uid, resultID)
	inlinePickMu.Lock()
	defer inlinePickMu.Unlock()
	p := inlinePicks[key]
	if p == nil || time.Now().After(p.expires) {
		delete(inlinePicks, key)
		return nil
	}
	return p
}

// stripBotInlineQuery reports whether text is "@bot" or "@bot query" for this bot.
func stripBotInlineQuery(text string) (query string, ok bool) {
	uname := strings.TrimSpace(botUsername)
	if uname == "" {
		return "", false
	}
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return "", false
	}
	if !strings.EqualFold(fields[0], "@"+uname) {
		return "", false
	}
	return strings.Join(fields[1:], " "), true
}

func handleInlineQuery(c tele.Context) error {
	q := c.Query()
	if q == nil {
		return nil
	}
	query := strings.TrimSpace(q.Text)
	uid := int64(0)
	if q.Sender != nil {
		uid = q.Sender.ID
	}
	offset := 0
	if q.Offset != "" {
		offset, _ = strconv.Atoi(q.Offset)
		if offset < 0 {
			offset = 0
		}
	}
	private := q.ChatType == "sender"
	lower := strings.ToLower(query)

	var results tele.Results
	next := ""
	kind := "library"

	switch {
	case query == "" || lower == "list" || lower == "play":
		results, next = inlineLibrary(offset)
	default:
		if sets.BTsets == nil || (!sets.BTsets.EnableRutorSearch && !sets.BTsets.EnableTorznabSearch) {
			kind = "disabled"
		} else {
			kind = "search"
			results, next = inlineSearch(uid, query, offset, private)
		}
	}

	n := len(results)
	log.TLogln("tg inline", logUser(q.Sender), "q="+logSafeStr(query, 60), "chat="+q.ChatType, "n="+strconv.Itoa(n), kind)

	resp := &tele.QueryResponse{
		Results:    results,
		CacheTime:  15,
		IsPersonal: true,
		NextOffset: next,
	}
	if n == 0 {
		title, desc, text := inlineEmptyCopy(uid, query, kind)
		resp.Results = tele.Results{&tele.ArticleResult{
			ResultBase:  tele.ResultBase{ID: "empty"},
			Title:       title,
			Description: desc,
			Text:        text,
		}}
		if kind == "library" {
			resp.SwitchPMText = tr(uid, "inline_open_bot")
			resp.SwitchPMParameter = "add"
		}
	}
	if err := c.Answer(resp); err != nil {
		log.TLogln("tg inline answer err", logUser(q.Sender), err)
		return err
	}
	return nil
}

func inlineEmptyCopy(uid int64, query, kind string) (title, desc, text string) {
	switch kind {
	case "disabled":
		msg := tr(uid, "inline_empty_disabled")
		return msg, msg, msg
	case "search":
		msg := fmt.Sprintf(tr(uid, "inline_empty_search"), query)
		return msg, msg, msg
	default:
		return tr(uid, "no_torrents"), tr(uid, "add_magnet"), tr(uid, "add_magnet")
	}
}

func inlineLibrary(offset int) (tele.Results, string) {
	torrents := torr.ListTorrent()
	if offset >= len(torrents) {
		return nil, ""
	}
	end := offset + inlinePageSize
	if end > len(torrents) {
		end = len(torrents)
	}
	var results tele.Results
	for i, t := range torrents[offset:end] {
		hash := t.Hash().HexString()
		title, poster, _, _ := t.Meta()
		if title == "" {
			title = hash[:8]
		}
		if len([]rune(title)) > 60 {
			title = string([]rune(title)[:57]) + "..."
		}
		id := fmt.Sprintf("l%d", offset+i)
		item := inlineLibraryArticle(id, hash, title, poster)
		results = append(results, item)
	}
	next := ""
	if end < len(torrents) {
		next = strconv.Itoa(end)
	}
	return results, next
}

func inlineLibraryArticle(id, hash, title, poster string) *tele.ArticleResult {
	startURL := torrentStartURL(hash)
	text := title
	if startURL != "" {
		text = title + "\n" + startURL
	}
	desc := hash
	if len(hash) >= 8 {
		desc = hash[:8] + "…"
	}
	item := &tele.ArticleResult{
		ResultBase:  tele.ResultBase{ID: id},
		Title:       "▶ " + title,
		Description: desc,
		Text:        text,
		URL:         startURL,
	}
	if isInlineThumbURL(poster) {
		item.ThumbURL = poster
	}
	return item
}

func inlineSearch(uid int64, query string, offset int, private bool) (tele.Results, string) {
	if sets.BTsets == nil || (!sets.BTsets.EnableRutorSearch && !sets.BTsets.EnableTorznabSearch) {
		return nil, ""
	}
	list := loadInlineSearch(uid, query)
	if offset >= len(list) {
		return nil, ""
	}
	end := offset + inlinePageSize
	if end > len(list) {
		end = len(list)
	}
	var results tele.Results
	for i, item := range list[offset:end] {
		if item == nil {
			continue
		}
		link := item.Magnet
		if link == "" {
			link = item.Link
		}
		if link == "" {
			continue
		}
		title := item.Title
		if len([]rune(title)) > 60 {
			title = string([]rune(title)[:57]) + "..."
		}
		size := item.Size
		if size == "" {
			size = "?"
		}
		id := fmt.Sprintf("s%d", offset+i)
		text := link
		if private {
			text = fmt.Sprintf(tr(uid, "inline_add_private"), title)
		}
		art := &tele.ArticleResult{
			ResultBase:  tele.ResultBase{ID: id},
			Title:       "➕ " + title,
			Description: fmt.Sprintf("%s S:%d P:%d", size, item.Seed, item.Peer),
			Text:        text,
		}
		storeInlinePick(uid, id, link, private)
		results = append(results, art)
	}
	next := ""
	if end < len(list) {
		next = strconv.Itoa(end)
	}
	return results, next
}

func loadInlineSearch(uid int64, query string) []*models.TorrentDetails {
	inlineSearchMu.Lock()
	cached := inlineSearches[uid]
	if cached != nil && cached.query == query && time.Now().Before(cached.expires) {
		list := cached.results
		inlineSearchMu.Unlock()
		return list
	}
	inlineSearchMu.Unlock()

	var list []*models.TorrentDetails
	if sets.BTsets.EnableRutorSearch {
		list = append(list, rutorSearch(query)...)
	}
	if sets.BTsets.EnableTorznabSearch {
		list = append(list, torznabSearch(context.Background(), query, -1, "", 0, 0)...)
	}

	inlineSearchMu.Lock()
	inlineSearches[uid] = &inlineSearchCache{
		query:   query,
		results: list,
		expires: time.Now().Add(2 * time.Minute),
	}
	inlineSearchMu.Unlock()
	return list
}

func handleInlineChosen(c tele.Context) error {
	ir := c.InlineResult()
	if ir == nil || ir.Sender == nil {
		return nil
	}
	pick := takeInlinePick(ir.Sender.ID, ir.ResultID)
	if pick == nil || !pick.private || pick.link == "" {
		return nil
	}
	if !strings.HasPrefix(strings.ToLower(pick.link), "magnet:") &&
		!strings.HasPrefix(strings.ToLower(pick.link), "torrs://") &&
		!isHash(pick.link) &&
		!strings.HasPrefix(strings.ToLower(pick.link), "http") {
		return nil
	}
	notifyTyping(c)
	tor, err := addTorrent(c, pick.link)
	if err != nil {
		return err
	}
	return afterAdd(c, tor)
}
