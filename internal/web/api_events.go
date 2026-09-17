package web

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// LibraryCard — краткие данные карточки для главного экрана темы
type LibraryCard struct {
	Hash          string  `json:"hash"`
	Title         string  `json:"title"`
	Poster        string  `json:"poster"`
	Category      string  `json:"category"`
	Stat          int     `json:"stat"`
	Size          int64   `json:"size"`
	Peers         int     `json:"peers"`
	Seeders       int     `json:"seeders"`
	DownloadSpeed float64 `json:"download_speed"`
	Torrs         string  `json:"torrs"`
}

func (s *Server) buildLibraryCards(u *user.User) ([]LibraryCard, error) {
	list, err := s.torrentMgr.ListTorrents(u)
	if err != nil {
		return nil, err
	}

	cards := make([]LibraryCard, 0, len(list))
	for _, st := range list {
		cards = append(cards, LibraryCard{
			Hash:          st.Hash,
			Title:         st.Title,
			Poster:        st.Poster,
			Category:      st.Category,
			Stat:          int(st.Stat),
			Size:          st.TorrentSize,
			Peers:         st.ActivePeers,
			Seeders:       st.ConnectedSeeders,
			DownloadSpeed: st.DownloadSpeed,
			Torrs:         st.Torrs,
		})
	}
	return cards, nil
}

// handleEventsStream — SSE-поток обновлений библиотеки.
func (s *Server) handleEventsStream(c *gin.Context) {
	val, _ := c.Get("user")
	u := val.(*user.User)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	prev := make(map[string]string)
	first := true

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case <-ticker.C:
		}

		cards, err := s.buildLibraryCards(u)
		if err != nil {
			return false
		}

		cur := make(map[string]string, len(cards))
		for _, card := range cards {
			data, err := json.Marshal(card)
			if err != nil {
				continue
			}
			cur[card.Hash] = string(data)
		}

		if first {
			all, _ := json.Marshal(cards)
			fmt.Fprintf(c.Writer, "event: sync\ndata: %s\n\n", all)
			c.Writer.Flush()
			first = false
		} else {
			for hash, data := range cur {
				if old, ok := prev[hash]; !ok || old != data {
					fmt.Fprintf(c.Writer, "event: card\ndata: %s\n\n", data)
					c.Writer.Flush()
				}
			}
			for hash := range prev {
				if _, ok := cur[hash]; !ok {
					fmt.Fprintf(c.Writer, "event: remove\ndata: {\"hash\":\"%s\"}\n\n", hash)
					c.Writer.Flush()
				}
			}
		}

		prev = cur
		return true
	})
}
