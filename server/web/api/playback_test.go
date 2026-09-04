package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"server/playback"
)

func TestValidatePlaybackRequest(t *testing.T) {
	validHash := strings.Repeat("a", 40)

	fileName, err := validatePlaybackRequest(playbackRequest{
		Path:  `folder\Movie.mkv`,
		Hash:  validHash,
		Index: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if fileName != "Movie.mkv" {
		t.Fatalf("unexpected filename: %q", fileName)
	}

	invalid := []playbackRequest{
		{Path: "Movie.mkv", Hash: "bad", Index: 1},
		{Path: "Movie.mkv", Hash: validHash, Index: -1},
		{Path: "Movie.mkv", Hash: validHash, Index: 100001},
		{Path: "..", Hash: validHash, Index: 1},
	}
	for _, request := range invalid {
		if _, err := validatePlaybackRequest(request); err == nil {
			t.Fatalf("expected validation error for %#v", request)
		}
	}
}

func TestPlaybackStreamURLUsesForwardedRequestAddress(t *testing.T) {
	if err := playback.Init(t.TempDir(), false); err != nil {
		t.Fatal(err)
	}
	if _, err := playback.Upsert(playback.DeviceInput{
		ID:       "tv",
		Name:     "TV",
		Endpoint: "http://127.0.0.1:9000",
	}); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest(http.MethodPost, "http://internal/playback/play", nil)
	request.Host = "movies.example.com"
	request.Header.Set("X-Forwarded-Proto", "https")
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = request

	streamURL, err := playbackStreamURL(context, "tv", "My Movie.mkv", strings.Repeat("b", 40), 7)
	if err != nil {
		t.Fatal(err)
	}
	expected := "https://movies.example.com/stream/My%20Movie.mkv?index=7&link=" + strings.Repeat("b", 40) + "&play="
	if streamURL != expected {
		t.Fatalf("unexpected stream URL:\nwant %s\n got %s", expected, streamURL)
	}
}

func TestPlaybackStreamURLUsesDeviceOverride(t *testing.T) {
	if err := playback.Init(t.TempDir(), false); err != nil {
		t.Fatal(err)
	}
	if _, err := playback.Upsert(playback.DeviceInput{
		ID:            "tv",
		Name:          "TV",
		Endpoint:      "http://127.0.0.1:9000",
		StreamBaseURL: "http://192.168.1.10:8090/torrserver",
	}); err != nil {
		t.Fatal(err)
	}

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "http://ignored/playback/play", nil)

	streamURL, err := playbackStreamURL(context, "tv", "Movie.mkv", strings.Repeat("c", 40), 1)
	if err != nil {
		t.Fatal(err)
	}
	expected := "http://192.168.1.10:8090/torrserver/stream/Movie.mkv?index=1&link=" + strings.Repeat("c", 40) + "&play="
	if streamURL != expected {
		t.Fatalf("unexpected stream URL:\nwant %s\n got %s", expected, streamURL)
	}
}
