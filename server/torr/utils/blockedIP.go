package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"server/log"
	"strings"

	"server/settings"

	"github.com/anacrolix/torrent/iplist"
)

func ReadBlockedIP() (ranger iplist.Ranger, err error) {
	buf, err := os.ReadFile(filepath.Join(settings.Path, "blocklist"))
	if err != nil {
		return nil, err
	}
	log.TLogln("Read block list...")
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
		log.TLogln("Readed ranges:", len(ranges))
	}
	return
}
