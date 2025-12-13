package utils

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"server/log"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func ParseFile(file multipart.File) (*torrent.TorrentSpec, error) {
	minfo, err := metainfo.Load(file)
	if err != nil {
		return nil, err
	}

	spec, err := torrent.TorrentSpecFromMetaInfoErr(minfo)
	if err != nil {
		log.TLogln("Error parse torrent file info:", err)
		return nil, err
	}

	return spec, nil
}

func ParseLink(link string) (*torrent.TorrentSpec, error) {
	urlLink, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	// log.TLogln("ParseLink:", urlLink)
	switch strings.ToLower(urlLink.Scheme) {
	case "magnet":
		return fromMagnet(urlLink.String())
	case "http", "https":
		return fromHttp(urlLink.String())
	case "file":
		return fromFile(urlLink.Path)
	case "": // infohash
		return fromMagnet("magnet:?xt=urn:btih:" + urlLink.Path)
	default:
		err = fmt.Errorf("%s unknown scheme: %s", urlLink, urlLink.Scheme)
	}
	return nil, err
}

func fromMagnet(link string) (*torrent.TorrentSpec, error) {
	spec, err := torrent.TorrentSpecFromMagnetUri(link)
	if err != nil {
		log.TLogln("Error parse torrent info:", err)
		return nil, err
	}

	return spec, nil
}

func fromHttp(link string) (*torrent.TorrentSpec, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	client := new(http.Client)
	client.Timeout = time.Duration(time.Second * 60)
	req.Header.Set("User-Agent", "DWL/1.1.1 (Torrent)")

	resp, err := client.Do(req)
	if er, ok := err.(*url.Error); ok {
		if strings.HasPrefix(er.URL, "magnet:") {
			return fromMagnet(er.URL)
		}
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	minfo, err := metainfo.Load(resp.Body)
	if err != nil {
		return nil, err
	}

	spec, err := torrent.TorrentSpecFromMetaInfoErr(minfo)
	if err != nil {
		log.TLogln("Error parse torrent info:", err)
		return nil, err
	}

	return spec, nil
}

func fromFile(path string) (*torrent.TorrentSpec, error) {
	if runtime.GOOS == "windows" && strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}

	minfo, err := metainfo.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	spec, err := torrent.TorrentSpecFromMetaInfoErr(minfo)
	if err != nil {
		log.TLogln("Error parse torrent info:", err)
		return nil, err
	}

	return spec, nil
}
