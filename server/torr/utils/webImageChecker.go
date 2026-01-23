package utils

import (
	"context"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/image/webp"

	"server/log"
)

func CheckImgUrl(link string) bool {
	if link == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", link, nil)
	if err != nil {
		log.TLogln("Error create request for image:", err)
		return false
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.TLogln("Error check image:", err)
		return false
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, 512*1024)

	if strings.HasSuffix(link, ".webp") {
		_, err = webp.Decode(limitedReader)
	} else {
		_, _, err = image.Decode(limitedReader)
	}
	if err != nil {
		log.TLogln("Error decode image:", err)
		return false
	}
	return true
}
