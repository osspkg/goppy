package main

import (
	"fmt"
	"time"

	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/unix"
)

func main() {
	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		unix.WithServer(),
		unix.WithClient(),
	)
	app.Plugins(
		plugins.Plugin{
			Resolve: func(s unix.Server, c unix.Client) error {

				s.Command("demo", func(bytes []byte) ([]byte, error) {
					fmt.Println("<", string(bytes))
					return append(bytes, " world"...), nil
				})

				time.AfterFunc(time.Second*5, func() {
					cc, err := c.Create("/tmp/demo-unix.sock")
					if err != nil {
						panic(err)
					}

					b, err := cc.ExecString("demo", "hello")
					if err != nil {
						panic(err)
					}
					fmt.Println(">", string(b))

					b, err = cc.ExecString("demo", "hello")
					if err != nil {
						panic(err)
					}
					fmt.Println(">", string(b))
				})

				time.AfterFunc(time.Second*15, func() {
					cc, err := c.Create("/tmp/demo-unix.sock")
					if err != nil {
						panic(err)
					}

					b, err := cc.ExecString("demo", "hello")
					if err != nil {
						panic(err)
					}
					fmt.Println(">", string(b))
				})

				return nil
			},
		},
	)
	app.Run()
}
