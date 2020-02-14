package utils

import (
	"bufio"
	"io/ioutil"
	"path/filepath"
	"strings"

	"server/settings"

	"github.com/anacrolix/torrent/iplist"
)

func ReadBlockedIP() (ranger iplist.Ranger, err error) {
	buf, err := ioutil.ReadFile(filepath.Join(settings.Path, "blocklist"))
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(buf)))
	var ranges []iplist.Range
	for scanner.Scan() {
		r, ok, err := iplist.ParseBlocklistP2PLine(scanner.Bytes())
		if err != nil {
			return nil, err
		}
		if ok {
			ranges = append(ranges, r)
		}
	}
	err = scanner.Err()
	if len(ranges) > 0 {
		ranger = iplist.New(ranges)
	}
	return
}
