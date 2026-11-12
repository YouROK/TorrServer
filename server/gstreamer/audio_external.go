//go:build gst

package gstreamer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"server/settings"
)

// External audio: an audio file that lies next to the video inside the torrent (found by
// the /audio endpoint, see server/extfiles) can be used instead of the audio track of the
// video container. It is selected by the same "audio" parameter as embedded tracks,
// with an index that starts at externalAudioBase: audio=1000 is the first track of the
// /audio list, 1001 the second, and so on.
//
// The pipeline reads the audio file from its own source, decodes it and encodes it to
// AAC, so that it can be muxed into the fragmented MP4 next to the video.

const (
	// Embedded audio tracks use small indexes (0, 1, ...). External tracks start here
	// so that the two ranges never overlap.
	externalAudioBase = 1000

	maxExternalAudioListSize = 1 << 20
	externalAudioListTimeout = 15 * time.Second
)

var (
	ErrAudioNotFound         = errors.New("external audio track not found")
	ErrExternalAudioUnusable = errors.New("external audio file has no audio stream")
)

// ExternalAudio is an external audio file chosen for a task.
type ExternalAudio struct {
	FileID    string    // index of the audio file in the torrent
	SourceURL string    // where the pipeline reads the file from
	Track     TrackInfo // the audio stream of the file
}

// externalAudioItem is an element of the JSON list returned by /audio/{hash}/{id}.
type externalAudioItem struct {
	Index  int    `json:"index"`
	Name   string `json:"name"`
	Title  string `json:"title"`
	Lang   string `json:"lang"`
	Format string `json:"format"`
	Size   int64  `json:"size"`
}

func isExternalAudio(audio int) bool {
	return audio >= externalAudioBase
}

func fetchExternalAudioList(ctx context.Context, base string) ([]externalAudioItem, error) {
	body, err := httpGetLimited(ctx, base, maxExternalAudioListSize)
	if err != nil {
		return nil, err
	}
	var list []externalAudioItem
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("decode audio list: %w", err)
	}
	return list, nil
}

// pickExternalAudio returns the track the "audio" index refers to.
func pickExternalAudio(list []externalAudioItem, audio int) (externalAudioItem, error) {
	position := audio - externalAudioBase
	if position < 0 || position >= len(list) {
		return externalAudioItem{}, fmt.Errorf("%w: audio=%d, the torrent offers %d", ErrAudioNotFound, audio, len(list))
	}
	return list[position], nil
}

// resolveExternalAudio finds the external audio track for the "audio" index of the video
// file fileID and probes the audio file.
func (s *Service) resolveExternalAudio(hash string, fileID string, audio int) (*ExternalAudio, error) {
	ctx, cancel := context.WithTimeout(context.Background(), externalAudioListTimeout)
	defer cancel()

	base := settings.LoopbackBaseURL() + "/audio/" + url.PathEscape(hash) + "/" + url.PathEscape(fileID)
	list, err := fetchExternalAudioList(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("list external audio: %w", err)
	}
	item, err := pickExternalAudio(list, audio)
	if err != nil {
		return nil, err
	}

	audioFileID := strconv.Itoa(item.Index)
	probe, err := s.probeAudioFile(hash, audioFileID)
	if err != nil {
		return nil, fmt.Errorf("probe external audio %q: %w", item.Name, err)
	}
	track := probe.Audio()
	if track == nil {
		return nil, ErrExternalAudioUnusable
	}
	return &ExternalAudio{
		FileID:    audioFileID,
		SourceURL: sourceURL(s.currentConfig(), hash, audioFileID),
		Track:     *track,
	}, nil
}

// probeAudioFile probes an audio file of the torrent. Unlike Probe it does not demand a
// video stream. The result shares the probe cache: the key is the index of the audio file.
func (s *Service) probeAudioFile(hash string, fileID string) (ProbeInfo, error) {
	if s.disposed.Load() {
		return ProbeInfo{}, ErrServiceClosed
	}
	if probe, ok := s.getCachedProbe(hash, fileID); ok && probe.HasAudio() {
		return probe, nil
	}

	value, err, _ := s.probeCalls.Do(probeCacheKey(hash, fileID), func() (any, error) {
		if s.disposed.Load() {
			return ProbeInfo{}, ErrServiceClosed
		}
		conf := s.currentConfig()
		result, err := probeSource(sourceURL(conf, hash, fileID), conf)
		if err != nil {
			return ProbeInfo{}, err
		}
		if !result.HasAudio() {
			return ProbeInfo{}, ErrExternalAudioUnusable
		}
		s.setCachedProbe(hash, fileID, result)
		return result, nil
	})
	if err != nil {
		return ProbeInfo{}, err
	}
	return value.(ProbeInfo), nil
}
