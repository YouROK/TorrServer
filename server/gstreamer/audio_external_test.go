//go:build gst

package gstreamer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsExternalAudio(t *testing.T) {
	for audio, want := range map[int]bool{0: false, 1: false, 999: false, 1000: true, 1001: true} {
		if isExternalAudio(audio) != want {
			t.Errorf("isExternalAudio(%d) = %v, want %v", audio, !want, want)
		}
	}
}

func TestPickExternalAudio(t *testing.T) {
	list := []externalAudioItem{{Index: 31, Name: "a.mka"}, {Index: 29, Name: "b.mka"}}

	item, err := pickExternalAudio(list, 1001)
	if err != nil || item.Index != 29 {
		t.Fatalf("audio=1001: item = %+v, err = %v; want the second track", item, err)
	}
	for _, audio := range []int{1002, 2000, 5} {
		if _, err := pickExternalAudio(list, audio); !errors.Is(err, ErrAudioNotFound) {
			t.Errorf("audio=%d: err = %v, want ErrAudioNotFound", audio, err)
		}
	}
}

func TestFetchExternalAudioList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"index":31,"name":"ep01.mka","title":"Group - ep01.mka","lang":"ru","format":"mka","size":35000000,"url":"/stream?link=h&index=31&play"}]`)
	}))
	defer server.Close()

	list, err := fetchExternalAudioList(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Index != 31 || list[0].Title != "Group - ep01.mka" || list[0].Lang != "ru" || list[0].Size != 35000000 {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestFetchExternalAudioListErrors(t *testing.T) {
	notFound := httptest.NewServer(http.NotFoundHandler())
	defer notFound.Close()
	if _, err := fetchExternalAudioList(context.Background(), notFound.URL); err == nil {
		t.Error("want an error for 404")
	}

	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not json")
	}))
	defer broken.Close()
	if _, err := fetchExternalAudioList(context.Background(), broken.URL); err == nil {
		t.Error("want an error for a broken body")
	}
}

func TestTaskAudioTrack(t *testing.T) {
	probe := ProbeInfo{Tracks: []TrackInfo{{Type: "video"}, {Type: "audio", Index: 0, Channels: 2}}}

	embedded := &Task{Audio: 0, Probe: probe}
	if track := embedded.audioTrack(); track == nil || track.Channels != 2 || !embedded.hasAudio() {
		t.Errorf("embedded: track = %+v", track)
	}

	external := &Task{Audio: 1000, Probe: ProbeInfo{Tracks: []TrackInfo{{Type: "video"}}},
		ExternalAudio: &ExternalAudio{Track: TrackInfo{Type: "audio", Channels: 6}}}
	if track := external.audioTrack(); track == nil || track.Channels != 6 || !external.hasAudio() {
		t.Errorf("external: track = %+v", track)
	}

	silent := &Task{Probe: ProbeInfo{Tracks: []TrackInfo{{Type: "video"}}}}
	if silent.hasAudio() || silent.audioTrack() != nil {
		t.Error("a task without audio reports an audio track")
	}
}
