/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package sqlite

import (
	"fmt"
	"net/url"
	"strings"

	"go.osspkg.com/goppy/v2/orm/dialect"
)

type ConfigGroup struct {
	Pool []Config `yaml:"sqlite"`
}

func (c *ConfigGroup) List() []dialect.ItemInterface {
	list := make([]dialect.ItemInterface, 0, len(c.Pool))
	for _, item := range c.Pool {
		list = append(list, item)
	}
	return list
}

func (c *ConfigGroup) Default() {
	if len(c.Pool) == 0 {
		c.Pool = []Config{
			{
				Tags:        "master",
				File:        "./sqlite.db",
				Cache:       "private",
				Mode:        "rwc",
				Journal:     "WAL",
				LockingMode: "EXCLUSIVE",
				OtherParams: "auto_vacuum=incremental",
			},
		}
	}
}

type Config struct {
	Tags        string `yaml:"tags"`
	File        string `yaml:"file"`
	Cache       string `yaml:"cache"`
	Mode        string `yaml:"mode"`
	Journal     string `yaml:"journal"`
	LockingMode string `yaml:"locking_mode"`
	OtherParams string `yaml:"other_params"`
}

func (i Config) GetTags() []string {
	return strings.Split(i.Tags, ",")
}

func (i Config) Setup(_ dialect.SetupInterface) {}

// GetDSN connection params
func (i Config) GetDSN() string {
	params, err := url.ParseQuery(i.OtherParams)
	if err != nil {
		params = url.Values{}
	}
	// ---
	if len(i.Cache) == 0 {
		i.Cache = "private"
	}
	params.Add("cache", i.Cache)
	// ---
	if len(i.Mode) == 0 {
		i.Mode = "rwc"
	}
	params.Add("mode", i.Mode)
	// ---
	if len(i.Journal) == 0 {
		i.Journal = "TRUNCATE"
	}
	params.Add("_journal", i.Journal)
	// ---
	if len(i.LockingMode) == 0 {
		i.LockingMode = "EXCLUSIVE"
	}
	params.Add("_locking_mode", i.LockingMode)
	// --
	return fmt.Sprintf("file:%s?%s", i.File, params.Encode())
}
