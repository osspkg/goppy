# Goppy Microservice Toolkit

[![Release](https://img.shields.io/github/release/osspkg/goppy.svg?style=flat-square)](https://github.com/osspkg/goppy/releases/latest)
![GitHub](https://img.shields.io/github/license/osspkg/goppy)
[![Forum](https://img.shields.io/badge/community-forum-red)](https://github.com/osspkg/goppy/discussions)

## Installation

```bash
go get -u github.com/osspkg/goppy
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

| Plugin         | Comment                                                                                                                                                             | Import                                                                                          |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| **debug**      | Profiling application (pprof) with HTTP access.                                                                                                                     | `web.WithHTTPDebug()`                                                                           |
| **http**       | Out of the box multi-server launch of web servers with separate routing. Grouping of routers with connection to a group of dedicated middleware. HTTP clients pool. | `web.WithHTTP()` `web.WithWebsocketServer()` `web.WithWebsocketClient()` `web.WithHTTPClient()` |
| **unixsocket** | Requests via unix socket.                                                                                                                                           | `unix.WithServer()` `unix.WithClient()`                                                         |
| **database**   | Multiple connection pools with MySQL, SQLite, Postgre databases (with automatic migration setup).                                                                   | `database.WithMySQL()` `database.WithSQLite()` `database.WithPostgreSQL()`                      |
| **geoip**      | Definition of geo-IP information.                                                                                                                                   | `geoip.WithMaxMindGeoIP()` + `geoip.CloudflareMiddleware()` `geoip.MaxMindMiddleware()`         |
| **oauth**      | Authorization via OAuth provider (Yandex, Google). JWT Cookie.                                                                                                      | `auth.WithOAuth()` `auth.WithJWT()` `auth.JWTGuardMiddleware()`                                 |

## Quick Start

Config:

```yaml
env: dev
level: 4 # 0-Fatal, 1-Error, 2-Warning, 3-Info, 4-Debug
log: /dev/stdout

http:
  main:
    addr: 127.0.0.1:8080
```

Code:

```go
package main

import (
	"fmt"
	"os"

	"github.com/osspkg/goppy"
	"github.com/osspkg/goppy/plugins"
	"github.com/osspkg/goppy/plugins/web"
)

func main() {
	// Specify the path to the config via the argument: `--config`.
	// Specify the path to the pidfile via the argument: `--pid`.
	app := goppy.New()
	app.Plugins(
		web.WithHTTP(),
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
	app.Command("env", func() {
		fmt.Println(os.Environ())
	})
	app.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Users(ctx web.Context) {
	data := []int64{1, 2, 3, 4}
	ctx.JSON(200, data)
}

func (v *Controller) User(ctx web.Context) {
	id, _ := ctx.Param("id").Int()
	ctx.String(200, "user id: %d", id)
	ctx.Log().Infof("user - %d", id)
}

```
