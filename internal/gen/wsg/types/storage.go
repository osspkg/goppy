package types

import "go.osspkg.com/syncing"

var _storage = syncing.NewMap[string, Module](10)

func Register(module Module) {
	if _, ok := _storage.Get(module.Name()); ok {
		panic("duplicate register module: " + module.Name())
	}
	_storage.Set(module.Name(), module)
}

func Resolve(name string) (Module, bool) {
	return _storage.Get(name)
}
