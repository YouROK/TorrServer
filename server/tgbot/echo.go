package tgbot

import (
	"fmt"

	tele "gopkg.in/telebot.v4"
	"server/version"
)

func cmdEcho(c tele.Context) error {
	v := version.Version
	if v == "" {
		v = "unknown"
	}
	return c.Send(fmt.Sprintf("🔄 TorrServer %s", v))
}
