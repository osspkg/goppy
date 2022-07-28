# Goppy Microservice Toolkit 


[![Release](https://img.shields.io/github/release/dewep-online/goppy.svg?style=flat-square)](https://github.com/dewep-online/goppy/releases/latest)
![GitHub](https://img.shields.io/github/license/dewep-online/goppy)
[![Forum](https://img.shields.io/badge/community-forum-red)](https://github.com/dewep-online/goppy/discussions)

## Installation

```bash
go get -u github.com/dewep-online/goppy
```

## Features

- Config auto generation
- Custom pool of HTTP servers with configuration via config
- Group APIs with middleware hanging on each group
- Extensible middleware framework
- Application customization via plugins
- Built-in dependency container
- Data binding for JSON

## Plugins

| Plugin       | Comment                                                                                                                                          | Import                                               |
|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------|
| **debug**    | Profiling application (pprof) with HTTP access.                                                                                                  | `http.WithHTTPDebug()`                               |
| **http**     | Out of the box multi-server launch of web servers with separate routing. Grouping of routers with connection to a group of dedicated middleware. | `http.WithHTTP()`  `http.WithWebsocket()`                                  |
| **database** | Multi connection pools with MySQL and SQLite databases (with initialization migration setup).                                                    | `database.WithMySQL()` `database.WithSQLite()`       |
| **geoip**    | Definition of geo-IP information.                                                                                                                | `geoip.WithMaxMindGeoIP()` + `middlewares.CloudflareMiddleware()` `middlewares.MaxMindMiddleware()` |


## Quick Start

Config:

```yaml
env: dev
level: 4
log: /dev/stdout

http:
    main:
        addr: 127.0.0.1:8088
```

Code:

```go
package main

import (
	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/middlewares"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/http"
)

func main() {
	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		http.WithHTTP(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes http.RouterPool, c *Controller) {
				router := routes.Main()
				router.Use(middlewares.ThrottlingMiddleware(100))
				router.Get("/users", c.Users)

				api := router.Collection("/api/v1", middlewares.ThrottlingMiddleware(100))
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
	ctx.SetBody(200).JSON(data)
}

func (v *Controller) User(ctx http.Ctx) {
	id, _ := ctx.Param("id").Int()
	ctx.SetBody(200).String("user id: %d", id)
	ctx.Log().Infof("user - %d", id)
}
```
