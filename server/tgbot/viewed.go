package tgbot

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
	sets "server/settings"
)

func cmdViewed(c tele.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return c.Send(tr(c.Sender().ID, "viewed_usage"))
	}

	action := strings.ToLower(args[0])
	if action == "set" || action == "rem" {
		if len(args) < 2 {
			return c.Send(fmt.Sprintf(tr(c.Sender().ID, "viewed_usage_action"), action))
		}
		hash := resolveHash(c, args[1])
		if hash == "" {
			return c.Send(tr(c.Sender().ID, "invalid_hash"))
		}
		if action == "set" {
			if len(args) < 3 {
				return c.Send(tr(c.Sender().ID, "viewed_usage_set"))
			}
			index, err := strconv.Atoi(args[2])
			if err != nil || index < 1 {
				return c.Send(tr(c.Sender().ID, "viewed_file_index"))
			}
			sets.SetViewed(&sets.Viewed{Hash: hash, FileIndex: index})
			return c.Send(fmt.Sprintf(tr(c.Sender().ID, "viewed_marked"), hash, index))
		}
		index := -1
		if len(args) >= 3 {
			if i, err := strconv.Atoi(args[2]); err == nil && i >= 1 {
				index = i
			}
		}
		sets.RemViewed(&sets.Viewed{Hash: hash, FileIndex: index})
		if index >= 1 {
			return c.Send(fmt.Sprintf(tr(c.Sender().ID, "viewed_unmarked"), hash, index))
		}
		return c.Send(fmt.Sprintf(tr(c.Sender().ID, "viewed_cleared"), hash))
	}

	hash := resolveHash(c, args[0])
	if hash == "" {
		return c.Send(tr(c.Sender().ID, "viewed_usage"))
	}

	list := sets.ListViewed(hash)
	if len(list) == 0 {
		return c.Send(tr(c.Sender().ID, "viewed_empty"))
	}

	var sb strings.Builder
	sb.WriteString("<b>" + tr(c.Sender().ID, "viewed_list") + "</b>\n\n")
	fmt.Fprintf(&sb, "<code>%s</code>\n\n", hash)
	for _, v := range list {
		fmt.Fprintf(&sb, "  #%d\n", v.FileIndex)
	}
	return c.Send(sb.String())
}
