package tgbot

import "server/torr"

type torrentMeta struct {
	Title    string
	Poster   string
	Category string
	Data     string
}

func mergeTorrentMeta(cur torrentMeta, title, poster, category *string) torrentMeta {
	if title != nil {
		cur.Title = *title
	}
	if poster != nil {
		cur.Poster = *poster
	}
	if category != nil {
		cur.Category = *category
	}
	return cur
}

// torrentTitle reads a torrent's title under its lock; a bare t.Title races with
// a rename or the add handler filling it.
func torrentTitle(t *torr.Torrent) string {
	title, _, _, _ := t.Meta()
	return title
}

func currentTorrentMeta(hash string) (torrentMeta, bool) {
	t := torr.GetTorrent(hash)
	if t == nil {
		return torrentMeta{}, false
	}
	st := t.Status()
	if st == nil {
		title, poster, category, data := t.Meta()
		return torrentMeta{Title: title, Poster: poster, Category: category, Data: data}, true
	}
	return torrentMeta{Title: st.Title, Poster: st.Poster, Category: st.Category, Data: st.Data}, true
}

func applyTorrentMeta(hash string, title, poster, category *string) *torr.Torrent {
	cur, ok := currentTorrentMeta(hash)
	if !ok {
		return nil
	}
	next := mergeTorrentMeta(cur, title, poster, category)
	return torr.SetTorrent(hash, next.Title, next.Poster, next.Category, next.Data)
}
