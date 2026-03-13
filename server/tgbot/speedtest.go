package tgbot

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
)

func cmdSpeedtest(c tele.Context) error {
	args := c.Args()
	size := 10
	if len(args) > 0 {
		if s, err := strconv.Atoi(strings.TrimSpace(args[0])); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	host := getHost()
	url := fmt.Sprintf("%s/download/%d", host, size)
	return c.Send(fmt.Sprintf(tr(c.Sender().ID, "speedtest_msg"), size, url))
}
