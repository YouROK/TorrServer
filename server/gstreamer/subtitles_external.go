//go:build gst

package gstreamer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"server/settings"
)

// External subtitles: subtitle files that lie next to the video inside the torrent
// (found by the /subs endpoint, see server/subs) are offered in the HLS master
// playlist as additional WebVTT renditions, next to the embedded subtitle tracks.
//
// The whole file is available up front, so it is loaded into a subtitleStore once
// and every segment of the subtitle playlist is immediately ready. Segments are
// rendered by the same code as for embedded tracks (renderVTT).

const (
	// Embedded subtitle tracks use small indexes (0, 1, ...). External tracks start
	// here so that the two ranges never overlap.
	externalSubtitleBase = 1000

	maxExternalSubtitles        = 16
	maxExternalSubtitleListSize = 1 << 20
	maxExternalSubtitleSize     = 4 << 20
	externalSubtitleTimeout     = 15 * time.Second

	// How many times a master playlist request may retry loading tracks that failed
	// for a temporary reason (timeout, 5xx) before the task gives up on them.
	maxExternalSubtitleAttempts = 3
)

var errSubtitleTooLarge = errors.New("subtitle file is too large")

// httpStatusError is returned by httpGetLimited for a non-200 response.
type httpStatusError struct {
	URL    string
	Status int
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("GET %s: status %d", e.URL, e.Status)
}

// isPermanentSubtitleError reports whether repeating the same request cannot help:
// a client error (404, 422 for a file without dialogue, ...) or a file over the size limit.
// Timeouts, 408, 429, 5xx and network errors are temporary.
func isPermanentSubtitleError(err error) bool {
	if errors.Is(err, errSubtitleTooLarge) {
		return true
	}
	var statusErr *httpStatusError
	if errors.As(err, &statusErr) {
		status := statusErr.Status
		return status >= 400 && status < 500 &&
			status != http.StatusRequestTimeout && status != http.StatusTooManyRequests
	}
	return false
}

type externalSubtitle struct {
	Index    int
	Name     string
	Language string

	store *subtitleStore
}

// subtitleTrackJSON is one element of the JSON list returned by GET /subs/{hash}/{id}.
type subtitleTrackJSON struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Title string `json:"title"`
	Lang  string `json:"lang"`
}

func (t *Task) externalSubtitles() []externalSubtitle {
	if t == nil {
		return nil
	}
	t.externalMu.RLock()
	defer t.externalMu.RUnlock()
	return t.external
}

func (t *Task) externalSubtitleStore(index int) *subtitleStore {
	if t == nil {
		return nil
	}
	t.externalMu.RLock()
	defer t.externalMu.RUnlock()
	for _, track := range t.external {
		if track.Index == index {
			return track.store
		}
	}
	return nil
}

func (t *Task) setExternalSubtitles(list []externalSubtitle) {
	// NAME must be unique within the whole subtitle group, embedded tracks included.
	used := make(map[string]bool, len(list)+len(t.Probe.Tracks))
	if t.Config.Subtitles {
		for _, track := range t.Probe.Tracks {
			if supportedSubtitleTrack(track) {
				used[embeddedSubtitleName(track)] = true
			}
		}
	}
	counters := make(map[string]int, len(list))
	for i := range list {
		base := list[i].Name
		if used[base] {
			n := counters[base] + 1
			if n < 2 {
				n = 2
			}
			for used[fmt.Sprintf("%s (%d)", base, n)] {
				n++
			}
			counters[base] = n
			list[i].Name = fmt.Sprintf("%s (%d)", base, n)
		}
		used[list[i].Name] = true
	}

	t.externalMu.Lock()
	t.external = list
	t.externalLoaded = true
	t.externalMu.Unlock()
}

// ensureExternalSubtitles loads the external subtitle tracks of the task's video.
// It never fails the caller: problems are logged and the tracks are simply not
// offered.
func (t *Task) ensureExternalSubtitles(ctx context.Context) {
	if t == nil || !t.Config.Subtitles {
		return
	}
	base := settings.LoopbackBaseURL() + "/subs/" + url.PathEscape(t.ID) + "/" + url.PathEscape(t.FileID)
	t.loadExternalSubtitles(ctx, base)
}

