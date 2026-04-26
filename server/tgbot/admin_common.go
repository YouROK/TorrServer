package tgbot

import (
	"server/tgbot/config"
)

func isAdmin(userID int64) bool {
	if len(config.Cfg.WhiteIds) == 0 {
		return false
	}
	for _, id := range config.Cfg.WhiteIds {
		if id == userID {
			return true
		}
	}
	return false
}
