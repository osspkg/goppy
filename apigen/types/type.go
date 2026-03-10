package types

import "go.osspkg.com/gogen/types"

type GlobalModule interface {
	Name() string
	Build(w Writer, m GlobalMeta, value []File) error
}

type FaceModule interface {
	Name() string
	Build(w Writer, m FaceMeta, value File) error
}

type MethodModule interface {
	Name() string
	Build(w Writer, m MethodMeta, value Method) error
}

type Writer interface {
	WriteFile(fileName string, tok types.Token) error
}

type GlobalMeta struct {
	PkgName string
}

type FaceMeta struct {
	PkgName string
}

type MethodMeta struct {
	PkgName string
}
