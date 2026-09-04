package tgbot

import tele "gopkg.in/telebot.v4"

func cmdCancel(c tele.Context) error {
	uid := c.Sender().ID
	cleared := clearPendingForUser(uid)
	if cleared {
		return sendWithMenu(c, tr(uid, "canceled"))
	}
	return sendWithMenu(c, tr(uid, "cancel_nothing"))
}

func clearPendingForUser(uid int64) bool {
	cleared := false
	pendingInputMu.Lock()
	for key, p := range pendingInputs {
		if p.UserID == uid {
			delete(pendingInputs, key)
			cleared = true
		}
	}
	pendingInputMu.Unlock()

	pendingPresetMu.Lock()
	for key, p := range pendingPresets {
		if p.UserID == uid {
			delete(pendingPresets, key)
			cleared = true
		}
	}
	pendingPresetMu.Unlock()

	if clearPendingSearch(uid) {
		cleared = true
	}
	return cleared
}
