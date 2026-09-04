package mcp

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/rutor"
	"server/rutor/models"
	set "server/settings"
	"server/torr"
	"server/torr/state"
	"server/torznab"
	"server/version"
)

type emptyInput struct{}

type torrentSummary struct {
	Title      string `json:"title"`
	Name       string `json:"name,omitempty"`
	Category   string `json:"category,omitempty"`
	Hash       string `json:"hash"`
	Poster     string `json:"poster,omitempty"`
	Stat       string `json:"stat,omitempty"`
	Size       int64  `json:"size,omitempty"`
	LoadedSize int64  `json:"loaded_size,omitempty"`
	FileCount  int    `json:"file_count,omitempty"`
}

type fileInfo struct {
	ID       int     `json:"id"`
	Path     string  `json:"path"`
	Length   int64   `json:"length,omitempty"`
	Mime     string  `json:"mime,omitempty"`
	Viewed   bool    `json:"viewed"`
	TimeCode float64 `json:"timecode,omitempty"`
	Season   int     `json:"season,omitempty"`
	Episode  int     `json:"episode,omitempty"`
	Code     string  `json:"code,omitempty"`
	PlayURL  string  `json:"play_url,omitempty"`
}

type serverInfoOut struct {
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	MCPEndpoint   string   `json:"mcp_endpoint"`
	BaseURL       string   `json:"base_url"`
	Categories    []string `json:"categories"`
	AuthRequired  bool     `json:"auth_required"`
	ReadOnly      bool     `json:"read_only"`
	RutorSearch   bool     `json:"rutor_search"`
	TorznabSearch bool     `json:"torznab_search"`
	TrackTimecode bool     `json:"track_timecode"`
}

type listTorrentsIn struct {
	Category string `json:"category,omitempty" jsonschema:"filter by category: movie, tv, music, other, uncategorized, or all"`
	Search   string `json:"search,omitempty" jsonschema:"case-insensitive substring match on torrent title"`
}

type listTorrentsOut struct {
	Torrents []torrentSummary `json:"torrents"`
	Count    int              `json:"count"`
}

type hashIn struct {
	Hash string `json:"hash" jsonschema:"40-character torrent infohash"`
}

type getTorrentOut struct {
	torrentSummary
	Files       []fileInfo `json:"files,omitempty"`
	PlaylistURL string     `json:"playlist_url,omitempty"`
}

type addTorrentIn struct {
	Link     string `json:"link" jsonschema:"magnet URI, infohash, http(s) .torrent URL, or torrs:// token"`
	Title    string `json:"title,omitempty" jsonschema:"optional display title"`
	Category string `json:"category,omitempty" jsonschema:"movie, tv, music, or other"`
	Poster   string `json:"poster,omitempty" jsonschema:"optional poster image URL"`
	Save     *bool  `json:"save,omitempty" jsonschema:"save to library DB; default true"`
}

type addTorrentOut struct {
	torrentSummary
	Files       []fileInfo `json:"files,omitempty"`
	PlaylistURL string     `json:"playlist_url,omitempty"`
	Saved       bool       `json:"saved"`
	Message     string     `json:"message,omitempty"`
}

type updateTorrentIn struct {
	Hash     string  `json:"hash" jsonschema:"40-character torrent infohash"`
	Title    *string `json:"title,omitempty" jsonschema:"new display title"`
	Category *string `json:"category,omitempty" jsonschema:"movie, tv, music, or other"`
	Poster   *string `json:"poster,omitempty" jsonschema:"poster image URL"`
}

type okOut struct {
	OK      bool   `json:"ok"`
	Hash    string `json:"hash,omitempty"`
	Message string `json:"message,omitempty"`
}

type playURLIn struct {
	Hash      string `json:"hash" jsonschema:"40-character torrent infohash"`
	FileIndex int    `json:"file_index" jsonschema:"1-based file index from get_torrent"`
}

type playURLOut struct {
	PlayURL      string `json:"play_url"`
	ShortPlayURL string `json:"short_play_url"`
	FilePath     string `json:"file_path,omitempty"`
	Title        string `json:"title,omitempty"`
}

type playlistURLIn struct {
	Hash     string `json:"hash,omitempty" jsonschema:"torrent infohash; omit for a playlist of all torrents"`
	Category string `json:"category,omitempty" jsonschema:"optional category filter for the all-torrents playlist"`
}

type playlistURLOut struct {
	PlaylistURL string `json:"playlist_url"`
}

type listViewedIn struct {
	Hash string `json:"hash,omitempty" jsonschema:"optional infohash; omit to list all viewed marks"`
}

