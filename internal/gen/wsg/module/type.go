package module

import (
	"go.osspkg.com/gogen/golang"
	"go.osspkg.com/goppy/v3/internal/gen/wsg/types"
)

type Module interface {
	Name() string
	CreateObject(object types.Object) *golang.Tokens
	CreateMethod(object types.Method) *golang.Tokens
}
