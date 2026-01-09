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
- Executing console commands
- Automatic dependency resolution at startup
- Database support and automatic migration

## Quick Start

### Config:

Write log to file:
```yaml
log:
  file_path: /dev/stdout
  format: string # json, string
  level: 4 # 0-Fatal, 1-Error, 2-Warning, 3-Info, 4-Debug
```

Write log to syslog:
```yaml
log:
  file_path: syslog
  format: string # json, string
  level: 4 # 0-Fatal, 1-Error, 2-Warning, 3-Info, 4-Debug
```

Write log to remote syslog:
```yaml
log:
  file_path: syslog=udp://syslog-server.example.com:514
  format: string # json, string
  level: 4 # 0-Fatal, 1-Error, 2-Warning, 3-Info, 4-Debug
```

## Example

Config
```yaml
log:
  file_path: /dev/stdout
  format: string 
  level: 4 

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

	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/console"
	"go.osspkg.com/goppy/v3/dic/broker"
	"go.osspkg.com/goppy/v3/metrics"
	"go.osspkg.com/goppy/v3/web"
	"go.osspkg.com/logx"
	"go.osspkg.com/xc"
)

type IStatus interface {
	GetStatus() int
}

func main() {
	// Specify the path to the config via the argument: `--config`.
	// Specify the path to the pidfile via the argument: `--pid`.
	app := goppy.New("app_name", "v1.0.0", "app description")
	app.Plugins(
		metrics.WithServer(),
		web.WithServer(),
	)
	app.Plugins(
		NewController,
		func(routes web.ServerPool, c *Controller) {
			router, ok := routes.Main()
			if !ok {
				return
			}

			router.Use(web.ThrottlingMiddleware(100))
			router.Get("/users", c.Users)

			api := router.Collection("/api/v1", web.ThrottlingMiddleware(100))
			api.Get("/user/{id}", c.User)
		},
		broker.WithUniversalBroker[IStatus](
			func(_ xc.Context, status IStatus) error {
				fmt.Println(">> UniversalBroker got status", status.GetStatus())
				return nil
			},
			func(status IStatus) error {
				return nil
			},
		),
	)
	app.Command(func(setter console.CommandSetter) {
		setter.Setup("env", "show all envs")
		setter.ExecFunc(func() {
			fmt.Println(os.Environ())
		})
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

func (v *Controller) GetStatus() int {
	return 200
}

type Model struct {
	data []int64
}

func (m Model) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.data)
}


```
