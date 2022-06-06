# Goppy Microservice Toolkit 

[![GoDoc](https://godoc.org/github.com/dewep-online/goppy?status.svg)](https://godoc.org/github.com/dewep-online/goppy) 
[![Coverage Status](https://coveralls.io/repos/github/dewep-online/goppy/badge.svg?branch=master)](https://coveralls.io/github/dewep-online/goppy?branch=master) 
[![Release](https://img.shields.io/github/release/dewep-online/goppy.svg?style=flat-square)](https://github.com/dewep-online/goppy/releases/latest) 
[![Go Report Card](https://goreportcard.com/badge/github.com/dewep-online/goppy)](https://goreportcard.com/report/github.com/dewep-online/goppy) 
[![CI](https://github.com/dewep-online/goppy/actions/workflows/ci.yml/badge.svg)](https://github.com/dewep-online/goppy/actions/workflows/ci.yml)

## Installation

```bash
go get -u github.com/dewep-online/goppy
```

## Features

- Custom pool of HTTP servers with configuration via config
- Group APIs with middleware hanging on each group
- Extensible middleware framework
- Application customization via plugins
- Built-in dependency container
- Data binding for JSON

## Plugins

| Plugin       |Comment| Import                  |
|--------------|---|-------------------------|
| **debug**    |profiling application (pprof) with HTTP access.| `http.WithHTTPDebug()`  |
| **http**     |Out of the box multi-server launch of web servers with separate routing. Grouping of routers with connection to a group of dedicated middleware.| `http.WithHTTP()`       |
| **database** |Multi connection pools with MySQL and SQLite databases (with initialization migration setup).| `database.WithMySQL()` `database.WithSQLite()` |

## Quick Start

Config:

```yaml
env: dev
pid: ""
level: 4
log: /dev/stdout

debug:
    addr: 127.0.0.1:12000

http:
    main:
        addr: 127.0.0.1:8080
```

Code:

```go
package main

import (
	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/http"
)

func main() {

	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		http.WithHTTPDebug(),
		http.WithHTTP(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes *http.RouterPool, c *Controller) {
				router := routes.Main()
				router.Use(http.ThrottlingMiddleware(100))
				router.Get("/users", c.Users)

				api := router.Collection("/api/v1", http.ThrottlingMiddleware(100))
				api.Get("/user/{id}", c.User)
			},
		},
	)
	app.Run()

}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Users(ctx http.Ctx) {
	data := []int64{1, 2, 3, 4}
	ctx.SetBody().JSON(data)
}

func (v *Controller) User(ctx http.Ctx) {
	id, _ := ctx.Param("id").Int()

	ctx.SetBody().Error(http.ErrMessage{
		HTTPCode:     400,
		InternalCode: "x1000",
		Message:      "user not found",
		Ctx:          map[string]interface{}{"id": id},
	})

	ctx.Log().Infof("user - %d", id)
}

```

## Contribute

**Use issues for everything**

- For a small change, just send a PR.
- For bigger changes open an issue for discussion before sending a PR.
- PR should have:
  - Test case
  - Documentation
  - Example (If it makes sense)
- You can also contribute by:
  - Reporting issues
  - Suggesting new features or enhancements
  - Improve/fix documentation

## Community

- [Forum](https://github.com/dewep-online/goppy/discussions)

## License

BSD-3-Clause License. See the LICENSE file for details.
