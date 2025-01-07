/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"go.osspkg.com/console"
	"go.osspkg.com/goppy/v2/internal/global"
	"go.osspkg.com/ioutils/codec"
	"go.osspkg.com/ioutils/fs"
)

var (
	rexHEAD = regexp.MustCompile(`(?mU)ref\: refs/heads/(\w+)\s+HEAD`)
	rexMOD  = regexp.MustCompile(`(?mU)module (.*)\n`)
)

type Data struct {
	Branch  string
	Repo    string
	Root    string
	Modules []string
}

func CmdGoSite() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("gosite", "Generate go pkg html")
		setter.ExecFunc(func(_ []string) {
			console.Infof("--- READ CONFIG ---")

			confpath := fs.CurrentDir() + "/.gosite.yaml"
			if !fs.FileExist(confpath) {
				console.Fatalf("File .gosite.yaml not found")
			}

			var configs []string
			result := make(map[string]*Data, 100)

			err := codec.FileEncoder(confpath).Decode(&configs)
			console.FatalIfErr(err, "Decode config")

			tempdir := fs.CurrentDir() + "/.tmp"
			defer os.RemoveAll(tempdir) // nolint: errcheck
			for _, config := range configs {
				os.RemoveAll(tempdir) // nolint: errcheck
				console.FatalIfErr(os.MkdirAll(tempdir, 0744), "Create temp dir")

				var b []byte
				b, err = global.Exec("git ls-remote --symref " + config + " HEAD")
				console.FatalIfErr(err, "Get remote HEAD")
				_strs := rexHEAD.FindStringSubmatch(string(b))
				if len(_strs) != 2 {
					console.Fatalf("HEAD not found")
				}
				HEAD := _strs[1]

				_, err = global.Exec("git clone --branch " + HEAD + " --single-branch " + config + " .tmp")
				console.FatalIfErr(err, "Clone remote HEAD")
				os.RemoveAll(tempdir + "/.git") // nolint: errcheck

				var mods map[string]*global.Module
				mods, err = global.SearchModule(tempdir)
				console.FatalIfErr(err, "Detect go.mod files")

				var dataMod *Data
				if dm, ok := result[config]; ok {
					dataMod = dm
				} else {
					dataMod = &Data{
						Branch:  HEAD,
						Repo:    strings.TrimSuffix(config, ".git"),
						Modules: make([]string, 0, 10),
					}
					result[config] = dataMod
				}

				for _, mod := range mods {
					b, err = os.ReadFile(mod.File)
					console.FatalIfErr(err, "Read go.mod [%s]", mod.File)
					_strs = rexMOD.FindStringSubmatch(string(b))
					if len(_strs) != 2 {
						console.Fatalf("Module not found in %s", mod.File)
					}
					module := _strs[1]
					dataMod.Modules = append(dataMod.Modules, module)
				}
				for i, module := range dataMod.Modules {
					if i == 0 {
						dataMod.Root = module
						continue
					}
					if len(dataMod.Root) > len(module) {
						dataMod.Root = module
					}
				}

				if len(dataMod.Modules) == 0 {
					delete(result, config)
				}
			}

			index := make(map[string][]string)
			for _, data := range result {
				var u *url.URL
				u, err = url.Parse("http://" + data.Root)
				console.FatalIfErr(err, "Decode module url [%s]", data.Root)
				domain := u.Host
				if _, ok := index[domain]; !ok {
					index[domain] = make([]string, 0, 10)
				}

				sort.Strings(data.Modules)
				for _, mod := range data.Modules {
					repl := strings.NewReplacer(
						"{%module%}", mod,
						"{%root%}", data.Root,
						"{%repo%}", data.Repo,
						"{%head%}", data.Branch,
					)
					err = os.MkdirAll(mod, 0744)
					console.FatalIfErr(err, "Create site dir [%s]", mod)
					index[domain] = append(index[domain], mod)

					tmpl := repl.Replace(htmlPageTemplate)
					err = os.WriteFile(mod+"/index.html", []byte(tmpl), 0664)
					console.FatalIfErr(err, "Write HTML [%s]", mod+"/index.html")
				}
			}
			for domain, links := range index {
				sort.Strings(links)

				linksHtml := ""
				for _, link := range links {
					linkName := strings.TrimPrefix(link, domain)
					linkName = strings.Trim(linkName, "/")
					linksHtml += "\n<li><a href=\"//" + link + "\">" + linkName + "</a></li>"
				}

				repl := strings.NewReplacer(
					"{%domain%}", domain,
					"{%links%}", linksHtml,
				)

				tmpl := repl.Replace(htmlIndexPage)
				err = os.WriteFile(domain+"/index.html", []byte(tmpl), 0664)
				console.FatalIfErr(err, "Write HTML [%s]", domain+"/index.html")
			}

		})
	})
}

const (
	htmlPageTemplate = `
<!DOCTYPE html>
<html lang="en" dir="ltr">

<head>
    <title>{%module%}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, height=device-height, minimum-scale=1.0, initial-scale=1.0">
    <meta name="go-import" content="{%root%} git {%repo%}">
    <meta name="go-source" content="{%root%} {%repo%} {%repo%}/tree/{%head%}{/dir} {%repo%}/tree/{%head%}{/dir}/{file}#L{line}">
</head>

<body>
    <aside>
        <a href="/">Back Home</a>
    </aside>
    <hr>
    <div>
        <h1>{%module%}</h1>
    </div>

    <div>
        <b>Install command:</b>
        <pre>go get {%module%}</pre>
    </div>

    <div>
        <b>Import in source code:</b>
        <pre>import "{%module%}"</pre>
    </div>
        
    <div>
        <b>Repository:</b>
        <a href="{%repo%}">{%repo%}</a>
    </div>
</body>

</html>
`
	htmlIndexPage = `
<!DOCTYPE html>
<html lang="en" dir="ltr">

<head>
    <title>{%domain%}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, height=device-height, minimum-scale=1.0, initial-scale=1.0">
</head>

<body>
    <div>
        <h1>{%domain%}</h1>
    </div>
	<hr>
    <aside>
        <ul>
            {%links%}
        </ul>
    </aside>
</body>

</html>
`
)
