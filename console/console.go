/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package console

import (
	"os"
	"reflect"
)

const helpArg = "help"

type Console struct {
	name        string
	description string
	global      []CommandGetter
	root        CommandGetter
}

func New(name, description string) *Console {
	return &Console{
		name:        name,
		description: description,
		global:      make([]CommandGetter, 0, 2),
		root:        NewCommand(func(_ CommandSetter) {}).AsRoot(),
	}
}

func (c *Console) recover() {
	if d := recover(); d != nil {
		Fatalf("%+v", d)
	}
}

func (c *Console) AddGlobal(getter ...CommandGetter) {
	defer c.recover()

	c.global = append(c.global, getter...)
}

func (c *Console) AddCommand(getter ...CommandGetter) {
	defer c.recover()

	c.root.AddCommand(getter...)
}

func (c *Console) RootCommand(getter CommandGetter) {
	defer c.recover()

	next := c.root.List()
	c.root = getter.AsRoot()
	if err := c.root.Validate(); err != nil {
		Fatalf(err.Error())
	}
	c.root.AddCommand(next...)
}

func (c *Console) Exec() {
	defer c.recover()

	args := NewArgs().Parse(os.Args[1:])
	cmd, cur, help := c.build(args)

	if help {
		helpView(os.Args[0], c.description, cmd, c.global, cur)
		return
	}

	for _, gc := range c.global {
		c.run(gc, args.Next()[len(cur):], args)
	}

	c.run(cmd, args.Next()[len(cur):], args)
}

func (c *Console) build(args *Args) (CommandGetter, []string, bool) {
	var (
		i   int
		cmd string

		command CommandGetter
		cur     []string
		help    bool
	)
	for i, cmd = range args.Next() {
		if i == 0 {
			if nc := c.root.Next(cmd); nc != nil {
				command = nc
				continue
			}
			command = c.root
			break
		} else {
			if nc := command.Next(cmd); nc != nil {
				command = nc
				continue
			}
			break
		}
	}

	if len(args.Next()) > 0 {
		cur = args.Next()[:i]
	} else {
		command = c.root
	}

	if args.Has(helpArg) {
		help = true
	}

	return command, cur, help
}

func (c *Console) run(command CommandGetter, a []string, args *Args) {
	rv := make([]reflect.Value, 0)

	if command == nil || command.Call() == nil {
		Fatalf("command not found (use --help for information)")
	}

	callRef := reflect.ValueOf(command.Call())

	if callRef.Type().NumIn() > 0 && callRef.Type().In(0).String() == "[]string" {
		val, err := command.ArgCall(a)
		if err != nil {
			Fatalf("command [%s] validate arguments: %s", command.Name(), err.Error())
		}
		rv = append(rv, reflect.ValueOf(val))
	}

	err := command.Flags().Call(args, func(i interface{}) {
		rv = append(rv, reflect.ValueOf(i))
	})
	if err != nil {
		Fatalf("command [%s] validate flags: %s", command.Name(), err.Error())
	}

	if callRef.Type().NumIn() != len(rv) {
		Fatalf("command [%s] Flags: fewer arguments declared than expected in ExecFunc", command.Name())
	}

	callRef.Call(rv)
}
