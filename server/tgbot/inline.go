package tgbot

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/rutor"
	"server/rutor/models"
	sets "server/settings"
	"server/torr"
	"server/torznab"
)

const inlineMaxResults = 20

func handleInlineQuery(c tele.Context) error {
	query := strings.TrimSpace(c.Query().Text)
	uid := int64(0)
	if c.Query().Sender != nil {
		uid = c.Query().Sender.ID
	}

	var results tele.Results
	id := 0

	if query == "" || strings.ToLower(query) == "list" || strings.ToLower(query) == "play" {
		torrents := torr.ListTorrent()
		host := getHost()
		for _, t := range torrents {
			if id >= inlineMaxResults {
				break
			}
			hash := t.Hash().HexString()
			url := fmt.Sprintf("%s/play/%s/1", host, hash)
			title := t.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			results = append(results, &tele.ArticleResult{
				ResultBase:  tele.ResultBase{ID: strconv.Itoa(id)},
				Title:       "▶ " + title,
				Description: hash[:8] + "...",
				URL:         url,
				Text:        url,
			})
			id++
		}
	}

	if len(query) >= 2 && sets.BTsets != nil && (sets.BTsets.EnableRutorSearch || sets.BTsets.EnableTorznabSearch) {
		var list []*models.TorrentDetails
		if sets.BTsets.EnableRutorSearch {
			list = append(list, rutor.Search(query)...)
		}
		if sets.BTsets.EnableTorznabSearch {
			list = append(list, torznab.Search(query, -1)...)
		}
		for _, item := range list {
			if id >= inlineMaxResults {
				break
			}
			link := item.Magnet
			if link == "" {
				link = item.Link
			}
			if link == "" {
				continue
			}
			title := item.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			size := item.Size
			if size == "" {
				size = "?"
			}
			results = append(results, &tele.ArticleResult{
				ResultBase:  tele.ResultBase{ID: strconv.Itoa(id)},
				Title:       "➕ " + title,
				Description: fmt.Sprintf("%s S:%d P:%d", size, item.Seed, item.Peer),
				Text:        link,
			})
			id++
		}
	}

	if len(results) == 0 {
		results = append(results, &tele.ArticleResult{
			ResultBase:  tele.ResultBase{ID: "0"},
			Title:       tr(uid, "no_torrents"),
			Description: tr(uid, "add_magnet"),
			Text:        "",
		})
	}

	return c.Answer(&tele.QueryResponse{
		Results:    results,
		CacheTime:  60,
		IsPersonal: true,
	})
}
