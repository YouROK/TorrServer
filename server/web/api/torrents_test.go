package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// The /torrents handler is documented as `@Produce json`, but its failure
// paths used to abort with an empty body, so any client calling .json() on
// the response got a parse error instead of the reason it failed.
func TestTorrentsRejectionsCarryJSONBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name string
		body string
	}{
		{"malformed json", `{`},
		{"wrong type on an optional key", `{"action":"add","link":"magnet:?xt=urn:btih:x","title":123}`},
		{"unknown action", `{"action":"bogus"}`},
		{"missing action", `{}`},
		{"add without link", `{"action":"add"}`},
		{"get without hash", `{"action":"get"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/torrents", torrents)

			req := httptest.NewRequest(http.MethodPost, "/torrents", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
			if w.Body.Len() == 0 {
				t.Fatal("response body is empty; clients parsing it as JSON fail")
			}
			var payload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatalf("body is not valid JSON (%v): %q", err, w.Body.String())
			}
			if payload.Error == "" {
				t.Errorf("no error message in body: %q", w.Body.String())
			}
		})
	}
}
