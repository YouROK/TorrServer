package torrfs

import (
	"io"
	"sort"
	"time"

	"silo/internal/user"
)

type CategoryNode struct {
	fs      *TorrFS
	user    *user.User
	name    string
	path    string
	entries []rootEntry
}

func (c *CategoryNode) Name() string                     { return c.name }
func (c *CategoryNode) Path() string                     { return c.path }
func (c *CategoryNode) IsDir() bool                      { return true }
func (c *CategoryNode) Size() int64                      { return 0 }
func (c *CategoryNode) ModTime() time.Time               { return time.Time{} }
func (c *CategoryNode) MimeType() string                 { return "" }
func (c *CategoryNode) Open() (io.ReadSeekCloser, error) { return nil, ErrIsDir }

func (c *CategoryNode) Children() ([]Node, error) {
	entries := append([]rootEntry(nil), c.entries...)
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
			fs:   c.fs,
			user: c.user,
			ut:   e.ut,
			rec:  e.rec,
			name: name,
			path: joinPath(c.path, name),
		})
	}
	return nodes, nil
}
