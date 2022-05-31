package main

import (
	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/plugins"
)

func main() {

	app := goppy.New()
	//app.WithConfig("./config.yaml")
	app.Plugins(
		plugins.WithHTTPDebug(),
		plugins.WithHTTP(),
	)
	app.Run()

}
