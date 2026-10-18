package mcp

import (
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"server/torr/state"
	"server/utils"
)

var (
	reSxxExx = regexp.MustCompile(`(?i)(?:^|[^0-9A-Za-z])s(\d{1,3})[.\-_ ]*e(\d{1,4})(?:[^0-9A-Za-z]|$)`)
	reNxNN   = regexp.MustCompile(`(?i)(?:^|[^0-9A-Za-z])(\d{1,2})x(\d{1,3})(?:[^0-9A-Za-z]|$)`)
	reSeason = regexp.MustCompile(`(?i)(?:season|sezon|сезон)[- .]*(\d{1,3})|(\d{1,3})[- .]*(?:season|sezon|сезон)`)
	reEpWord = regexp.MustCompile(`(?i)(?:episode|ep(?:isode)?|серия|серии)[- .]*(\d{1,4})`)
	reEOnly  = regexp.MustCompile(`(?i)(?:^|[^0-9A-Za-z])e(\d{1,4})(?:[^0-9A-Za-z]|$)`)
	reSample = regexp.MustCompile(`(?i)(?:^|[._\-\s\[(])(?:sample|trailer|preview)(?:[._\-\s\])]|$)`)
)

// EpisodeRef is a parsed season/episode from a torrent file path.
type EpisodeRef struct {
	Season  int  `json:"season,omitempty"`
	Episode int  `json:"episode,omitempty"`
	Parsed  bool `json:"parsed"`
}

// TorrentSnapshot is the data get_next_unwatched needs, without the BT client.
type TorrentSnapshot struct {
	Title    string
	Category string
	Hash     string
	Files    []*state.TorrentFileStat
}

// ViewedMap is hash -> file_index -> timecode.
type ViewedMap map[string]map[int]float64

// NextUnwatched is the structured result of episode selection.
type NextUnwatched struct {
	CaughtUp           bool        `json:"caught_up"`
	ShowTitle          string      `json:"show_title,omitempty"`
	Hash               string      `json:"hash,omitempty"`
	Season             int         `json:"season,omitempty"`
	Episode            int         `json:"episode,omitempty"`
	Code               string      `json:"code,omitempty"`
	FileIndex          int         `json:"file_index,omitempty"`
	FilePath           string      `json:"file_path,omitempty"`
	PlayURL            string      `json:"play_url,omitempty"`
	ShortPlayURL       string      `json:"short_play_url,omitempty"`
	PlaylistURL        string      `json:"playlist_url,omitempty"`
	RemainingUnwatched int         `json:"remaining_unwatched"`
	LastWatched        *EpisodeRef `json:"last_watched,omitempty"`
	LastWatchedTitle   string      `json:"last_watched_title,omitempty"`
	LastWatchedPath    string      `json:"last_watched_path,omitempty"`
	MatchedTorrents    int         `json:"matched_torrents"`
	Message            string      `json:"message,omitempty"`
}

type episodeCandidate struct {
	snap    TorrentSnapshot
	file    *state.TorrentFileStat
	season  int
	episode int
	parsed  bool
	viewed  bool
}

