package types

import (
	"fmt"

	"go.osspkg.com/syncing"
)

const (
	globalMod = "gg"
	faceMod   = "fg"
	methodMod = "mg"
)

var (
	_storage = syncing.NewMap[string, any](10)
)

func Register[T any](module T) {
	addr := ""

	switch vv := any(module).(type) {
	case GlobalModule:
		addr = globalMod + "/" + vv.Name()
	case FaceModule:
		addr = faceMod + "/" + vv.Name()
	case MethodModule:
		addr = methodMod + "/" + vv.Name()
	default:
		panic("unknown type")
	}

	_storage.Set(addr, module)
}

func Resolve[T any](name string) (T, bool) {
	addr := ""
	nt := new(T)

	switch any(nt).(type) {
	case *GlobalModule:
		addr = globalMod + "/" + name
	case *FaceModule:
		addr = faceMod + "/" + name
	case *MethodModule:
		addr = methodMod + "/" + name
	default:
		panic(fmt.Sprintf("unknown type: %T", *nt))
	}

	raw, ok := _storage.Get(addr)
	if !ok {
		var zeroValue T
		return zeroValue, false
	}

	module, ok := raw.(T)
	return module, ok
}
