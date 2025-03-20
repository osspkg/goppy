/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"go.osspkg.com/console"
	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v2/internal/global"
)

func CmdGenerate() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("generate", "Generate new app with goppy sdk")
		setter.ExecFunc(func(_ []string) {
			currdir := fs.CurrentDir()

			data := make(map[string]any, 100)
			data["go_version"] = strings.TrimLeft(global.GoVersion(), "go")
			data["app_module"] = console.Input("Input project name", nil, "app")
			data["app_name"] = func() string {
				vv := strings.Split(data["app_module"].(string), "/") //nolint: errcheck
				return vv[len(vv)-1]
			}()

			for _, blocks := range modules {
				for _, name := range blocks {
					data["mod_"+name] = false
				}
			}

			userInput("Add modules", modules, "q", func(s string) {
				data["mod_"+s] = true
			})

			for _, folder := range folders {
				console.FatalIfErr(os.MkdirAll(currdir+"/"+folder, 0744), "Create folder")
			}

			for filename, tmpl := range generateTemplates {
				if strings.Contains(filename, "{{") {
					for key, value := range data {
						filename = strings.ReplaceAll(
							filename,
							"{{"+key+"}}",
							fmt.Sprintf("%+v", value),
						)
					}
				}
				writeFile(currdir+"/"+filename, tmpl, data)
			}

			global.ExecPack("gofmt -w .", "go mod tidy -v", "goppy setup-app")
		})
	})
}

var modules = [][]string{
	{
		"metrics",
		"geoip",
		"oauth",
		"auth_jwt",
	},
	{
		"db_mysql",
		"db_sqlite",
		"db_postgre",
	},
	{
		"web_server",
		"web_client",
	},
	{
		"websocket_server",
		"websocket_client",
	},
	{
		"dns_server",
		"dns_client",
	},
}

var folders = []string{
	"app",
	"config",
	"cmd",
	"pkg",
}

var generateTemplates = map[string]string{
	".gitignore":               tmplGitIgnore,
	"README.md":                tmplReadMe,
	"go.mod":                   tmplGoMod,
	"docker-compose.yaml":      tmplDockerFile,
	"cmd/{{app_name}}/main.go": tmplMainGO,
	"app/plugin.go":            tmplAppGo,
	"pkg/plugin.go":            tmplPkgGo,
}

func writeFile(filename, t string, data map[string]any) {
	tmpl, err := template.New("bt").Parse(t)
	console.FatalIfErr(err, "Parse template")
	tmpl.Option("missingkey=error")

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	console.FatalIfErr(err, "Build template")

	console.FatalIfErr(os.MkdirAll(filepath.Dir(filename), 0744), "Create folder")
	console.FatalIfErr(os.WriteFile(filename, buf.Bytes(), 0664), "Write %s", filename)
}

func userInput(msg string, mods [][]string, exit string, call func(s string)) {
	fmt.Printf("--- %s ---\n", msg)

	list := make(map[string]string, len(mods)*4)
	i := 0
	for _, blocks := range mods {
		for _, name := range blocks {
			i++
			fmt.Printf("(%d) %s, ", i, name)
			list[fmt.Sprintf("%d", i)] = name
		}
		fmt.Printf("\n")
	}
	fmt.Printf("and (%s) Done: \n", exit)

	scan := bufio.NewScanner(os.Stdin)
	for {
		if scan.Scan() {
			r := scan.Text()
			if r == exit {
				fmt.Printf("\u001B[1A\u001B[K--- Done ---\n\n")
				return
			}
			if name, ok := list[r]; ok {
				call(name)
				fmt.Printf("\033[1A\033[K + %s\n", name)
				continue
			}
			fmt.Printf("\u001B[1A\u001B[KBad answer! Try again: ")
		}
	}
}

