package plugin

import (
	"fmt"

	"silo/internal/torrent"
	"silo/internal/user"

	"github.com/dop251/goja"
)

func (rt *JSRuntime) createTorrentModule(mgr *torrent.Manager, userSvc *user.Service) *goja.Object {
	torrObj := rt.vm.NewObject()

	// ts.torrent.add(magnetOrHash, [options])
	torrObj.Set("add", func(call goja.FunctionCall) goja.Value {
		input := call.Argument(0).String()
		spec, err := torrent.ParseTorrentSpec(input)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("invalid torrent input: %v", err)))
		}

		// Параметры по умолчанию
		title := spec.DisplayName
		var poster, category string
		saveToDB := true
		targetUserID := "owner"

		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			opts := call.Argument(1).Export().(map[string]any)
			if t, ok := opts["title"].(string); ok && t != "" {
				title = t
			}
			if p, ok := opts["poster"].(string); ok {
				poster = p
			}
			if c, ok := opts["category"].(string); ok {
				category = c
			}
			if s, ok := opts["save_to_db"].(bool); ok {
				saveToDB = s
			}
			if u, ok := opts["user_id"].(string); ok && u != "" {
				targetUserID = u
			}
		}

		u, err := userSvc.GetUserByID(targetUserID)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("user not found: %s", targetUserID)))
		}

		status, err := mgr.AddTorrent(u, spec, title, poster, category, saveToDB)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to add torrent: %v", err)))
		}

		return rt.vm.ToValue(status)
	})

	// ts.torrent.list([userID])
	torrObj.Set("list", func(call goja.FunctionCall) goja.Value {
		targetUserID := "owner"
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Argument(0)) && !goja.IsNull(call.Argument(0)) {
			targetUserID = call.Argument(0).String()
		}

		u, err := userSvc.GetUserByID(targetUserID)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("user not found: %s", targetUserID)))
		}

		list, err := mgr.ListTorrents(u)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to list torrents: %v", err)))
		}
		return rt.vm.ToValue(list)
	})

	// ts.torrent.get(hash, [userID])
	torrObj.Set("get", func(call goja.FunctionCall) goja.Value {
		hash := call.Argument(0).String()
		targetUserID := "owner"
		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			targetUserID = call.Argument(1).String()
		}

		u, err := userSvc.GetUserByID(targetUserID)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("user not found: %s", targetUserID)))
		}

		st, err := mgr.GetTorrentStatus(u, hash)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to get torrent status: %v", err)))
		}
		return rt.vm.ToValue(st)
	})

	// ts.torrent.remove(hash, [userID])
	torrObj.Set("remove", func(call goja.FunctionCall) goja.Value {
		hash := call.Argument(0).String()
		targetUserID := "owner"
		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			targetUserID = call.Argument(1).String()
		}

		u, err := userSvc.GetUserByID(targetUserID)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("user not found: %s", targetUserID)))
		}

		if err := mgr.RemoveTorrent(u, hash); err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to remove torrent: %v", err)))
		}
		return goja.Undefined()
	})

	// ts.torrent.setViewed(hash, fileIdx, viewed, [userID])
	torrObj.Set("setViewed", func(call goja.FunctionCall) goja.Value {
		hash := call.Argument(0).String()
		fileIdx := int(call.Argument(1).ToInteger())
		viewed := call.Argument(2).ToBoolean()
		targetUserID := "owner"
		if len(call.Arguments) > 3 && !goja.IsUndefined(call.Argument(3)) && !goja.IsNull(call.Argument(3)) {
			targetUserID = call.Argument(3).String()
		}

		u, err := userSvc.GetUserByID(targetUserID)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("user not found: %s", targetUserID)))
		}

		if err := mgr.SetFileViewed(u, hash, fileIdx, viewed); err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to set viewed status: %v", err)))
		}
		return goja.Undefined()
	})

	// ts.torrent.setBlocklist(text)
	torrObj.Set("setBlocklist", func(call goja.FunctionCall) goja.Value {
		text := call.Argument(0).String()
		if err := mgr.SetBlocklistText(text); err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to set blocklist: %v", err)))
		}
		return goja.Undefined()
	})

	// ts.torrent.setTrackerPolicy({ mode: "replace", trackers: [...] })
	torrObj.Set("setTrackerPolicy", func(call goja.FunctionCall) goja.Value {
		opts := call.Argument(0).Export().(map[string]any)
		mode := torrent.TrackerMode(opts["mode"].(string))
		var trackers []string
		if tList, ok := opts["trackers"].([]any); ok {
			for _, t := range tList {
				trackers = append(trackers, fmt.Sprintf("%v", t))
			}
		}
		mgr.SetTrackerPolicy(mode, trackers)
		return goja.Undefined()
	})

	return torrObj
}
