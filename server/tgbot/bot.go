package tgbot

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"
	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"

	"server/log"
	"server/tgbot/config"
	up "server/tgbot/upload"
)

func newTelegramHTTPClient() *http.Client {
	const timeout = 5 * time.Minute
	trimmed := strings.TrimSpace(config.Cfg.Socks5)
	if trimmed == "" {
		return &http.Client{Timeout: timeout}
	}
	raw := trimmed
	if !strings.Contains(raw, "://") {
		raw = "socks5://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		log.TLogln("tg cfg Socks5 parse err, using direct", err)
		return &http.Client{Timeout: timeout}
	}
	if u.Scheme != "socks5" {
		log.TLogln("tg cfg Socks5: only socks5 is supported, got", u.Scheme)
		return &http.Client{Timeout: timeout}
	}
	proxyHost := u.Host
	if proxyHost == "" {
		log.TLogln("tg cfg Socks5: empty host, using direct")
		return &http.Client{Timeout: timeout}
	}
	var auth *proxy.Auth
	if u.User != nil {
		pw, _ := u.User.Password()
		auth = &proxy.Auth{User: u.User.Username(), Password: pw}
	}
	socksDial, err := proxy.SOCKS5("tcp", proxyHost, auth, proxy.Direct)
	if err != nil {
		log.TLogln("tg socks5 dialer err, using direct", err)
		return &http.Client{Timeout: timeout}
	}
	log.TLogln("tg using SOCKS5 proxy", proxyHost)
	transport := &http.Transport{
		Proxy: nil, // respect explicit socks only, not HTTP_PROXY, for this client
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			_ = ctx
			return socksDial.Dial(network, address)
		},
	}
	return &http.Client{Transport: transport, Timeout: timeout}
}

