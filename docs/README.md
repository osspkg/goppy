# Goppy Microservice Toolkit

[![Release](https://img.shields.io/github/release/osspkg/goppy.svg?style=flat-square)](https://github.com/osspkg/goppy/releases/latest)
![GitHub](https://img.shields.io/github/license/osspkg/goppy)
[![Forum](https://img.shields.io/badge/community-forum-red)](https://github.com/osspkg/goppy/discussions)

## Installation

```bash
go get -u go.osspkg.com/goppy/v2
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

## Plugins

| Plugin        | Comment                                                                                                                                                             | Import                                                                                                                           |
|---------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------|
| **metrics**   | Profiling application (pprof) and metrics collection (prometheus) with access via HTTP.                                                                             | `go.osspkg.com/goppy/v2/metrics`<br/> `metrics.WithServer()`                                                                     |
| **http**      | Out of the box multi-server launch of web servers with separate routing. Grouping of routers with connection to a group of dedicated middleware. HTTP clients pool. | `go.osspkg.com/goppy/v2/web`<br/> `web.WithServer()`<br/> `web.WithClient()`                                                     |
| **websocket** | Ready-made websocket handler for server and client. Websocket server pool.                                                                                          | `go.osspkg.com/goppy/v2/ws`<br/> `ws.WithServer()`<br/> `ws.WithClient()`<br/> `ws.WithServerPool()`                             |
| **database**  | Multiple connection pools with MySQL, SQLite, Postgre databases (with automatic migration setup).                                                                   | `go.osspkg.com/goppy/v2/orm`<br/> `orm.WithORM()` <br/><br/> `orm.WithMysql()`  <br/> `orm.WithSqlite()`<br/> `orm.WithPGSql()`  |
| **geoip**     | Definition of geo-IP information.                                                                                                                                   | `go.osspkg.com/goppy/v2/geoip`<br/> `geoip.WithMaxMindGeoIP()` + `geoip.CloudflareMiddleware()`<br/> `geoip.MaxMindMiddleware()` |
| **oauth**     | Authorization via OAuth provider (Yandex, Google). JWT Cookie.                                                                                                      | `go.osspkg.com/goppy/v2/auth`<br/> `auth.WithOAuth()`<br/> `auth.WithJWT()` + `auth.JWTGuardMiddleware()`                        |

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
	"fmt"
	"os"

	"go.osspkg.com/console"
	"go.osspkg.com/goppy/v2"
	"go.osspkg.com/goppy/v2/metrics"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
	"go.osspkg.com/logx"
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
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes web.RouterPool, c *Controller) {
				router := routes.Main()
				router.Use(web.ThrottlingMiddleware(100))
				router.Get("/users", c.Users)

				api := router.Collection("/api/v1", web.ThrottlingMiddleware(100))
				api.Get("/user/{id}", c.User)
			},
		},
	)
	app.Command("env", func(s console.CommandSetter) {
		fmt.Println(os.Environ())
	})
	app.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Users(ctx web.Context) {
	metrics.Gauge("users_request").Inc()
	data := []int64{1, 2, 3, 4}
	ctx.JSON(200, data)
}

func (v *Controller) User(ctx web.Context) {
	id, _ := ctx.Param("id").Int() // nolint: errcheck
	ctx.String(200, "user id: %d", id)
	logx.Info("user - %d", id)
}

```
