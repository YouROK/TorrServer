package plugin

import (
	"github.com/dop251/goja"
)

// createI18nModule — объект ts.i18n: регистрация переводов от плагина
func (rt *JSRuntime) createI18nModule() *goja.Object {
	obj := rt.vm.NewObject()

	// ts.i18n.add(lang, key, value) — добавить одну строку
	obj.Set("add", func(call goja.FunctionCall) goja.Value {
		lang := call.Argument(0).String()
		key := call.Argument(1).String()
		value := call.Argument(2).String()
		rt.i18n.Add(rt.pluginID, lang, key, value)
		return goja.Undefined()
	})

	// ts.i18n.addMany(lang, {key: value, ...}) — добавить пачку строк
	obj.Set("addMany", func(call goja.FunctionCall) goja.Value {
		lang := call.Argument(0).String()
		var kv map[string]string
		if err := rt.vm.ExportTo(call.Argument(1), &kv); err != nil {
			return goja.Undefined()
		}
		rt.i18n.AddMany(rt.pluginID, lang, kv)
		return goja.Undefined()
	})

	// ts.i18n.lang(code, name) — объявить отображаемое имя языка
	obj.Set("lang", func(call goja.FunctionCall) goja.Value {
		code := call.Argument(0).String()
		name := call.Argument(1).String()
		rt.i18n.SetLangName(code, name)
		return goja.Undefined()
	})

	return obj
}
