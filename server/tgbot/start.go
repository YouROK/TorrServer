package tgbot

import (
	"fmt"
	"strings"

	"server/library"
	"server/torr"

	tele "gopkg.in/telebot.v4"
)

type startKind int

const (
	startWelcome startKind = iota
	startList
	startNext
	startTorrent
	startAdd
	startSearch
	startUnknown
)

type startAction struct {
	kind   startKind
	cat    string
	prefix string
	query  string
}

func parseStartPayload(raw string) startAction {
	p := strings.TrimSpace(raw)
	if p == "" {
		return startAction{kind: startWelcome}
	}
	lower := strings.ToLower(p)
	switch {
	case lower == "list" || lower == "library":
		return startAction{kind: startList}
	case strings.HasPrefix(lower, "list_"):
		rest := strings.TrimPrefix(lower, "list_")
		cat := library.NormalizeCategory(rest)
		if rest != "" && cat == "" {
			return startAction{kind: startUnknown}
		}
		if cat == "all" {
			cat = ""
		}
		return startAction{kind: startList, cat: cat}
	case lower == "next":
		return startAction{kind: startNext}
	case lower == "add":
		return startAction{kind: startAdd}
	case strings.HasPrefix(lower, "t_"):
		hex := strings.TrimPrefix(lower, "t_")
		if !isHashPrefix(hex) {
			return startAction{kind: startUnknown}
		}
		return startAction{kind: startTorrent, prefix: hex}
	case strings.HasPrefix(lower, "search_"):
		rest := p[len("search_"):]
		q := strings.TrimSpace(strings.ReplaceAll(rest, "_", " "))
		if q == "" {
			return startAction{kind: startAdd}
		}
		return startAction{kind: startSearch, query: q}
	default:
		if isHashPrefix(lower) {
			return startAction{kind: startTorrent, prefix: lower}
		}
		return startAction{kind: startUnknown}
	}
}

func startPayloadOf(c tele.Context) string {
	if m := c.Message(); m != nil && strings.TrimSpace(m.Payload) != "" {
		return strings.TrimSpace(m.Payload)
	}
	return strings.TrimSpace(strings.Join(c.Args(), " "))
}

func cmdStart(c tele.Context) error {
	return handleStart(c, startPayloadOf(c))
}

func handleStart(c tele.Context, payload string) error {
	ensureUserLang(c)
	setUserSlashCommands(c)
	uid := c.Sender().ID
	act := parseStartPayload(payload)
	switch act.kind {
	case startList:
		return sendListHub(c, 0, act.cat, false)
	case startNext:
		return runNextUnwatched(c, "", "tv", "")
	case startAdd:
		return sendAddPrompt(c)
	case startSearch:
		return runSearchQuery(c, act.query)
	case startTorrent:
		matches := torrentsByHashPrefix(act.prefix)
		if len(matches) == 0 {
			_ = sendWithMenu(c, tr(uid, "torrent_not_found"))
			return sendWelcome(c)
		}
		if len(matches) == 1 {
			return showTorrentCard(c, matches[0], "0", false)
		}
		var b strings.Builder
		b.WriteString(tr(uid, "start_multi_hash") + "\n")
		for i, h := range matches {
			if i >= 8 {
				break
			}
			t := torr.GetTorrent(h)
			title := h
			if t != nil && torrentTitle(t) != "" {
				title = torrentTitle(t)
			}
			fmt.Fprintf(&b, "• <code>%s</code> %s\n", h, escapeHtml(title))
		}
		return sendWithMenu(c, b.String(), tele.ModeHTML)
	case startUnknown:
		_ = sendWithMenu(c, tr(uid, "start_unknown"))
		return sendWelcome(c)
	default:
		return sendWelcome(c)
	}
}

func sendWelcome(c tele.Context) error {
	uid := c.Sender().ID
	msg := "🤖 <b>" + tr(uid, "help") + "</b>\n\n" + tr(uid, "menu_welcome")
	if u := botUsername; u != "" {
		msg += "\n\n" + fmt.Sprintf(tr(uid, "help_deeplink"), "https://t.me/"+u+"?start=list")
	}
	return sendWithMenu(c, msg, tele.ModeHTML)
}
