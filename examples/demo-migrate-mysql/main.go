package main

import (
	"github.com/deweppro/goppy"
	"github.com/deweppro/goppy/plugins/database"
)

func main() {

	app := goppy.New()
	app.Plugins(
		database.WithMySQL(),
	)
	app.Run()

}
