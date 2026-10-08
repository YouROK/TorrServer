package torr

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestStreamBufferReadAndSeek(t *testing.T) {
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 1000)
	r := newBufferedStreamReader(bytes.NewReader(data), 4096)
	p := make([]byte, 17)
	if _, e := io.ReadFull(r, p); e != nil || !bytes.Equal(p, data[:17]) {
		t.Fatal(e)
	}
	if pos, e := r.Seek(0, io.SeekCurrent); e != nil || pos != 17 {
		t.Fatalf("current %d %v", pos, e)
	}
	if _, e := io.ReadFull(r, p); e != nil || !bytes.Equal(p, data[17:34]) {
		t.Fatal(e)
	}
	if pos, e := r.Seek(-10, io.SeekCurrent); e != nil || pos != 24 {
		t.Fatalf("relative %d %v", pos, e)
	}
	if _, e := io.ReadFull(r, p); e != nil || !bytes.Equal(p, data[24:41]) {
		t.Fatal(e)
	}
	if _, e := r.Seek(-19, io.SeekEnd); e != nil {
		t.Fatal(e)
	}
	tail, e := io.ReadAll(r)
	if e != nil || !bytes.Equal(tail, data[len(data)-19:]) {
		t.Fatal("tail", e)
	}
}

func TestStreamBufferFailedSeekKeepsUnreadData(t *testing.T) {
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 100)
	r := newBufferedStreamReader(bytes.NewReader(data), 4096)
	p := make([]byte, 17)
	if _, err := io.ReadFull(r, p); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Seek(-1, io.SeekStart); err == nil {
		t.Fatal("expected failed seek")
	}
	if _, err := io.ReadFull(r, p); err != nil || !bytes.Equal(p, data[17:34]) {
		t.Fatalf("failed seek discarded unread data: %v", err)
	}
}

type streamCountingReader struct {
	source *bytes.Reader
	reads  int
}

func (r *streamCountingReader) Read(p []byte) (int, error) {
	r.reads++
	return r.source.Read(p)
}

func (r *streamCountingReader) Seek(offset int64, whence int) (int64, error) {
	return r.source.Seek(offset, whence)
}

func TestStreamBufferAmortizesHTTPReads(t *testing.T) {
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 160000)
	serve := func(buffered bool) int {
		source := &streamCountingReader{source: bytes.NewReader(data)}
		var reader io.ReadSeeker = source
		if buffered {
			reader = newBufferedStreamReader(source, 1<<20)
		}
		response := httptest.NewRecorder()
		response.Header().Set("Content-Type", "application/octet-stream")
		http.ServeContent(response, httptest.NewRequest(http.MethodGet, "http://local/video.mkv", nil), "video.mkv", time.Unix(1, 0), reader)
		if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), data) {
			t.Fatal("HTTP body changed")
		}
		return source.reads
	}
	unbuffered, buffered := serve(false), serve(true)
	if buffered*8 >= unbuffered {
		t.Fatalf("small HTTP reads were not amortized: unbuffered=%d buffered=%d", unbuffered, buffered)
	}
}

func TestStreamBufferHTTPRanges(t *testing.T) {
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 1000)
	for _, rangeHeader := range []string{"bytes=13-512", "bytes=-19", "bytes=25000-", "bytes=27000-", "bytes=100-20", ""} {
		serve := func(buffered bool) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodGet, "http://local/video.mkv", nil)
			req.Header.Set("Range", rangeHeader)
			resp := httptest.NewRecorder()
			var r io.ReadSeeker = bytes.NewReader(data)
			if buffered {
				r = newBufferedStreamReader(r, 4096)
			}
			http.ServeContent(resp, req, "video.mkv", time.Unix(1, 0), r)
			return resp
		}
		a, b := serve(false), serve(true)
		if a.Code != b.Code || !bytes.Equal(a.Body.Bytes(), b.Body.Bytes()) || a.Header().Get("Content-Range") != b.Header().Get("Content-Range") {
			t.Fatal(rangeHeader)
		}
	}
}

func TestStreamBufferMultipartAndHead(t *testing.T) {
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 100000)
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		req := httptest.NewRequest(method, "http://local/video.mkv", nil)
		req.Header.Set("Range", "bytes=17-8191,1048589-1058589")
		a, b := httptest.NewRecorder(), httptest.NewRecorder()
		http.ServeContent(a, req, "video.mkv", time.Unix(1, 0), bytes.NewReader(data))
		http.ServeContent(b, req, "video.mkv", time.Unix(1, 0), newBufferedStreamReader(bytes.NewReader(data), 1<<20))
		if a.Code != b.Code || a.Header().Get("Content-Length") != b.Header().Get("Content-Length") {
			t.Fatal("headers differ")
		}
		if method == http.MethodHead {
			if a.Body.Len() != 0 || b.Body.Len() != 0 {
				t.Fatal("HEAD returned a body")
			}
			continue
		}
		decode := func(resp *httptest.ResponseRecorder) [][]byte {
			_, params, err := mime.ParseMediaType(resp.Header().Get("Content-Type"))
			if err != nil {
				t.Fatal(err)
			}
			reader := multipart.NewReader(bytes.NewReader(resp.Body.Bytes()), params["boundary"])
			var parts [][]byte
			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				body, err := io.ReadAll(part)
				if err != nil {
					t.Fatal(err)
				}
				parts = append(parts, body)
			}
			return parts
		}
		if !reflect.DeepEqual(decode(a), decode(b)) {
			t.Fatal("multipart bytes differ")
		}
	}
}
