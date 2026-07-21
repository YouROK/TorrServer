package tgbot

import (
	"fmt"

	"server/version"

	tele "gopkg.in/telebot.v4"
)

func cmdEcho(c tele.Context) error {
	v := version.Version
	if v == "" {
		v = "unknown"
	}
	return c.Send(fmt.Sprintf("🔄 TorrServer %s", v))
}
