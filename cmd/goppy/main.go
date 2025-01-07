/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/console"
	"go.osspkg.com/goppy/v2/internal/commands"
	"go.osspkg.com/goppy/v2/internal/global"
)

func main() {
	console.ShowDebug(true)
	app := console.New("goppy", "Goppy SDK Development Tool")

	global.SetupEnv()

	app.AddCommand(
		commands.CmdLicense(),
		commands.CmdLint(),
		commands.CmdTest(),
		commands.CmdBuild(),
		commands.CmdSetupLib(),
		commands.CmdSetupApp(),
		commands.CmdGoSite(),
		// commands.CmdGenerate(),
	)

	app.Exec()
}
