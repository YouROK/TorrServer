package tgbot

import (
	"strings"
	"sync"
	"time"

	tele "gopkg.in/telebot.v4"

	"server/log"
	up "server/tgbot/upload"
)

var (
	pendingSearchMu sync.Mutex
	pendingSearch   = make(map[int64]time.Time)
)

const pendingSearchTTL = 30 * time.Minute

func setPendingSearch(uid int64) {
	pendingSearchMu.Lock()
	pendingSearch[uid] = time.Now()
	pendingSearchMu.Unlock()
}

func takePendingSearch(uid int64) bool {
	pendingSearchMu.Lock()
	defer pendingSearchMu.Unlock()
	t, ok := pendingSearch[uid]
	if !ok {
		return false
	}
	delete(pendingSearch, uid)
	return time.Since(t) <= pendingSearchTTL
}

func clearPendingSearch(uid int64) bool {
	pendingSearchMu.Lock()
	defer pendingSearchMu.Unlock()
	if _, ok := pendingSearch[uid]; ok {
		delete(pendingSearch, uid)
		return true
	}
	return false
}

// Reply-keyboard labels must match exactly what we send (localized).
func mainMenuKeyboard(uid int64) *tele.ReplyMarkup {
	m := &tele.ReplyMarkup{ResizeKeyboard: true}
	m.Reply(
		m.Row(m.Text(tr(uid, "menu_library")), m.Text(tr(uid, "menu_search"))),
		m.Row(m.Text(tr(uid, "menu_status")), m.Text(tr(uid, "menu_add"))),
		m.Row(m.Text(tr(uid, "menu_more"))),
	)
	return m
}

// sendWithMenu sends a plain text message with the reply keyboard.
// Never combine with an inline markup — telebot keeps only the last ReplyMarkup.
func sendWithMenu(c tele.Context, what interface{}, opts ...interface{}) error {
	uid := c.Sender().ID
	all := make([]interface{}, 0, len(opts)+1)
	all = append(all, opts...)
	all = append(all, mainMenuKeyboard(uid))
	return c.Send(what, all...)
}

func isMenuButton(uid int64, text string) bool {
	t := strings.TrimSpace(text)
	switch t {
	case tr(uid, "menu_library"), tr(uid, "menu_search"), tr(uid, "menu_status"),
		tr(uid, "menu_add"), tr(uid, "menu_more"):
		return true
	default:
		return false
	}
}

func handleMenuButton(c tele.Context, text string) error {
	uid := c.Sender().ID
	clearPendingSearch(uid)
	switch strings.TrimSpace(text) {
	case tr(uid, "menu_library"):
		return sendListHub(c, 0, false)
	case tr(uid, "menu_search"):
		setPendingSearch(uid)
		return sendWithMenu(c, tr(uid, "menu_search_pending"))
	case tr(uid, "menu_status"):
		return cmdStat(c)
	case tr(uid, "menu_add"):
		return sendWithMenu(c, tr(uid, "add_magnet"))
	case tr(uid, "menu_more"):
		return sendMoreHub(c)
	default:
		return nil
	}
}

func sendMoreHub(c tele.Context) error {
	return showMoreHub(c, "root", false)
}

func showMoreHub(c tele.Context, section string, edit bool) error {
	uid := c.Sender().ID
	text, kbd := moreHubContent(uid, section)
	if edit && c.Callback() != nil && c.Callback().Message != nil {
		_, err := c.Bot().Edit(c.Callback().Message, text, kbd, tele.ModeHTML)
		if err == nil {
			return nil
		}
		log.TLogln("tg more hub edit err", err)
	}
	if err := c.Send(text, kbd, tele.ModeHTML); err != nil {
		log.TLogln("tg more hub send err", err)
		return err
	}
	return nil
}

