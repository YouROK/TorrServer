package tgbot

import tele "gopkg.in/telebot.v4"

// adminOnly wraps a handler to allow only admin users (when whitelist is used)
func adminOnly(h tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Sender() == nil {
			return nil
		}
		if !isAdmin(c.Sender().ID) {
			return c.Send(tr(c.Sender().ID, "admin_only"))
		}
		return h(c)
	}
}
