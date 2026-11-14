//go:build gst

package bridge

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"server/gstreamer"
	"server/settings"
	"server/web/auth"
)

func TestGStreamerRoutesRequireHTTPAuth(t *testing.T) {
	oldAuth, oldPath, oldSets := settings.HttpAuth, settings.Path, settings.BTsets
	t.Cleanup(func() {
		gstreamer.Stop()
		settings.HttpAuth, settings.Path, settings.BTsets = oldAuth, oldPath, oldSets
	})
	settings.HttpAuth, settings.Path = true, t.TempDir()
	settings.BTsets = &settings.BTSets{}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.BasicAuth(gin.Accounts{"viewer": "test-password"}))
	SetupRoute(router)

	for _, path := range []string{
		"/gst/remove", "/gst/echo", "/gst/fixture/heartbeat", "/gst/fixture/probe",
		"/gst/fixture/master.m3u8", "/gst/fixture/video.m3u8", "/gst/fixture/init.mp4",
		"/gst/fixture/seg/0.m4s", "/gst/fixture/subs/0.m3u8",
	} {
		for _, wrong := range []bool{false, true} {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.Header.Set("Range", "bytes=0-9")
			if wrong {
				request.SetBasicAuth("viewer", "wrong-password")
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Errorf("%s wrong=%t: status=%d, want401", path, wrong, response.Code)
			}
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/gst/remove", nil)
	request.SetBasicAuth("viewer", "test-password")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("valid auth: status=%d, want handler's400", response.Code)
	}
}