type listViewedOut struct {
	Viewed []set.Viewed `json:"viewed"`
	Count  int          `json:"count"`
}

type markViewedIn struct {
	Hash      string  `json:"hash" jsonschema:"40-character torrent infohash"`
	FileIndex int     `json:"file_index" jsonschema:"1-based file index"`
	TimeCode  float64 `json:"timecode,omitempty" jsonschema:"optional playback position in seconds"`
}

type unmarkViewedIn struct {
	Hash      string `json:"hash" jsonschema:"40-character torrent infohash"`
	FileIndex *int   `json:"file_index,omitempty" jsonschema:"1-based file index; omit or pass -1 to clear all files for this torrent"`
}

type nextUnwatchedIn struct {
	Query    string `json:"query,omitempty" jsonschema:"show title substring to match"`
	Hash     string `json:"hash,omitempty" jsonschema:"limit to one torrent infohash"`
	Category string `json:"category,omitempty" jsonschema:"library category; default tv"`
}

type searchTorrentsIn struct {
	Query  string `json:"query" jsonschema:"search query"`
	Source string `json:"source,omitempty" jsonschema:"all (default), rutor, or torznab"`
	Limit  int    `json:"limit,omitempty" jsonschema:"max results to return; default 25"`
}

type searchHit struct {
	Title      string `json:"title"`
	Magnet     string `json:"magnet,omitempty"`
	Hash       string `json:"hash,omitempty"`
	Size       string `json:"size,omitempty"`
	Seed       int    `json:"seed,omitempty"`
	Peer       int    `json:"peer,omitempty"`
	Tracker    string `json:"tracker,omitempty"`
	Categories string `json:"categories,omitempty"`
	Year       int    `json:"year,omitempty"`
	IMDBID     string `json:"imdb_id,omitempty"`
}

type searchTorrentsOut struct {
	Results []searchHit `json:"results"`
	Count   int         `json:"count"`
	Message string      `json:"message,omitempty"`
}

func registerTools(server *mcpsdk.Server) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "get_server_info",
		Description: "Get TorrServer version, base URL, categories, and whether search/auth are enabled.",
	}, getServerInfo)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "list_torrents",
		Description: "List torrents in the TorrServer library. Optionally filter by category (movie, tv, music, other, uncategorized, all) and title search.",
	}, listTorrents)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "get_torrent",
		Description: "Get one torrent by infohash, including files, viewed flags, season/episode codes, and play URLs.",
	}, getTorrent)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "add_torrent",
		Description: "Add a torrent from a magnet URI, 40-character infohash, http(s) .torrent URL, or torrs:// token. Waits for metadata and returns files and play URLs. Saves to the library by default.",
	}, addTorrent)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "update_torrent",
		Description: "Update a torrent's title, category, or poster in memory and in the library DB.",
	}, updateTorrent)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "remove_torrent",
		Description: "Permanently remove a torrent from memory, the library DB, and disk cache.",
	}, removeTorrent)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "drop_torrent",
		Description: "Unload a torrent from the BitTorrent client but keep it in the library DB.",
	}, dropTorrent)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "get_play_url",
		Description: "Build HTTP stream URLs for a torrent file so the user can open them in VLC, mpv, or a browser.",
	}, getPlayURL)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "get_playlist_url",
		Description: "Build an M3U playlist URL for one torrent (hash) or the whole library.",
	}, getPlaylistURL)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "list_viewed",
		Description: "List viewed file marks (and optional timecodes) for one torrent or the whole library.",
	}, listViewed)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "mark_viewed",
		Description: "Mark a torrent file as viewed. Use after the user watches an episode so get_next_unwatched skips it.",
	}, markViewed)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "unmark_viewed",
		Description: "Clear viewed status for one file or all files in a torrent.",
	}, unmarkViewed)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "get_next_unwatched",
		Description: "Find the next unwatched TV episode from library torrents by parsing filenames (SxxEyy, 1x05, Season N / Episode M, Russian сезон/серия) and comparing viewed marks. Default category is tv. Returns a play URL.",
	}, getNextUnwatched)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "search_torrents",
		Description: "Search RuTor and/or Torznab indexers (if enabled in settings). Returns magnet links you can pass to add_torrent.",
	}, searchTorrents)
}

