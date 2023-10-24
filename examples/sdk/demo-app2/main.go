/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"

	"go.osspkg.com/goppy/sdk/app"
	"go.osspkg.com/goppy/sdk/log"
)

type (
	Test0 struct{}
	Test1 struct{}
	Test2 struct{}

	Config struct {
		Env   string `yaml:"env"`
		Level string `yaml:"level"`
	}

	Params struct {
		Test1  *Test1
		Config Config
	}
)

func (s *Test2) Up() error {
	fmt.Println("--> call *Test2.Up")
	return nil
}

func (s *Test2) Down() error {
	fmt.Println("--> call *Test2.Down")
	return nil
}

func NewTest0(p Params) *Test0 {
	fmt.Println("--> call NewTest0")
	fmt.Println("--> Params.Config.Env=" + p.Config.Env)
	return &Test0{}
}

func NewTest2(_ *Test0) *Test2 {
	fmt.Println("--> call NewTest2")
	return &Test2{}
}

func main() {
	app.New().
		Logger(log.Default()).
		ConfigFile(
			"./config.yaml",
			Config{},
		).
		Modules(
			&Test1{},
			NewTest0,
			NewTest2,
		).
		Run()
}
