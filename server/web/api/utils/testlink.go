package utils

import (
	"fmt"
	"net/url"
)

func TestLink(link string, auth bool) error {
	link, err := url.QueryUnescape(link)
	if err != nil {
		return err
	}
	ur, err := url.Parse(link)
	if err != nil {
		return err
	}

	if ur.Scheme == "magnet" || ur.Scheme == "torrs" || ur.Scheme == "" {
		return nil
	}

	if !auth {
		return fmt.Errorf("auth required")
	}

	return nil
}
