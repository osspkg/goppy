package main

import (
	"github.com/osspkg/goppy"
	"github.com/osspkg/goppy/plugins/database"
)

func main() {

	app := goppy.New()
	app.Plugins(
		database.WithMySQL(),
	)
	app.Run()

}