// loadExternalSubtitles does the work of ensureExternalSubtitles; base is the URL of
// the /subs endpoint for the video file. If some track could not be read for a
// temporary reason (for example, the torrent is slow to start), the next call retries,
// at most maxExternalSubtitleAttempts times. Tracks that fail permanently (for example,
// 422 for an ASS file without dialogue) are skipped and never cause a retry.
func (t *Task) loadExternalSubtitles(ctx context.Context, base string) {
	t.externalLoadMu.Lock()
	defer t.externalLoadMu.Unlock()

	t.externalMu.RLock()
	loaded := t.externalLoaded
	t.externalMu.RUnlock()
	if loaded {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, externalSubtitleTimeout)
	defer cancel()

	tracks, failed, err := fetchExternalSubtitles(ctx, base)
	if err != nil {
		log.Printf("[gst] external subtitles for %s/%s: %v", t.ID, t.FileID, err)
		return
	}

	t.setExternalSubtitles(tracks)
	if failed == 0 {
		return
	}
	t.externalAttempts++
	if t.externalAttempts >= maxExternalSubtitleAttempts {
		log.Printf("[gst] external subtitles for %s/%s: giving up on %d track(s) after %d attempts", t.ID, t.FileID, failed, t.externalAttempts)
		return
	}
	t.externalMu.Lock()
	t.externalLoaded = false
	t.externalMu.Unlock()
}

// fetchExternalSubtitles asks the /subs endpoint (base is its URL for the video file)
// for the list of tracks and downloads each of them as WebVTT. Tracks that cannot be
// read are skipped; only those that failed for a temporary reason are counted in failed.
func fetchExternalSubtitles(ctx context.Context, base string) ([]externalSubtitle, int, error) {
	body, err := httpGetLimited(ctx, base, maxExternalSubtitleListSize)
	if err != nil {
		return nil, 0, err
	}
	var list []subtitleTrackJSON
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, 0, fmt.Errorf("decode subtitle list: %w", err)
	}
	if len(list) > maxExternalSubtitles {
		list = list[:maxExternalSubtitles]
	}

	tracks := make([]externalSubtitle, 0, len(list))
	failed := 0
	for i, item := range list {
		vtt, err := httpGetLimited(ctx, base+"?track="+strconv.Itoa(item.Index)+"&format=vtt", maxExternalSubtitleSize)
		if err != nil {
			log.Printf("[gst] external subtitle %q: %v", item.Name, err)
			if !isPermanentSubtitleError(err) {
				failed++
			}
			continue
		}

		store := newSubtitleStore()
		// The last cue is parsed only when followed by a blank line.
		store.appendVTT(string(vtt)+"\n\n", 0, 0)
		if len(store.cues) == 0 {
			continue
		}
		// The whole file is available up front: every segment is ready.
		store.advanceVideoReadTo(^uint64(0))

		tracks = append(tracks, externalSubtitle{
			Index:    externalSubtitleBase + i,
			Name:     firstNonEmpty(firstNonEmpty(item.Title, item.Name), "Subtitle "+strconv.Itoa(i+1)),
			Language: item.Lang,
			store:    store,
		})
	}
	return tracks, failed, nil
}

func httpGetLimited(ctx context.Context, rawURL string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &httpStatusError{URL: rawURL, Status: resp.StatusCode}
	}
	// Read one byte more than the limit to tell a file that fits from one that was cut.
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("GET %s: %w (limit %d bytes)", rawURL, errSubtitleTooLarge, limit)
	}
	return body, nil
}

// writeSubtitleMedia writes one EXT-X-MEDIA line of the subtitle group.
func writeSubtitleMedia(playlist *strings.Builder, taskID string, name string, language string, index int) {
	playlist.WriteString("#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\",NAME=\"")
	playlist.WriteString(hlsQuoted(name))
	playlist.WriteString("\",LANGUAGE=\"")
	playlist.WriteString(hlsQuoted(language))
	playlist.WriteString("\",DEFAULT=NO,AUTOSELECT=YES,FORCED=NO,URI=\"/gst/")
	playlist.WriteString(taskID)
	playlist.WriteString("/subs/")
	playlist.WriteString(strconv.Itoa(index))
	playlist.WriteString(".m3u8\"\n")
}
