package plugin

import (
	"fmt"

	"silo/internal/database"

	"github.com/dop251/goja"
	bolt "go.etcd.io/bbolt"
)

func (rt *JSRuntime) createStorageModule(db *database.DB) *goja.Object {
	storeObj := rt.vm.NewObject()
	prefix := rt.pluginID + ":"

	storeObj.Set("set", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		val := call.Argument(1).String()

		err := db.GetRawConn().Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(database.BucketPluginData)
			return b.Put([]byte(prefix+key), []byte(val))
		})

		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to save data: %v", err)))
		}
		return goja.Undefined()
	})

	storeObj.Set("get", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		var result string

		_ = db.GetRawConn().View(func(tx *bolt.Tx) error {
			b := tx.Bucket(database.BucketPluginData)
			data := b.Get([]byte(prefix + key))
			if data != nil {
				result = string(data)
			}
			return nil
		})

		if result == "" {
			return goja.Undefined()
		}
		return rt.vm.ToValue(result)
	})

	return storeObj
}