func Start(token string) error {
	config.LoadConfig()
	loadUserLangs()

	pref := tele.Settings{
		URL:       config.Cfg.HostTG,
		Token:     token,
		Poller:    &tele.LongPoller{Timeout: 5 * time.Minute},
		ParseMode: tele.ModeHTML,
		Client:    newTelegramHTTPClient(),
	}

	log.TLogln("tg bot starting")

	b, err := tele.NewBot(pref)
	if err != nil {
		log.TLogln("tg bot start err", err)
		return err
	}

	up.TrFunc = tr
	up.EscapeFunc = escapeHtml

	if err := setBotCommands(b); err != nil {
		log.TLogln("tg setcmd err", err)
	}

	setupMenuButton(b)

	if len(config.Cfg.WhiteIds) > 0 {
		b.Use(middleware.Whitelist(config.Cfg.WhiteIds...))
	}
	if len(config.Cfg.BlackIds) > 0 {
		b.Use(middleware.Blacklist(config.Cfg.BlackIds...))
	}

	b.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Sender() == nil {
				return nil
			}
			if c.Message() != nil && c.Message().Text != "" {
				cmd := logSafeStr(c.Message().Text, 60)
				log.TLogln("tg cmd", logUser(c.Sender()), cmd)
			}
			err := next(c)
			if err != nil {
				log.TLogln("tg cmd err", logUser(c.Sender()), err)
			}
			return err
		}
	})

	b.Handle("help", help)
	b.Handle("Help", help)
	b.Handle("/help", help)
	b.Handle("/Help", help)
	b.Handle("/start", cmdStart)
	b.Handle("/id", help)
	b.Handle("/cancel", cmdCancel)
	b.Handle("/more", sendMoreHub)

	b.Handle("/list", list)
	b.Handle("/clear", clear)
	b.Handle("/add", cmdAdd)
	b.Handle("/remove", cmdRemove)
	b.Handle("/drop", cmdDrop)
	b.Handle("/status", cmdStatus)
	b.Handle("/server", cmdServer)
	b.Handle("/link", cmdLink)
	b.Handle("/play", cmdLink)
	b.Handle("/cache", cmdCache)
	b.Handle("/m3u", cmdM3u)
	b.Handle("/m3uall", cmdM3uAll)
	b.Handle("/search", cmdSearch)
	b.Handle("/rutor", cmdSearchRutor)
	b.Handle("/torznab", cmdTorznab)
	b.Handle("/preload", cmdPreload)
	b.Handle("/queue", up.ShowQueue)
	b.Handle("/set", cmdSet)
	b.Handle("/hash", cmdHash)
	b.Handle("/export", cmdExport)
	b.Handle("/import", cmdImport)
	b.Handle("/categories", cmdCategories)
	b.Handle("/echo", cmdEcho)
	b.Handle("/db", cmdDb)
	b.Handle("/viewed", cmdViewed)
	b.Handle("/ffp", cmdFfp)
	b.Handle("/speedtest", cmdSpeedtest)
	b.Handle("/shutdown", adminOnly(cmdShutdown))
	b.Handle("/settings", adminOnly(cmdSettings))
	b.Handle("/preset", adminOnly(cmdPreset))
	b.Handle("/lang", cmdLang)
	b.Handle("/stats", cmdStats)
	b.Handle("/stat", cmdStat)
	b.Handle("/snake", cmdSnake)

	b.Handle(tele.OnDocument, func(c tele.Context) error {
		if c.Message() == nil {
			return nil
		}
		doc := c.Message().Document
		if doc == nil {
			return nil
		}
		lowerName := strings.ToLower(doc.FileName)
		isTorrent := strings.HasSuffix(lowerName, ".torrent") ||
			strings.Contains(strings.ToLower(doc.MIME), "bittorrent")
		if isTorrent {
			err := addTorrentFromDocument(c, doc)
			if err != nil {
				return err
			}
			return sendListHub(c, 0, false)
		}
		return nil
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		txt := c.Text()
		if handleSettingsInputReply(c) {
			return nil
		}
		uid := c.Sender().ID
		if isMenuButton(uid, txt) {
			return handleMenuButton(c, txt)
		}
		lower := strings.ToLower(txt)
		isLink := strings.HasPrefix(lower, "magnet:") || strings.HasPrefix(lower, "torrs://") ||
			strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") ||
			isHash(txt)
		if !isLink && takePendingSearch(uid) {
			return runSearchQuery(c, strings.TrimSpace(txt))
		}
		if isLink {
			clearPendingSearch(uid)
			err := addTorrent(c, txt)
			if err != nil {
				return err
			}
			return sendListHub(c, 0, false)
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.ReplyMarkup != nil && len(c.Message().ReplyTo.ReplyMarkup.InlineKeyboard) > 0 {
			var hash string
			for _, row := range c.Message().ReplyTo.ReplyMarkup.InlineKeyboard {
				for _, btn := range row {
					if btn.Data == "" {
						continue
					}
					if idx := strings.Index(btn.Data, "all|"); idx >= 0 {
						h := btn.Data[idx+4:]
						if len(h) >= 40 && isHash(h[:40]) {
							hash = h[:40]
						} else if isHash(h) {
							hash = h
						}
					} else if isHash(btn.Data) {
						hash = btn.Data
					}
					if hash != "" {
						break
					}
				}
				if hash != "" {
					break
				}
			}
			if hash != "" {
				from, to, err := ParseRange(c.Sender().ID, c.Message().Text)
				if err != nil {
					_ = c.Send(tr(c.Sender().ID, "range_error"))
					return err
				}
				up.AddRange(c, hash, from, to)
			}
			return nil
		} else {
			return sendWithMenu(c, tr(c.Sender().ID, "add_magnet"))
		}
	})

	b.Handle(tele.OnQuery, handleInlineQuery)

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		args := c.Args()
		if len(args) > 0 {
			cbInfo := strings.TrimPrefix(args[0], "\f")
			if len(args) >= 2 {
				cbInfo += " " + args[1]
			}
			cbInfo = logSafeStr(cbInfo, 80)
			log.TLogln("tg cb", logUser(c.Sender()), cbInfo)
		}
		err := handleCallback(c)
		if err != nil && len(args) > 0 {
			log.TLogln("tg cb err", logUser(c.Sender()), logSafeStr(args[0], 40), err)
		}
		return err
	})

	up.Start()

	go b.Start()
	return nil
}

func setBotCommands(b *tele.Bot) error {
	makeCmds := func(lang string) []tele.Command {
		return []tele.Command{
			{Text: "start", Description: trLang(lang, "cmd_desc_start")},
			{Text: "help", Description: trLang(lang, "cmd_desc_help")},
			{Text: "list", Description: trLang(lang, "cmd_desc_list")},
			{Text: "add", Description: trLang(lang, "cmd_desc_add")},
			{Text: "search", Description: trLang(lang, "cmd_desc_search")},
			{Text: "more", Description: trLang(lang, "cmd_desc_more")},
			{Text: "cancel", Description: trLang(lang, "cmd_desc_cancel")},
			{Text: "lang", Description: trLang(lang, "cmd_desc_lang")},
			{Text: "settings", Description: trLang(lang, "cmd_desc_settings")},
			{Text: "preset", Description: trLang(lang, "cmd_desc_preset")},
			{Text: "shutdown", Description: trLang(lang, "cmd_desc_shutdown")},
		}
	}
	if err := b.SetCommands(makeCmds(LangEN)); err != nil {
		return err
	}
	for _, lang := range []string{LangEN, LangRU} {
		if err := b.SetCommands(makeCmds(lang), lang); err != nil {
			return err
		}
	}
	return nil
}