func getServerInfo(ctx context.Context, req *mcpsdk.CallToolRequest, _ emptyInput) (*mcpsdk.CallToolResult, serverInfoOut, error) {
	_ = ctx
	base := baseURL(req)
	rutorOn, torznabOn, trackTC := false, false, false
	if set.BTsets != nil {
		rutorOn = set.BTsets.EnableRutorSearch
		torznabOn = set.BTsets.EnableTorznabSearch
		trackTC = set.BTsets.TrackTimecode
	}
	return nil, serverInfoOut{
		Name:          "torrserver",
		Version:       version.Version,
		MCPEndpoint:   strings.TrimRight(base, "/") + "/mcp",
		BaseURL:       base,
		Categories:    []string{"movie", "tv", "music", "other"},
		AuthRequired:  set.HttpAuth,
		ReadOnly:      set.ReadOnly,
		RutorSearch:   rutorOn,
		TorznabSearch: torznabOn,
		TrackTimecode: trackTC,
	}, nil
}

func listTorrents(ctx context.Context, req *mcpsdk.CallToolRequest, in listTorrentsIn) (*mcpsdk.CallToolResult, listTorrentsOut, error) {
	_ = ctx
	_ = req
	category := strings.ToLower(strings.TrimSpace(in.Category))
	search := strings.ToLower(strings.TrimSpace(in.Search))
	var items []torrentSummary
	for _, t := range torr.ListTorrent() {
		st := t.Status()
		if st == nil {
			continue
		}
		if category != "" && category != "all" {
			cat := strings.ToLower(strings.TrimSpace(st.Category))
			if category == "uncategorized" {
				if cat != "" {
					continue
				}
			} else if cat != category {
				continue
			}
		}
		if search != "" && !strings.Contains(strings.ToLower(st.Title), search) &&
			!strings.Contains(strings.ToLower(st.Name), search) {
			continue
		}
		items = append(items, compactStatus(st))
	}
	if items == nil {
		items = []torrentSummary{}
	}
	return nil, listTorrentsOut{Torrents: items, Count: len(items)}, nil
}

func getTorrent(ctx context.Context, req *mcpsdk.CallToolRequest, in hashIn) (*mcpsdk.CallToolResult, getTorrentOut, error) {
	_ = ctx
	tor, err := loadTorrent(in.Hash)
	if err != nil {
		return nil, getTorrentOut{}, err
	}
	st := tor.Status()
	base := baseURL(req)
	out := getTorrentOut{
		torrentSummary: compactStatus(st),
		Files:          fileInfos(base, st),
		PlaylistURL:    playlistURL(base, st.Hash),
	}
	return nil, out, nil
}

func addTorrent(ctx context.Context, req *mcpsdk.CallToolRequest, in addTorrentIn) (*mcpsdk.CallToolResult, addTorrentOut, error) {
	_ = ctx
	if set.ReadOnly {
		return nil, addTorrentOut{}, fmt.Errorf("server is in read-only mode")
	}
	spec, title, poster, category, err := parseTorrentLink(in.Link)
	if err != nil {
		return nil, addTorrentOut{}, err
	}
	if in.Title != "" {
		title = in.Title
	}
	if in.Poster != "" {
		poster = in.Poster
	}
	if in.Category != "" {
		category = in.Category
	}
	logAdd(in.Link)
	tor, err := torr.AddTorrent(spec, title, poster, "", category)
	if err != nil {
		return nil, addTorrentOut{}, err
	}
	if tor == nil {
		return nil, addTorrentOut{}, fmt.Errorf("torrent not created")
	}

	got := tor.GotInfo()
	if got {
		if tor.Title == "" {
			tor.Title = spec.DisplayName
			tor.Title = strings.ReplaceAll(tor.Title, "rutor.info", "")
			tor.Title = strings.ReplaceAll(tor.Title, "_", " ")
			tor.Title = strings.TrimSpace(tor.Title)
			if tor.Title == "" {
				tor.Title = tor.Name()
			}
		}
	}

	save := true
	if in.Save != nil {
		save = *in.Save
	}
	if save && got {
		torr.SaveTorrentToDB(tor)
	}
	restartDLNA()

	st := tor.Status()
	base := baseURL(req)
	out := addTorrentOut{
		torrentSummary: compactStatus(st),
		Files:          fileInfos(base, st),
		PlaylistURL:    playlistURL(base, st.Hash),
		Saved:          save && got,
	}
	if !got {
		out.Message = "added but metadata timed out; files may appear later via get_torrent"
	} else {
		out.Message = "torrent added"
	}
	return nil, out, nil
}

