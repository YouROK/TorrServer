package tgbot

import (
	"strconv"

	tele "gopkg.in/telebot.v4"
	up "server/tgbot/upload"
)

func upload(c tele.Context) error {
	args := c.Args()
	idstr := args[2]
	id, err := strconv.Atoi(idstr)
	if err != nil {
		return err
	}
	up.AddRange(c, args[1], id, id)
	return nil
}

func uploadall(c tele.Context) {
	args := c.Args()
	up.AddRange(c, args[1], 1, -1)
}