const tmplMainGO = `package main

import (
	app "{{.app_module}}/app"
	pkg "{{.app_module}}/pkg"

	"go.osspkg.com/goppy"
	{{if .mod_metrics}}"go.osspkg.com/goppy/v2/metrics"
{{end}}{{if .mod_geoip}}"go.osspkg.com/goppy/v2/geoip"
{{end}}{{if or .mod_oauth .mod_auth_jwt}}"go.osspkg.com/goppy/v2/auth"
{{end}}{{if .mod_db_mysql}}"go.osspkg.com/goppy/v2/ormmysql"
{{end}}{{if .mod_db_sqlite}}"go.osspkg.com/goppy/v2/ormsqlite"
{{end}}{{if .mod_db_postgre}}"go.osspkg.com/goppy/v2/ormpgsql"
{{end}}{{if or .mod_web_server .mod_web_client}}"go.osspkg.com/goppy/v2/web"
{{end}}{{if or .mod_websocket_server .mod_websocket_client}}"go.osspkg.com/goppy/v2/ws"
{{end}}{{if or .mod_dns_server .mod_dns_client}}"go.osspkg.com/goppy/v2/xdns"
{{end}}
)

var Version = "v0.0.0-dev"

func main() {
	gop := goppy.New()
	gop.AppName("{{.app_name}}")
	gop.AppVersion(Version)
	gop.Plugins(
		{{if .mod_metrics}}metrics.WithServer(),{{end}}
		{{if .mod_geoip}}geoip.WithMaxMindGeoIP(),{{end}}
		{{if .mod_oauth}}auth.WithOAuth(),{{end}}
		{{if .mod_auth_jwt}}auth.WithJWT(),{{end}}
		{{if .mod_db_mysql}}ormmysql.WithClient(),{{end}}
		{{if .mod_db_sqlite}}ormsqlite.WithClient(),{{end}}
		{{if .mod_db_postgre}}ormpgsql.WithClient(),{{end}}
		{{if .mod_web_server}}web.WithServer(),{{end}}
		{{if .mod_web_client}}web.WithClient(),{{end}}
		{{if .mod_websocket_server}}ws.WithServer(),{{end}}
		{{if .mod_websocket_client}}ws.WithClient(),{{end}}
		{{if .mod_dns_server}}xdns.WithServer(),{{end}}
		{{if .mod_dns_client}}xdns.WithClient(),{{end}}
	)
	gop.Plugins(app.Plugins...)
	gop.Plugins(pkg.Plugins...)
	gop.Run()
}
`

const tmplAppGo = `package app

import (
	"go.osspkg.com/goppy/v2/plugins"
)

var Plugins = plugins.Inject()

`

const tmplPkgGo = `package pkg

import (
	"go.osspkg.com/goppy/v2/plugins"
)

var Plugins = plugins.Inject()

`

const tmplReadMe = `# {{.app_name}}

`

const tmplGitIgnore = `
*.exe
*.exe~
*.dll
*.so
*.dylib
*.db
*.db-journal
*.mmdb
*.test
*.out

.idea/
.vscode/
.tools/

coverage.txt
coverage.out

bin/
vendor/
build/

`

const tmplDockerFile = `version: '2.4'

networks:
  database:
    name: {{.app_name}}-dev-net

services:

  db:
    image: library/mysql:5.7.25
    restart: on-failure
    environment:
      MYSQL_ROOT_PASSWORD: 'root'
      MYSQL_USER: 'test'
      MYSQL_PASSWORD: 'test'
      MYSQL_DATABASE: 'test_database'
    healthcheck:
      test: [ "CMD", "mysql", "--user=root", "--password=root", "-e", "SHOW DATABASES;" ]
      interval: 15s
      timeout: 30s
      retries: 30
    ports:
      - "127.0.0.1:3306:3306"
    networks:
      - database

  adminer:
    image: adminer:latest
    restart: on-failure
    links:
      - db
    ports:
      - "127.0.0.1:8000:8080"
    networks:
      - database
`

const tmplGoMod = `module {{.app_module}}

go {{.go_version}}
`
