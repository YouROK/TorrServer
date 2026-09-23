package torrfs

import (
	"io"
	"sort"
	"time"

	"silo/internal/torrent"
	"silo/internal/user"
)

// rootEntry — торрент юзера + его физическая карточка.
type rootEntry struct {
	ut  *user.UserTorrent
	rec *torrent.TorrentRecord
}

type RootNode struct {
	fs   *TorrFS
	user *user.User
}

func (r *RootNode) Name() string                     { return "" }
func (r *RootNode) Path() string                     { return "" }
func (r *RootNode) IsDir() bool                      { return true }
func (r *RootNode) Size() int64                      { return 0 }
func (r *RootNode) ModTime() time.Time               { return time.Time{} }
func (r *RootNode) MimeType() string                 { return "" }
func (r *RootNode) Open() (io.ReadSeekCloser, error) { return nil, ErrIsDir }

func (r *RootNode) Children() ([]Node, error) {
	entries, err := r.loadEntries()
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}

	cats := map[string]bool{}
	for _, e := range entries {
		cats[effectiveCategory(e.ut.Category)] = true
	}

	if len(cats) >= 2 {
		return r.categoryChildren(entries), nil
	}
	return r.torrentChildren(entries, ""), nil
}

func (r *RootNode) loadEntries() ([]rootEntry, error) {
	uts, err := r.fs.users.ListTorrents(r.user)
	if err != nil {
		return nil, err
	}

	var entries []rootEntry
	for _, ut := range uts {
		rec, err := r.fs.store.Get(ut.TorrentHash)
		if err != nil || rec == nil || len(rec.Files) == 0 {
			continue
		}
		entries = append(entries, rootEntry{ut: ut, rec: rec})
	}
	return entries, nil
}

func (r *RootNode) categoryChildren(entries []rootEntry) []Node {
	byCat := map[string][]rootEntry{}
	for _, e := range entries {
		cat := effectiveCategory(e.ut.Category)
		byCat[cat] = append(byCat[cat], e)
	}

	names := make([]string, 0, len(byCat))
	for name := range byCat {
		names = append(names, name)
	}
	sort.Strings(names)

	nodes := make([]Node, 0, len(names))
	for _, name := range names {
		nodes = append(nodes, &CategoryNode{
			fs:      r.fs,
			user:    r.user,
			name:    name,
			path:    name,
			entries: byCat[name],
		})
	}
	return nodes
}

func (r *RootNode) torrentChildren(entries []rootEntry, prefix string) []Node {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ut.AddedAt.After(entries[j].ut.AddedAt)
	})

	names := make([]string, len(entries))
	counts := map[string]int{}
	for i, e := range entries {
		names[i] = sanitizeName(e.ut.Title)
		counts[names[i]]++
	}

	nodes := make([]Node, 0, len(entries))
	for i, e := range entries {
		name := names[i]
		if counts[name] > 1 {
			name = name + " (" + shortHash(e.ut.TorrentHash) + ")"
		}
		nodes = append(nodes, &TorrNode{
			fs:   r.fs,
			user: r.user,
			ut:   e.ut,
			rec:  e.rec,
			name: name,
			path: joinPath(prefix, name),
		})
	}
	return nodes
}

func effectiveCategory(c string) string {
	s := sanitizeName(c)
	if s == "unnamed" {
		return "other"
	}
	return s
}
