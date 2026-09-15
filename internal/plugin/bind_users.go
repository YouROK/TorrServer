package plugin

import (
	"fmt"

	"silo/internal/user"

	"github.com/dop251/goja"
)

func (rt *JSRuntime) createUsersModule(userSvc *user.Service) *goja.Object {
	usersObj := rt.vm.NewObject()

	// ts.users.list() -> массив пользователей
	usersObj.Set("list", func(call goja.FunctionCall) goja.Value {
		users, err := userSvc.ListUsers()
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to list users: %v", err)))
		}

		var result []map[string]any
		for _, u := range users {
			result = append(result, map[string]any{
				"id":         u.ID,
				"username":   u.Username,
				"rank":       int(u.Rank),
				"is_banned":  u.IsBanned,
				"created_at": u.CreatedAt.Unix(),
			})
		}
		return rt.vm.ToValue(result)
	})

	// ts.users.get(userID)
	usersObj.Set("get", func(call goja.FunctionCall) goja.Value {
		userID := call.Argument(0).String()
		u, err := userSvc.GetUserByID(userID)
		if err != nil {
			return goja.Undefined()
		}

		return rt.vm.ToValue(map[string]any{
			"id":         u.ID,
			"username":   u.Username,
			"rank":       int(u.Rank),
			"is_banned":  u.IsBanned,
			"created_at": u.CreatedAt.Unix(),
		})
	})

	return usersObj
}
