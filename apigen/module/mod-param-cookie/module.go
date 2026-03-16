package mod_param_cookie

import (
	"fmt"

	. "go.osspkg.com/gogen/golang"
	at "go.osspkg.com/goppy/v3/apigen/types"
)

type Module struct{}

func (Module) Name() string {
	return "cookie"
}

func (v Module) Build(w at.Joiner, m at.ParamMeta, value at.Param) error {
	switch m.Type {
	case at.ParamIn:
		return v.generateIn(w, m, value)
	case at.ParamOut:
		return v.generateOut(w, m, value)
	default:
		return fmt.Errorf("unknown type")
	}
}

func (v Module) generateIn(w at.Joiner, m at.ParamMeta, p at.Param) error {
	m.Import.Set("web", "go.osspkg.com/goppy/v3/web")

	w.Join(
		List(ID(m.CodeName), Raw("_")).Op("=").
			ID("web").Op(".").ID("StrTo").Raw("[").Pkg(p.Pkg).ID(p.Type).Raw("]").
			Call(
				ID("webCtx").Op(".").ID("Cookie").Call().Op(".").ID("Get").Call(Text(m.Value)),
			),
	)

	return nil
}

func (v Module) generateOut(w at.Joiner, m at.ParamMeta, _ at.Param) error {
	m.Import.Set("nethttp", "net/http")
	m.Import.Set("fmt", "fmt")
	m.Import.Set("time", "time")

	w.Join(
		If().ID("err").Op("!=").Nil().Block(Return()),
		ID("webCtx").Op(".").ID("Cookie").Call().Op(".").ID("Set").Bracket(
			Op("&").Pkg("nethttp").ID("Cookie").Block(
				ID("Name").Op(":").Text(m.Value).Op(","),
				ID("Value").Op(":").Pkg("fmt").ID("Sprintf").Bracket(
					List(Text("%v"), Raw(m.CodeName)),
				).Op(","),
				ID("Path").Op(":").Text(m.Args.Get("path", "/")).Op(","),
				ID("Expires").Op(":").Pkg("time").ID("Now").Call().Op(".").ID("Add").Call(
					Raw(m.Args.Get("time", "86400")).Op("*").Pkg("time").ID("Second"),
				).Op(","),
				ID("Secure").Op(":").Raw(m.Args.Get("secure", "true")).Op(","),
				ID("HttpOnly").Op(":").Raw(m.Args.Get("httpOnly", "true")).Op(","),
				ID("SameSite").Op(":").Pkg("nethttp").ID(m.Args.Get("sameSite", "SameSiteStrictMode")).Op(","),
			),
		),
	)
	return nil
}
