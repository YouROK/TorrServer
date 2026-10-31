package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/extfiles"
)

type audioTrackJS struct {
	extfiles.AudioTrack
	URL string `json:"url"` // relative link to the file itself: prepend scheme and host
}

// audioTracks godoc
//
//	@Summary		External audio tracks of a video file
//	@Description	List audio files found in the torrent that belong to the given video file
//	@Description	(for example voice-overs in "Sound [Group]/ep01.mka"). "url" is the link to the file itself.
//
//	@Tags			API
//
//	@Param			hash	path	string	true	"Torrent infohash"
//	@Param			id		path	string	true	"Video file index in torrent"
//
//	@Produce		json
//	@Success		200	"List of tracks (JSON)"
//	@Router			/audio/{hash}/{id} [get]
func audioTracks(c *gin.Context) {
	vc, ok := loadVideoContext(c)
	if !ok {
		return
	}

	tracks := extfiles.MatchAudio(vc.video, vc.files)
	out := make([]audioTrackJS, 0, len(tracks))
	for _, t := range tracks {
		out = append(out, audioTrackJS{
			AudioTrack: t,
			URL:        "/stream?link=" + vc.infoHash + "&index=" + strconv.Itoa(t.FileIndex) + "&play",
		})
	}
	c.JSON(http.StatusOK, out)
}
