/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

const (
	indexFilename = "/store/root.bolt"

	customAnalyzerName = "goppy_analyzer"
	customFilterName   = "goppy_filter"

	structScoreField = "Score"

	FieldText = "text"
	FieldDate = "date"
)

type (
	Config struct {
		Search ConfigItem `yaml:"search"`
	}
	ConfigItem struct {
		Folder  string        `yaml:"folder"`
		Indexes []ConfigIndex `yaml:"indexes"`
	}
	ConfigIndex struct {
		Name   string             `yaml:"name"`
		Fields []ConfigIndexField `yaml:"fields"`
	}
	ConfigIndexField struct {
		Name string `yaml:"name"`
		Type string `yaml:"type"`
	}
)