// ParseEpisode extracts season/episode from a file path.
func ParseEpisode(path string) (season, episode int, ok bool) {
	if path == "" {
		return 0, 0, false
	}
	norm := strings.ReplaceAll(path, "\\", "/")

	if m := reSxxExx.FindStringSubmatch(norm); len(m) == 3 {
		return atoi(m[1]), atoi(m[2]), true
	}
	if m := reNxNN.FindStringSubmatch(norm); len(m) == 3 {
		return atoi(m[1]), atoi(m[2]), true
	}

	season = firstSubint(reSeason.FindStringSubmatch(norm))
	episode = firstSubint(reEpWord.FindStringSubmatch(norm))
	if episode == 0 && season > 0 {
		episode = firstSubint(reEOnly.FindStringSubmatch(norm))
	}
	if season > 0 && episode > 0 {
		return season, episode, true
	}
	if episode > 0 {
		return season, episode, true
	}
	if season > 0 {
		return season, 0, true
	}
	return 0, 0, false
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func firstSubint(m []string) int {
	for i := 1; i < len(m); i++ {
		if m[i] != "" {
			return atoi(m[i])
		}
	}
	return 0
}

func isSampleFile(path string) bool {
	base := filepath.Base(path)
	return reSample.MatchString(base)
}

func isVideoFile(path string) bool {
	return utils.GetMimeType(path) == "video/*"
}

func filterSnapshots(all []TorrentSnapshot, query, category, hash string) []TorrentSnapshot {
	query = strings.ToLower(strings.TrimSpace(query))
	category = strings.ToLower(strings.TrimSpace(category))
	hash = strings.ToLower(strings.TrimSpace(hash))

	var out []TorrentSnapshot
	for _, s := range all {
		if hash != "" && !strings.EqualFold(s.Hash, hash) {
			continue
		}
		if hash == "" && category != "" && category != "all" {
			cat := strings.ToLower(strings.TrimSpace(s.Category))
			if category == "uncategorized" {
				if cat != "" {
					continue
				}
			} else if cat != category {
				if cat != "" || query == "" {
					continue
				}
			}
		}
		if query != "" && !strings.Contains(strings.ToLower(s.Title), query) &&
			!strings.Contains(strings.ToLower(s.Hash), query) {
			titleFromFiles := false
			for _, f := range s.Files {
				if strings.Contains(strings.ToLower(f.Path), query) {
					titleFromFiles = true
					break
				}
			}
			if !titleFromFiles {
				continue
			}
		}
		out = append(out, s)
	}
	return out
}

// SelectNextUnwatched picks the next unwatched video file across snapshots.
func SelectNextUnwatched(all []TorrentSnapshot, viewed ViewedMap, query, category, hash string) NextUnwatched {
	snaps := filterSnapshots(all, query, category, hash)
	if len(snaps) == 0 {
		return NextUnwatched{
			CaughtUp: true,
			Message:  "no matching torrents",
		}
	}

	var cands []episodeCandidate
	var last *episodeCandidate
	for _, snap := range snaps {
		vm := viewed[strings.ToLower(snap.Hash)]
		if vm == nil {
			vm = viewed[snap.Hash]
		}
		for _, f := range snap.Files {
			if f == nil || !isVideoFile(f.Path) || isSampleFile(f.Path) {
				continue
			}
			s, e, parsed := ParseEpisode(f.Path)
			c := episodeCandidate{snap: snap, file: f, season: s, episode: e, parsed: parsed}
			if vm != nil {
				if _, ok := vm[f.Id]; ok {
					c.viewed = true
					if last == nil || cmpCandidate(c, *last) > 0 {
						cp := c
						last = &cp
					}
				}
			}
			cands = append(cands, c)
		}
	}

	unwatched := 0
	for _, c := range cands {
		if !c.viewed {
			unwatched++
		}
	}

	sort.SliceStable(cands, func(i, j int) bool {
		return cmpCandidate(cands[i], cands[j]) < 0
	})

	var next *episodeCandidate
	for i := range cands {
		if !cands[i].viewed {
			next = &cands[i]
			break
		}
	}

	res := NextUnwatched{
		MatchedTorrents:    len(snaps),
		RemainingUnwatched: unwatched,
	}
	if last != nil {
		res.LastWatched = &EpisodeRef{Season: last.season, Episode: last.episode, Parsed: last.parsed}
		res.LastWatchedTitle = last.snap.Title
		res.LastWatchedPath = last.file.Path
	}

	if next == nil {
		res.CaughtUp = true
		if last != nil {
			res.ShowTitle = last.snap.Title
			res.Hash = last.snap.Hash
			res.Season = last.season
			res.Episode = last.episode
			res.Code = formatEpisodeCode(last.season, last.episode)
			res.FileIndex = last.file.Id
			res.FilePath = last.file.Path
			res.Message = "all matching episodes are marked viewed"
		} else {
			res.Message = "no video files found in matching torrents"
		}
		return res
	}

	res.CaughtUp = false
	res.ShowTitle = next.snap.Title
	res.Hash = next.snap.Hash
	res.Season = next.season
	res.Episode = next.episode
	res.Code = formatEpisodeCode(next.season, next.episode)
	res.FileIndex = next.file.Id
	res.FilePath = next.file.Path
	if next.parsed && next.season > 0 {
		res.Message = "next unwatched episode"
	} else {
		res.Message = "next unwatched video file"
	}
	return res
}

func cmpCandidate(a, b episodeCandidate) int {
	as, bs := a.season, b.season
	ae, be := a.episode, b.episode
	if !a.parsed {
		as, ae = 1<<30, a.file.Id
	}
	if !b.parsed {
		bs, be = 1<<30, b.file.Id
	}
	if as != bs {
		return as - bs
	}
	if ae != be {
		return ae - be
	}
	if a.file.Id != b.file.Id {
		return a.file.Id - b.file.Id
	}
	return strings.Compare(strings.ToLower(a.snap.Title), strings.ToLower(b.snap.Title))
}
