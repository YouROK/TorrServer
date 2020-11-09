package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func ParseLink(link string) (*torrent.TorrentSpec, error) {
	urlLink, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(urlLink.Scheme) {
	case "magnet":
		return fromMagnet(urlLink.String())
	case "http", "https":
		return fromHttp(urlLink.String())
	case "":
		return fromMagnet("magnet:?xt=urn:btih:" + urlLink.Path)
	case "file":
		return fromFile(urlLink.Path)
	default:
		err = fmt.Errorf("unknown scheme:", urlLink, urlLink.Scheme)
	}
	return nil, err
}

func fromMagnet(link string) (*torrent.TorrentSpec, error) {
	mag, err := metainfo.ParseMagnetURI(link)
	if err != nil {
		return nil, err
	}

	return &torrent.TorrentSpec{
		InfoBytes:   nil,
		Trackers:    [][]string{mag.Trackers},
		DisplayName: mag.DisplayName,
		InfoHash:    mag.InfoHash,
	}, nil
}

func fromHttp(url string) (*torrent.TorrentSpec, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := new(http.Client)
	client.Timeout = time.Duration(time.Second * 30)
	req.Header.Set("User-Agent", "DWL/1.1.1 (Torrent)")

	resp, err := client.Do(req)
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
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, err
	}
	mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())

	return &torrent.TorrentSpec{
		InfoBytes:   minfo.InfoBytes,
		Trackers:    [][]string{mag.Trackers},
		DisplayName: info.Name,
		InfoHash:    minfo.HashInfoBytes(),
	}, nil
}

func fromFile(path string) (*torrent.TorrentSpec, error) {
	if runtime.GOOS == "windows" && strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	minfo, err := metainfo.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, err
	}

	mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	return &torrent.TorrentSpec{
		InfoBytes:   minfo.InfoBytes,
		Trackers:    [][]string{mag.Trackers},
		DisplayName: info.Name,
		InfoHash:    minfo.HashInfoBytes(),
	}, nil
}
