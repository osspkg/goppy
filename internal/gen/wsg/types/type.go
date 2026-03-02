package types

import "go.osspkg.com/gogen/types"

const (
	IFace   = "iface."
	Methods = "method."
)

type Module interface {
	Name() string
	Build(ctx Ctx) error
}

type Writer interface {
	WriteFile(fileName string, tok types.Token) error
}

type Ctx struct {
	W       Writer
	PkgName string
	File    File
	Object  Object
	Tags    Tags
}
