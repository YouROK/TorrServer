package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dop251/goja"
)

var sharedHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
}

func (rt *JSRuntime) createHTTPModule() *goja.Object {
	httpObj := rt.vm.NewObject()

	httpObj.Set("get", func(call goja.FunctionCall) goja.Value {
		rawURL := call.Argument(0).String()
		if rawURL == "" {
			panic(rt.vm.ToValue("url is required"))
		}

		var opts map[string]any
		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			opts = call.Argument(1).Export().(map[string]any)
		}

		return rt.doHTTPRequest("GET", rawURL, nil, opts)
	})

	httpObj.Set("post", func(call goja.FunctionCall) goja.Value {
		rawURL := call.Argument(0).String()
		if rawURL == "" {
			panic(rt.vm.ToValue("url is required"))
		}

		bodyRaw := call.Argument(1).Export()
		var bodyReader io.Reader

		if strBody, ok := bodyRaw.(string); ok {
			bodyReader = strings.NewReader(strBody)
		} else if bodyRaw != nil {
			b, err := json.Marshal(bodyRaw)
			if err != nil {
				panic(rt.vm.ToValue(fmt.Sprintf("failed to serialize body to JSON: %v", err)))
			}
			bodyReader = bytes.NewReader(b)
		}

		var opts map[string]any
		if len(call.Arguments) > 2 && !goja.IsUndefined(call.Argument(2)) && !goja.IsNull(call.Argument(2)) {
			opts = call.Argument(2).Export().(map[string]any)
		}

		return rt.doHTTPRequest("POST", rawURL, bodyReader, opts)
	})

	return httpObj
}

func (rt *JSRuntime) doHTTPRequest(method, rawURL string, body io.Reader, opts map[string]any) goja.Value {
	timeout := 5 * time.Second
	if tVal, ok := opts["timeout"].(int64); ok && tVal > 0 {
		timeout = time.Duration(tVal) * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("failed to create request: %v", err)))
	}

	req.Header.Set("User-Agent", "Silo/1.3 (TorrServer)")

	if headersRaw, ok := opts["headers"].(map[string]any); ok {
		for k, v := range headersRaw {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("HTTP request failed: %v", err)))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("failed to read response body: %v", err)))
	}

	resObj := rt.vm.NewObject()
	resObj.Set("status", resp.StatusCode)
	resObj.Set("ok", resp.StatusCode >= 200 && resp.StatusCode < 300)
	resObj.Set("body", string(respBody))

	headersObj := rt.vm.NewObject()
	for k, vals := range resp.Header {
		if len(vals) > 0 {
			headersObj.Set(k, vals[0])
		}
	}
	resObj.Set("headers", headersObj)

	resObj.Set("json", func(call goja.FunctionCall) goja.Value {
		var parsed any
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to parse JSON response: %v", err)))
		}
		return rt.vm.ToValue(parsed)
	})

	return resObj
}