func updateTorrent(ctx context.Context, req *mcpsdk.CallToolRequest, in updateTorrentIn) (*mcpsdk.CallToolResult, getTorrentOut, error) {
	_ = ctx
	if set.ReadOnly {
		return nil, getTorrentOut{}, fmt.Errorf("server is in read-only mode")
	}
	if !validInfoHash(in.Hash) {
		return nil, getTorrentOut{}, fmt.Errorf("invalid torrent hash")
	}
	tor := torr.GetTorrent(in.Hash)
	if tor == nil {
		return nil, getTorrentOut{}, fmt.Errorf("torrent not found")
	}
	st := tor.Status()
	title, poster, category := st.Title, st.Poster, st.Category
	if in.Title != nil {
		title = *in.Title
	}
	if in.Poster != nil {
		poster = *in.Poster
	}
	if in.Category != nil {
		category = *in.Category
	}
	tor = torr.SetTorrent(in.Hash, title, poster, category, "")
	if tor == nil {
		return nil, getTorrentOut{}, fmt.Errorf("torrent not found")
	}
	st = tor.Status()
	base := baseURL(req)
	return nil, getTorrentOut{
		torrentSummary: compactStatus(st),
		Files:          fileInfos(base, st),
		PlaylistURL:    playlistURL(base, st.Hash),
	}, nil
}

func removeTorrent(ctx context.Context, req *mcpsdk.CallToolRequest, in hashIn) (*mcpsdk.CallToolResult, okOut, error) {
	_ = ctx
	_ = req
	if set.ReadOnly {
		return nil, okOut{}, fmt.Errorf("server is in read-only mode")
	}
	if !validInfoHash(in.Hash) {
		return nil, okOut{}, fmt.Errorf("invalid torrent hash")
	}
	if torr.GetTorrent(in.Hash) == nil {
		return nil, okOut{}, fmt.Errorf("torrent not found")
	}
	torr.RemTorrent(in.Hash)
	afterRemove(in.Hash)
	return nil, okOut{OK: true, Hash: in.Hash, Message: "torrent removed"}, nil
}

func dropTorrent(ctx context.Context, req *mcpsdk.CallToolRequest, in hashIn) (*mcpsdk.CallToolResult, okOut, error) {
	_ = ctx
	_ = req
	if !validInfoHash(in.Hash) {
		return nil, okOut{}, fmt.Errorf("invalid torrent hash")
	}
	if torr.GetTorrent(in.Hash) == nil {
		return nil, okOut{}, fmt.Errorf("torrent not found")
	}
	torr.DropTorrent(in.Hash)
	afterRemove(in.Hash)
	return nil, okOut{OK: true, Hash: in.Hash, Message: "torrent dropped from memory"}, nil
}

func getPlayURL(ctx context.Context, req *mcpsdk.CallToolRequest, in playURLIn) (*mcpsdk.CallToolResult, playURLOut, error) {
	_ = ctx
	tor, err := loadTorrent(in.Hash)
	if err != nil {
		return nil, playURLOut{}, err
	}
	st := tor.Status()
	var file *state.TorrentFileStat
	for _, f := range st.FileStats {
		if f != nil && f.Id == in.FileIndex {
			file = f
			break
		}
	}
	if file == nil {
		return nil, playURLOut{}, fmt.Errorf("file index %d not found", in.FileIndex)
	}
	base := baseURL(req)
	play, short := filePlayURLs(base, st, file)
	return nil, playURLOut{
		PlayURL:      play,
		ShortPlayURL: short,
		FilePath:     file.Path,
		Title:        st.Title,
	}, nil
}

func getPlaylistURL(ctx context.Context, req *mcpsdk.CallToolRequest, in playlistURLIn) (*mcpsdk.CallToolResult, playlistURLOut, error) {
	_ = ctx
	base := baseURL(req)
	u := playlistURL(base, in.Hash)
	if in.Hash == "" && in.Category != "" {
		u = strings.TrimRight(base, "/") + "/playlistall/all.m3u?category=" + url.QueryEscape(in.Category)
	}
	return nil, playlistURLOut{PlaylistURL: u}, nil
}

func listViewed(ctx context.Context, req *mcpsdk.CallToolRequest, in listViewedIn) (*mcpsdk.CallToolResult, listViewedOut, error) {
	_ = ctx
	_ = req
	if in.Hash != "" && !validInfoHash(in.Hash) {
		return nil, listViewedOut{}, fmt.Errorf("invalid torrent hash")
	}
	list := set.ListViewed(in.Hash)
	if list == nil {
		list = []*set.Viewed{}
	}
	vals := make([]set.Viewed, 0, len(list))
	for _, v := range list {
		if v != nil {
			vals = append(vals, *v)
		}
	}
	return nil, listViewedOut{Viewed: vals, Count: len(vals)}, nil
}

