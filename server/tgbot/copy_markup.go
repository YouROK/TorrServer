package tgbot

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type copyTextFixTransport struct {
	base http.RoundTripper
}

func wrapTelegramClient(c *http.Client) *http.Client {
	if c == nil {
		c = &http.Client{}
	}
	c.Transport = copyTextFixTransport{base: c.Transport}
	return c
}

func (t copyTextFixTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	if req != nil && req.Body != nil && strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		body, err := io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, err
		}
		body = rewriteTelegramJSONBody(body)
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.ContentLength = int64(len(body))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}
	}
	return base.RoundTrip(req)
}

func (t copyTextFixTransport) CloseIdleConnections() {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	if c, ok := base.(interface{ CloseIdleConnections() }); ok {
		c.CloseIdleConnections()
	}
}

func rewriteTelegramJSONBody(body []byte) []byte {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(body, &root); err != nil {
		return body
	}
	rm, ok := root["reply_markup"]
	if !ok {
		return body
	}
	var asString string
	if err := json.Unmarshal(rm, &asString); err == nil {
		fixed := omitEmptySwitchOnCopy([]byte(asString))
		if bytes.Equal(fixed, []byte(asString)) {
			return body
		}
		enc, err := json.Marshal(string(fixed))
		if err != nil {
			return body
		}
		root["reply_markup"] = enc
		out, err := json.Marshal(root)
		if err != nil {
			return body
		}
		return out
	}
	fixed := omitEmptySwitchOnCopy(rm)
	if bytes.Equal(fixed, rm) {
		return body
	}
	root["reply_markup"] = append(json.RawMessage(nil), fixed...)
	out, err := json.Marshal(root)
	if err != nil {
		return body
	}
	return out
}

func omitEmptySwitchOnCopy(markupJSON []byte) []byte {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(markupJSON, &root); err != nil {
		return markupJSON
	}
	kbRaw, ok := root["inline_keyboard"]
	if !ok {
		return markupJSON
	}
	var rows [][]map[string]json.RawMessage
	if err := json.Unmarshal(kbRaw, &rows); err != nil {
		return markupJSON
	}
	changed := false
	for i := range rows {
		for j := range rows[i] {
			btn := rows[i][j]
			if _, hasCopy := btn["copy_text"]; !hasCopy {
				continue
			}
			sw, ok := btn["switch_inline_query_current_chat"]
			if !ok {
				continue
			}
			var s string
			if err := json.Unmarshal(sw, &s); err != nil || s != "" {
				continue
			}
			delete(btn, "switch_inline_query_current_chat")
			changed = true
		}
	}
	if !changed {
		return markupJSON
	}
	kb, err := json.Marshal(rows)
	if err != nil {
		return markupJSON
	}
	root["inline_keyboard"] = kb
	out, err := json.Marshal(root)
	if err != nil {
		return markupJSON
	}
	return out
}
