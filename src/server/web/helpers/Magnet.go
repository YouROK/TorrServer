package helpers

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/anacrolix/torrent/metainfo"
)

func GetMagnet(link string) (*metainfo.Magnet, error) {
	url, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	var mag *metainfo.Magnet
	switch strings.ToLower(url.Scheme) {
	case "magnet":
		mag, err = getMag(url.String())
	case "http", "https":
		mag, err = getMagFromHttp(url.String())
	case "":
		mag, err = getMag("magnet:?xt=urn:btih:" + url.Path)
	default:
		mag, err = getMagFromFile(url.Path)
	}
	if err != nil {
		return nil, err
	}

	return mag, nil
}

func getMag(link string) (*metainfo.Magnet, error) {
	mag, err := metainfo.ParseMagnetURI(link)
	return &mag, err
}

func getMagFromHttp(url string) (*metainfo.Magnet, error) {
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
	return &mag, nil
}

func getMagFromFile(path string) (*metainfo.Magnet, error) {

	minfo, err := metainfo.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, err
	}

	mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	return &mag, nil
}