func moreHubContent(uid int64, section string) (string, *tele.ReplyMarkup) {
	m := &tele.ReplyMarkup{}
	title := "⋯ <b>" + tr(uid, "menu_more_title") + "</b>"
	var rows []tele.Row

	switch section {
	case "lib":
		title += "\n" + tr(uid, "menu_section_lib")
		rows = []tele.Row{
			m.Row(m.Data(tr(uid, "menu_act_clear"), "fmenu", "act", "clear"),
				m.Data(tr(uid, "menu_act_hash"), "fmenu", "act", "hash")),
			m.Row(m.Data(tr(uid, "menu_act_categories"), "fmenu", "act", "categories")),
			m.Row(m.Data(tr(uid, "menu_export"), "fmenu", "act", "export"),
				m.Data(tr(uid, "menu_import"), "fmenu", "act", "import")),
			m.Row(m.Data(tr(uid, "menu_back"), "fmenu", "root")),
		}
	case "tools":
		title += "\n" + tr(uid, "menu_section_tools")
		rows = []tele.Row{
			m.Row(m.Data(tr(uid, "menu_act_snake"), "fmenu", "act", "snake"),
				m.Data(tr(uid, "menu_act_preload"), "fmenu", "act", "preload")),
			m.Row(m.Data(tr(uid, "menu_act_queue"), "fmenu", "act", "queue"),
				m.Data(tr(uid, "menu_act_ffp"), "fmenu", "act", "ffp")),
			m.Row(m.Data(tr(uid, "menu_act_speedtest"), "fmenu", "act", "speedtest"),
				m.Data(tr(uid, "menu_act_echo"), "fmenu", "act", "echo")),
			m.Row(m.Data(tr(uid, "menu_act_db"), "fmenu", "act", "db"),
				m.Data(tr(uid, "menu_act_viewed"), "fmenu", "act", "viewed")),
			m.Row(m.Data(tr(uid, "menu_act_server"), "fmenu", "act", "server"),
				m.Data(tr(uid, "menu_act_stats"), "fmenu", "act", "stats")),
			m.Row(m.Data(tr(uid, "menu_act_stat"), "fmenu", "act", "stat")),
			m.Row(m.Data(tr(uid, "menu_back"), "fmenu", "root")),
		}
	case "links":
		title += "\n" + tr(uid, "menu_section_links")
		rows = []tele.Row{
			m.Row(m.Data(tr(uid, "menu_act_m3uall"), "fmenu", "act", "m3uall"),
				m.Data(tr(uid, "menu_act_cache"), "fmenu", "act", "cache")),
			m.Row(m.Data(tr(uid, "menu_back"), "fmenu", "root")),
		}
	case "admin":
		title += "\n" + tr(uid, "menu_section_admin")
		rows = []tele.Row{
			m.Row(m.Data(tr(uid, "menu_settings"), "fmenu", "act", "settings")),
			m.Row(m.Data(tr(uid, "menu_act_preset"), "fmenu", "act", "preset"),
				m.Data(tr(uid, "menu_act_shutdown"), "fmenu", "act", "shutdown")),
			m.Row(m.Data(tr(uid, "menu_back"), "fmenu", "root")),
		}
	default:
		rows = []tele.Row{
			m.Row(m.Data(tr(uid, "menu_section_lib"), "fmenu", "hub", "lib"),
				m.Data(tr(uid, "menu_section_tools"), "fmenu", "hub", "tools")),
			m.Row(m.Data(tr(uid, "menu_section_links"), "fmenu", "hub", "links")),
		}
		if isAdmin(uid) {
			rows = append(rows, m.Row(m.Data(tr(uid, "menu_section_admin"), "fmenu", "hub", "admin")))
		}
		rows = append(rows, m.Row(m.Data(tr(uid, "menu_help"), "fmenu", "act", "help")))
		rows = append(rows, openWebButtonRow(m, uid)...)
	}

	m.Inline(rows...)
	return title, m
}

