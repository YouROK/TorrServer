package helpers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"server/settings"

	"github.com/anacrolix/torrent/metainfo"
)

func GetMagnet(link string) (*metainfo.Magnet, []byte, error) {
	url, err := url.Parse(link)
	if err != nil {
		return nil, nil, err
	}

	var mag *metainfo.Magnet
	var infoBytes []byte
	switch strings.ToLower(url.Scheme) {
	case "magnet":
		mag, err = getMag(url.String())
		if err == nil {
			torDb, err := settings.LoadTorrentDB(mag.InfoHash.HexString())
			if err == nil && torDb != nil {
				infoBytes = torDb.InfoBytes
			}
		}
	case "http", "https":
		mag, infoBytes, err = getMagFromHttp(url.String())
	case "":
		mag, err = getMag("magnet:?xt=urn:btih:" + url.Path)
	case "file":
		mag, infoBytes, err = getMagFromFile(url.Path)
	default:
		err = fmt.Errorf("unknown scheme:", url, url.Scheme)
	}
	if err != nil {
		return nil, nil, err
	}

	return mag, infoBytes, nil
}

func getMag(link string) (*metainfo.Magnet, error) {
	mag, err := metainfo.ParseMagnetUri(link)
	return &mag, err
}

func getMagFromHttp(url string) (*metainfo.Magnet, []byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	client := new(http.Client)
	client.Timeout = time.Duration(time.Second * 30)
	req.Header.Set("User-Agent", "DWL/1.1.1 (Torrent)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, nil, errors.New(resp.Status)
	}

	minfo, err := metainfo.Load(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, nil, err
	}
	mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	return &mag, minfo.InfoBytes, nil
}

func getMagFromFile(path string) (*metainfo.Magnet, []byte, error) {
	if runtime.GOOS == "windows" && strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	minfo, err := metainfo.LoadFromFile(path)
	if err != nil {
		return nil, nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, nil, err
	}

	mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	return &mag, minfo.InfoBytes, nil
}
