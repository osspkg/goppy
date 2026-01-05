# Goppy Microservice Toolkit

[![Release](https://img.shields.io/github/release/osspkg/goppy.svg?style=flat-square)](https://github.com/osspkg/goppy/releases/latest)
![GitHub](https://img.shields.io/github/license/osspkg/goppy)
[![Forum](https://img.shields.io/badge/community-forum-red)](https://github.com/osspkg/goppy/discussions)

## Installation

```bash
go get -u go.osspkg.com/goppy/v3
```

## Features

- Config auto generation
- Custom pool of HTTP servers with configuration via config
- Group APIs with middleware hanging on each group
- Extensible middleware framework
- Application customization via plugins
- Built-in dependency container
- Data binding for JSON
- Command support
- Database support and automatic migration

## Quick Start

Config:

```yaml
env: dev

log:
  file_path: /dev/stdout
  format: string # json, string, syslog
  level: 4 # 0-Fatal, 1-Error, 2-Warning, 3-Info, 4-Debug

http:
  - tag: main
    addr: 0.0.0.0:10000
```

Code:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"go.osspkg.com/logx"

	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/metrics"
	"go.osspkg.com/goppy/v3/plugins"
	"go.osspkg.com/goppy/v3/web"
)

func main() {
	// Specify the path to the config via the argument: `--config`.
	// Specify the path to the pidfile via the argument: `--pid`.
	app := goppy.New("app_name", "v1.0.0", "app description")
	app.Plugins(
		metrics.WithServer(),
		web.WithServer(),
	)
	app.Plugins(
		plugins.Kind{
			Inject: NewController,
			Resolve: func(routes web.ServerPool, c *Controller) {
				router, ok := routes.Main()
				if !ok {
					return
				}

				router.Use(web.ThrottlingMiddleware(100))
				router.Get("/users", c.Users)

				api := router.Collection("/api/v1", web.ThrottlingMiddleware(100))
				api.Get("/user/{id}", c.User)
			},
		},
	)
	app.Command("env", func() {
		fmt.Println(os.Environ())
	})
	app.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Users(ctx web.Ctx) {
	metrics.Gauge("users_request").Inc()
	data := Model{
		data: []int64{1, 2, 3, 4},
	}
	ctx.JSON(200, data)
}

func (v *Controller) User(ctx web.Ctx) {
	id, _ := ctx.Param("id").Int() // nolint: errcheck
	ctx.String(200, "user id: %d", id)
	logx.Info("user - %d", id)
}

type Model struct {
	data []int64
}

func (m Model) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.data)
}

```
