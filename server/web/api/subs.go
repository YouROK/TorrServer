package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server/subs"
)

type subTrackJS struct {
	subs.Track
	URL string `json:"url"` // relative link: prepend scheme and host
}

// subsTracks godoc
//
//	@Summary		External subtitles of a video file
//	@Description	Without "track": list external subtitle tracks found in the torrent for the given video file.
//	@Description	With "track": return that subtitle file, by default converted to WebVTT.
//	@Description	"format": "vtt" (default) or "original" (no conversion). "encoding": output charset, default utf8.
//
//	@Tags			API
//
//	@Param			hash	path	string	true	"Torrent infohash"
//	@Param			id		path	string	true	"Video file index in torrent"
//	@Param			track	query	int		false	"Subtitle file index (from the list)"
//	@Param			format	query	string	false	"vtt (default) or original"
//	@Param			encoding	query	string	false	"Output encoding, default utf8 (also: windows-1251)"
//
//	@Produce		json
//	@Produce		text/vtt
//	@Success		200	"List of tracks (JSON) or WebVTT"
//	@Router			/subs/{hash}/{id} [get]
func subsTracks(c *gin.Context) {
	vc, ok := loadVideoContext(c)
	if !ok {
		return
	}
	tor := vc.tor

	tracks := subs.Match(vc.video, vc.files)

	// Mode 1: list the tracks.
	trackStr := c.Query("track")
	if trackStr == "" {
		base := "/subs/" + vc.infoHash + "/" + strconv.Itoa(vc.videoID)
		out := make([]subTrackJS, 0, len(tracks))
		for _, t := range tracks {
			out = append(out, subTrackJS{Track: t, URL: base + "?track=" + strconv.Itoa(t.FileIndex)})
		}
		c.JSON(http.StatusOK, out)
		return
	}

	// Mode 2: one track (WebVTT or original). Only tracks found by Match are served.
	trackID, err := strconv.Atoi(trackStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("\"track\" is wrong"))
		return
	}
	var track *subs.Track
	for i := range tracks {
		if tracks[i].FileIndex == trackID {
			track = &tracks[i]
			break
		}
	}
	if track == nil {
		c.AbortWithError(http.StatusNotFound, errors.New("subtitle track not found"))
		return
	}

	// Output options (validated before reading the file, so we do not wait for the torrent in vain):
	//   format   - "vtt" (default) or "original";
	//   encoding - response encoding, utf8 by default (see subs.Encode).
	format := strings.ToLower(c.DefaultQuery("format", "vtt"))
	if format != "vtt" && format != "original" {
		c.AbortWithError(http.StatusBadRequest, errors.New("\"format\" must be \"vtt\" or \"original\""))
		return
	}
	encoding := c.DefaultQuery("encoding", "utf8")
	if _, _, err := subs.Encode("", encoding); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	started := time.Now()
	raw, err := tor.ReadFile(ctx, track.FileIndex, subs.MaxSubSize)
	if err != nil {
		log.Printf("subs: read file %d (%s) failed after %s: %v", track.FileIndex, track.Name, time.Since(started).Round(time.Millisecond), err)
		if errors.Is(err, context.DeadlineExceeded) {
			c.AbortWithError(http.StatusGatewayTimeout, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	var (
		text        string
		contentType string
	)
	if format == "original" {
		// Original format (srt/ass/ssa/vtt) without conversion; only the encoding changes
		// (the source encoding is detected automatically and converted to the requested one).
		text = subs.Decode(raw)
		contentType = "text/plain"
	} else {
		text, err = subs.ToVTT(raw, track.Format)
		if err != nil {
			log.Printf("subs: convert file %d (%s) failed: %v", track.FileIndex, track.Name, err)
			c.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}
		contentType = "text/vtt"
	}

	out, charset, err := subs.Encode(text, encoding)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// CORS headers are not set here on purpose: the global CORS middleware already adds
	// Access-Control-Allow-Origin (a duplicated header breaks browsers).
	c.Data(http.StatusOK, contentType+"; charset="+charset, out)
}
