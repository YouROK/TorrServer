package utils

import (
	"encoding/json"
	"fmt"

	"server/settings"
)

func AddInfo(hash, js string) {
	info := settings.GetInfo(hash)
	if info != "{}" {
		var jsset map[string]interface{}
		var err error
		if err = json.Unmarshal([]byte(js), &jsset); err == nil {
			var jsdb map[string]interface{}
			if err = json.Unmarshal([]byte(info), &jsdb); err == nil {
				for k, v := range jsset {
					jsdb[k] = v
				}
				jsstr, err := json.Marshal(jsdb)
				if err == nil {
					settings.AddInfo(hash, string(jsstr))
					return
				}
			}
		}
		if err != nil {
			fmt.Println(err)
		}
	}
	settings.AddInfo(hash, js)
}
