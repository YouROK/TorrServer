package tgbot

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"server/log"
	set "server/settings"
	"server/torr"
	"server/web/api/utils"

	"github.com/anacrolix/torrent"
	tele "gopkg.in/telebot.v4"
)

func addTorrentFromSpec(c tele.Context, torrSpec *torrent.TorrentSpec, displayLabel, title, poster, category string) (*torr.Torrent, error) {
	notifyTyping(c)
	msg, err := c.Bot().Send(c.Sender(), tr(c.Sender().ID, "connecting"))
	if err != nil {
		return nil, err
	}

	tor, err := torr.AddTorrent(torrSpec, title, poster, "", category)
	if err != nil {
		log.TLogln("tg add err", err)
		_, _ = c.Bot().Edit(msg, fmt.Sprintf(tr(c.Sender().ID, "add_error"), err.Error()))
		return nil, err
	}
	if tor == nil {
		_, _ = c.Bot().Edit(msg, tr(c.Sender().ID, "add_not_created"))
		return nil, errors.New("torrent not created")
	}

	if set.BTsets != nil && set.BTsets.EnableDebug {
		_, _, category, data := tor.Meta()
		if data != "" {
			log.TLogln("tg add data", logSafeStr(data, 60))
		}
		if category != "" {
			log.TLogln("tg add category", logSafeStr(category, 40))
		}
	}

	_, _ = c.Bot().Edit(msg, tr(c.Sender().ID, "add_getting_meta"))
	if !tor.GotInfo() {
		log.TLogln("tg add err", "timeout get torrent info")
		_, _ = c.Bot().Edit(msg, tr(c.Sender().ID, "add_timeout"))
		return nil, errors.New("timeout connection get torrent info")
	}

	tor.EnsureTitle(func() string {
		title := torrSpec.DisplayName
		title = strings.ReplaceAll(title, "rutor.info", "")
		title = strings.ReplaceAll(title, "_", " ")
		title = strings.Trim(title, " ")
		if title == "" {
			title = tor.Name()
		}
		return title
	})

	torr.SaveTorrentToDB(tor)

	if len(displayLabel) > 80 {
		displayLabel = displayLabel[:77] + "..."
	}
	_, _ = c.Bot().Edit(msg, fmt.Sprintf(tr(c.Sender().ID, "add_success"), displayLabel))
	return tor, nil
}

func addTorrent(c tele.Context, link string) (*torr.Torrent, error) {
	log.TLogln("tg add torrent", logHashOrTruncate(link))
	link = strings.ReplaceAll(link, "&amp;", "&")
	var torrSpec *torrent.TorrentSpec
	var title, poster, category string
	var err error
	if strings.HasPrefix(strings.ToLower(link), "torrs://") {
		spec, parsed, perr := utils.ParseTorrsHash(link)
		err = perr
		torrSpec = spec
		if parsed != nil {
			title, poster, category = parsed.Title(), parsed.Poster(), parsed.Category()
		}
	} else {
		torrSpec, err = utils.ParseLink(link)
	}
	if err != nil {
		log.TLogln("tg add parse err", err)
		return nil, err
	}
	return addTorrentFromSpec(c, torrSpec, link, title, poster, category)
}

func addTorrentFromDocument(c tele.Context, doc *tele.Document) (*torr.Torrent, error) {
	if doc == nil || doc.FileID == "" {
		return nil, errors.New("no document")
	}
	reader, err := c.Bot().File(&doc.File)
	if err != nil {
		log.TLogln("tg add document getfile err", err)
		return nil, err
	}
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(reader)
	if err != nil {
		log.TLogln("tg add document read err", err)
		return nil, err
	}
	torrSpec, err := utils.ParseFromBytes(data)
	if err != nil {
		log.TLogln("tg add document parse err", err)
		return nil, err
	}
	displayLabel := doc.FileName
	if displayLabel == "" {
		displayLabel = ".torrent"
	}
	return addTorrentFromSpec(c, torrSpec, displayLabel, "", "", "")
}

func afterAdd(c tele.Context, tor *torr.Torrent) error {
	if tor == nil {
		return nil
	}
	hash := tor.Hash().HexString()
	if _, _, category, _ := tor.Meta(); strings.TrimSpace(category) == "" {
		return sendCategoryPicker(c, hash)
	}
	return showTorrentCard(c, hash, "0", false)
}

func cmdAdd(c tele.Context) error {
	uid := c.Sender().ID
	args := c.Args()
	if len(args) == 0 {
		return sendAddPrompt(c)
	}
	link := strings.TrimSpace(strings.Join(args, " "))
	if link == "" {
		return c.Send(tr(uid, "add_no_link"))
	}
	tor, err := addTorrent(c, link)
	if err != nil {
		return err
	}
	return afterAdd(c, tor)
}