func help(c tele.Context) error {
	uid := c.Sender().ID
	id := strconv.FormatInt(uid, 10)
	msg := "🤖 <b>" + tr(uid, "help") + "</b>\n\n"
	msg += tr(uid, "help_short") + "\n\n"

	msg += "<b>" + tr(uid, "help_menu_section") + "</b>\n"
	msg += "• " + tr(uid, "menu_library") + " — /list\n"
	msg += "• " + tr(uid, "menu_search") + " — /search\n"
	msg += "• " + tr(uid, "menu_status") + " — /stat\n"
	msg += "• " + tr(uid, "menu_add") + " — /add\n"
	msg += "• " + tr(uid, "menu_more") + " — /more\n"
	msg += "• /cancel — " + tr(uid, "help_cancel") + "\n\n"

	msg += "<b>" + tr(uid, "help_all_commands") + "</b>\n"
	msg += "<b>" + tr(uid, "help_main") + "</b>\n"
	msg += "• /help, /start, /id — " + tr(uid, "help_help") + "\n"
	msg += "• " + tr(uid, "help_list") + "\n"
	msg += "• " + tr(uid, "help_add") + "\n"
	msg += "• " + tr(uid, "help_clear") + "\n"
	msg += "• " + tr(uid, "help_hash") + "\n"
	msg += "• /more — " + tr(uid, "cmd_desc_more") + "\n"
	msg += "• /cancel — " + tr(uid, "help_cancel") + "\n"
	msg += "• " + tr(uid, "help_lang") + "\n\n"

	msg += "<b>" + tr(uid, "help_manage") + "</b> " + tr(uid, "help_manage_desc") + "\n"
	msg += "• " + tr(uid, "help_remove") + "\n"
	msg += "• " + tr(uid, "help_use_index") + "\n"
	msg += "• " + tr(uid, "help_reply") + "\n\n"

	msg += "<b>" + tr(uid, "help_status") + "</b>\n"
	msg += "• " + tr(uid, "help_links") + "\n"
	msg += "• " + tr(uid, "help_m3uall") + "\n"
	msg += "• " + tr(uid, "help_stat") + "\n"
	msg += "• " + tr(uid, "help_stats") + "\n"
	msg += "• " + tr(uid, "help_server_cmd") + "\n\n"

	msg += "<b>" + tr(uid, "help_search") + "</b> " + tr(uid, "help_search_desc") + "\n"
	msg += "• " + tr(uid, "help_search_cmd") + "\n\n"

	msg += "<b>" + tr(uid, "help_other") + "</b>\n"
	msg += "• " + tr(uid, "help_export") + "\n"
	msg += "• " + tr(uid, "help_import") + "\n"
	msg += "• " + tr(uid, "help_categories") + "\n"
	msg += "• " + tr(uid, "help_other_cmd") + "\n"
	msg += "• " + tr(uid, "help_echo") + "\n"
	msg += "• " + tr(uid, "help_db") + "\n"

	if isAdmin(uid) {
		msg += "\n<b>" + tr(uid, "help_server") + "</b>\n"
		msg += "• " + tr(uid, "help_admin") + "\n"
	}

	msg += "\n👤 " + tr(uid, "help_id") + ": <code>" + id + "</code>"
	return sendWithMenu(c, msg)
}

func isHash(txt string) bool {
	if len(txt) == 40 {
		for _, c := range strings.ToLower(txt) {
			switch c {
			case 'a', 'b', 'c', 'd', 'e', 'f', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			default:
				return false
			}
		}
		return true
	}
	return false
}

func ParseRange(userID int64, rng string) (int, int, error) {
	parts := strings.Split(rng, "-")

	if len(parts) != 2 {
		return -1, -1, errors.New(tr(userID, "parse_range_err"))
	}

	num1, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err1 != nil {
		return -1, -1, err1
	}

	num2, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err2 != nil {
		return -1, -1, err2
	}
	if num1 < 1 || num2 < 1 || num1 > num2 {
		return -1, -1, errors.New(tr(userID, "parse_range_err"))
	}
	return num1, num2, nil
}