func openWebButtonRow(m *tele.ReplyMarkup, uid int64) []tele.Row {
	host := getHost()
	url := strings.TrimRight(host, "/") + "/?tg=1"
	if isHTTPSURL(host) {
		return []tele.Row{m.Row(m.WebApp(tr(uid, "menu_open_web"), &tele.WebApp{URL: url}))}
	}
	return []tele.Row{m.Row(m.URL(tr(uid, "menu_open_web"), url))}
}

func isHTTPSURL(host string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(host)), "https://")
}

func miniAppURL() string {
	return strings.TrimRight(getHost(), "/") + "/?tg=1"
}

func setupMenuButton(b *tele.Bot) {
	if !isHTTPSURL(getHost()) {
		return
	}
	mb := &tele.MenuButton{
		Type:   tele.MenuButtonWebApp,
		Text:   "TorrServer",
		WebApp: &tele.WebApp{URL: miniAppURL()},
	}
	if err := b.SetMenuButton(nil, mb); err != nil {
		log.TLogln("tg SetMenuButton", err)
	}
}

func cmdStart(c tele.Context) error {
	uid := c.Sender().ID
	msg := "🤖 <b>" + tr(uid, "help") + "</b>\n\n" + tr(uid, "menu_welcome")
	return sendWithMenu(c, msg)
}

func callbackMenu(c tele.Context, parts []string) error {
	uid := c.Sender().ID
	_ = c.Respond(&tele.CallbackResponse{})
	if len(parts) == 0 {
		return showMoreHub(c, "root", true)
	}
	switch parts[0] {
	case "root":
		return showMoreHub(c, "root", true)
	case "hub":
		sec := "root"
		if len(parts) >= 2 {
			sec = parts[1]
		}
		if sec == "admin" && !isAdmin(uid) {
			return sendWithMenu(c, tr(uid, "admin_only"))
		}
		return showMoreHub(c, sec, true)
	case "act":
		if len(parts) < 2 {
			return nil
		}
		return callbackMenuAct(c, parts[1])
	// Legacy single-token actions from older hub
	case "export", "import", "help", "settings":
		return callbackMenuAct(c, parts[0])
	default:
		return nil
	}
}

func callbackMenuAct(c tele.Context, act string) error {
	uid := c.Sender().ID
	switch act {
	case "clear":
		return clear(c)
	case "hash":
		return cmdHash(c)
	case "categories":
		return cmdCategories(c)
	case "export":
		return cmdExport(c)
	case "import":
		return sendWithMenu(c, tr(uid, "menu_import_hint"))
	case "snake":
		return sendWithMenu(c, tr(uid, "snake_usage")+"\n\n"+tr(uid, "menu_pick_torrent"))
	case "preload":
		return sendWithMenu(c, tr(uid, "preload_usage")+"\n\n"+tr(uid, "menu_pick_torrent"))
	case "queue":
		return up.ShowQueue(c)
	case "ffp":
		return sendWithMenu(c, tr(uid, "ffp_usage")+"\n\n"+tr(uid, "menu_pick_torrent"))
	case "speedtest":
		return cmdSpeedtest(c)
	case "echo":
		return cmdEcho(c)
	case "db":
		return cmdDb(c)
	case "viewed":
		return cmdViewed(c)
	case "server":
		return cmdServer(c)
	case "stats":
		return cmdStats(c)
	case "stat":
		return cmdStat(c)
	case "m3uall":
		return cmdM3uAll(c)
	case "cache":
		return sendWithMenu(c, tr(uid, "cache_usage")+"\n\n"+tr(uid, "menu_pick_torrent"))
	case "help":
		return help(c)
	case "settings":
		if !isAdmin(uid) {
			return sendWithMenu(c, tr(uid, "admin_only"))
		}
		return cmdSettings(c)
	case "preset":
		if !isAdmin(uid) {
			return sendWithMenu(c, tr(uid, "admin_only"))
		}
		return sendWithMenu(c, tr(uid, "preset_usage"))
	case "shutdown":
		if !isAdmin(uid) {
			return sendWithMenu(c, tr(uid, "admin_only"))
		}
		return cmdShutdown(c)
	default:
		return nil
	}
}
