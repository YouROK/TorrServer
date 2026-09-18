package utils

import (
	"context"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/image/webp"

	"server/log"
)

// CheckImgUrl reports whether link may be stored.
// verified is true only when a real image body was decoded; a timeout or
// transport error is ok but not verified, so a caller with an existing poster
// should keep it.
func CheckImgUrl(link string) (ok, verified bool) {
	if link == "" {
		return false, false
	}
	u, err := url.Parse(link)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", link, nil)
	if err != nil {
		log.TLogln("Error create request for image:", err)
		return false, false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.TLogln("Error check image:", err)
		return true, false
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, 2*1024*1024)

	if strings.HasSuffix(link, ".webp") {
		_, err = webp.Decode(limitedReader)
	} else {
		_, _, err = image.Decode(limitedReader)
	}
	if err != nil {
		log.TLogln("Error decode image:", err)
		if ctx.Err() != nil {
			return true, false
		}
		return false, false
	}
	return true, true
}
