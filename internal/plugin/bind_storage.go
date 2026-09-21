package plugin

import (
	"bytes"
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

	storeObj.Set("rem", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()

		err := db.GetRawConn().Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(database.BucketPluginData)
			if b == nil {
				return nil
			}
			return b.Delete([]byte(prefix + key))
		})

		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to remove data: %v", err)))
		}
		return goja.Undefined()
	})

	storeObj.Set("clear", func(call goja.FunctionCall) goja.Value {
		err := db.GetRawConn().Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(database.BucketPluginData)
			if b == nil {
				return nil
			}

			c := b.Cursor()
			prefixBytes := []byte(prefix)

			var keysToDelete [][]byte
			for k, _ := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, _ = c.Next() {
				keyCopy := make([]byte, len(k))
				copy(keyCopy, k)
				keysToDelete = append(keysToDelete, keyCopy)
			}

			for _, key := range keysToDelete {
				if err := b.Delete(key); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to clear data: %v", err)))
		}
		return goja.Undefined()
	})

	return storeObj
}
