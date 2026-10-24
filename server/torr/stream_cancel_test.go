package torr

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stallReadSeeker struct{}

func (stallReadSeeker) Read([]byte) (int, error) {
	time.Sleep(time.Hour)
	return 0, io.EOF
}

func (stallReadSeeker) ReadContext(ctx context.Context, _ []byte) (int, error) {
	<-ctx.Done()
	return 0, ctx.Err()
}

func (stallReadSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekEnd {
		return 4096, nil
	}
	return 0, nil
}

func TestStreamReadReturnsWhenClientGoesAway(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "http://local/v.mkv", nil).WithContext(ctx)
	body := newBufferedStreamReader(bindStreamReadContext(ctx, stallReadSeeker{}), 32)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ServeStreamContent(httptest.NewRecorder(), req, "v.mkv", time.Unix(0, 0), body)
	}()
	time.Sleep(30 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream stayed blocked after the client went away")
	}
}