func markViewed(ctx context.Context, req *mcpsdk.CallToolRequest, in markViewedIn) (*mcpsdk.CallToolResult, okOut, error) {
	_ = ctx
	_ = req
	if set.ReadOnly {
		return nil, okOut{}, fmt.Errorf("server is in read-only mode")
	}
	if !validInfoHash(in.Hash) {
		return nil, okOut{}, fmt.Errorf("invalid torrent hash")
	}
	if in.FileIndex < 1 {
		return nil, okOut{}, fmt.Errorf("file_index must be >= 1")
	}
	set.SetViewed(&set.Viewed{Hash: in.Hash, FileIndex: in.FileIndex, TimeCode: in.TimeCode})
	return nil, okOut{OK: true, Hash: in.Hash, Message: "marked viewed"}, nil
}

func unmarkViewed(ctx context.Context, req *mcpsdk.CallToolRequest, in unmarkViewedIn) (*mcpsdk.CallToolResult, okOut, error) {
	_ = ctx
	_ = req
	if set.ReadOnly {
		return nil, okOut{}, fmt.Errorf("server is in read-only mode")
	}
	if !validInfoHash(in.Hash) {
		return nil, okOut{}, fmt.Errorf("invalid torrent hash")
	}
	idx := -1
	if in.FileIndex != nil {
		idx = *in.FileIndex
	}
	set.RemViewed(&set.Viewed{Hash: in.Hash, FileIndex: idx})
	return nil, okOut{OK: true, Hash: in.Hash, Message: "viewed status cleared"}, nil
}

func getNextUnwatched(ctx context.Context, req *mcpsdk.CallToolRequest, in nextUnwatchedIn) (*mcpsdk.CallToolResult, NextUnwatched, error) {
	_ = ctx
	category := strings.TrimSpace(in.Category)
	if category == "" {
		category = "tv"
	}
	res := SelectNextUnwatched(listSnapshots(), viewedMap(in.Hash), in.Query, category, in.Hash)
	base := baseURL(req)
	if res.Hash != "" && res.FileIndex > 0 {
		res.PlayURL = playURL(base, res.Hash, res.FilePath, res.FileIndex)
		res.ShortPlayURL = shortPlayURL(base, res.Hash, res.FileIndex)
		res.PlaylistURL = playlistURL(base, res.Hash)
	}
	return nil, res, nil
}

func searchTorrents(ctx context.Context, req *mcpsdk.CallToolRequest, in searchTorrentsIn) (*mcpsdk.CallToolResult, searchTorrentsOut, error) {
	_ = ctx
	_ = req
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, searchTorrentsOut{}, fmt.Errorf("query is required")
	}
	source := strings.ToLower(strings.TrimSpace(in.Source))
	if source == "" {
		source = "all"
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 50 {
		limit = 50
	}

	rutorOn := set.BTsets != nil && set.BTsets.EnableRutorSearch
	torznabOn := set.BTsets != nil && set.BTsets.EnableTorznabSearch
	var list []*models.TorrentDetails
	switch source {
	case "rutor":
		if !rutorOn {
			return nil, searchTorrentsOut{Results: []searchHit{}, Message: "RuTor search is disabled"}, nil
		}
		list = rutor.Search(query)
	case "torznab":
		if !torznabOn {
			return nil, searchTorrentsOut{Results: []searchHit{}, Message: "Torznab search is disabled"}, nil
		}
		list = torznab.Search(ctx, query, -1, "", 0, 0)
	case "all":
		if !rutorOn && !torznabOn {
			return nil, searchTorrentsOut{Results: []searchHit{}, Message: "search is disabled in settings"}, nil
		}
		if rutorOn {
			list = append(list, rutor.Search(query)...)
		}
		if torznabOn {
			list = append(list, torznab.Search(ctx, query, -1, "", 0, 0)...)
		}
	default:
		return nil, searchTorrentsOut{}, fmt.Errorf("source must be all, rutor, or torznab")
	}

	out := searchTorrentsOut{Results: []searchHit{}}
	for _, d := range list {
		if d == nil {
			continue
		}
		out.Results = append(out.Results, searchHit{
			Title:      d.Title,
			Magnet:     d.Magnet,
			Hash:       d.Hash,
			Size:       d.Size,
			Seed:       d.Seed,
			Peer:       d.Peer,
			Tracker:    d.Tracker,
			Categories: d.Categories,
			Year:       d.Year,
			IMDBID:     d.IMDBID,
		})
		if len(out.Results) >= limit {
			break
		}
	}
	out.Count = len(out.Results)
	return nil, out, nil
}
