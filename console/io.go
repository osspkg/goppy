/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package console

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"go.osspkg.com/errors"
)

const (
	cRESET  = "\u001B[0m"
	cBLACK  = "\u001B[30m"
	cRED    = "\u001B[31m"
	cGREEN  = "\u001B[32m"
	cYELLOW = "\u001B[33m"
	cBLUE   = "\u001B[34m"
	cPURPLE = "\u001B[35m"
	cCYAN   = "\u001B[36m"

	eof = "\n"
)

var (
	scan       *bufio.Scanner
	yesNo             = []string{"y", "n"}
	debugLevel uint32 = 0
)

func init() {
	scan = bufio.NewScanner(os.Stdin)
}

func output(msg string, vars []string, def string) {
	if len(def) > 0 {
		def = fmt.Sprintf(" [%s]", def)
	}
	v := ""
	if len(vars) > 0 {
		v = fmt.Sprintf(" (%s)", strings.Join(vars, "/"))
	}
	Rawf("%s%s%s: ", msg, v, def)
}

// Switch console multi input requests
func Switch(msg string, vars [][]string, exit string, call func(s string)) {
	fmt.Printf("%s\n", msg)

	list := make(map[string]string, len(vars)*4)
	i := 0
	for _, blocks := range vars {
		for _, name := range blocks {
			i++
			fmt.Printf("(%d) %s, ", i, name)
			list[fmt.Sprintf("%d", i)] = name
		}
		fmt.Printf("\n")
	}
	fmt.Printf("and (%s) Done: \n", exit)

	for {
		if scan.Scan() {
			r := scan.Text()
			if r == exit {
				fmt.Printf("\u001B[1A\u001B[K: Done\n\n")
				return
			}
			if name, ok := list[r]; ok {
				call(name)
				fmt.Printf("\033[1A\033[K + %s\n", name)
				continue
			}
			fmt.Printf("\u001B[1A\u001B[KBad answer! Try again: ")
		}
	}
}

// Input console input request
func Input(msg string, vars []string, def string) string {
	output(msg, vars, def)

	for {
		if scan.Scan() {
			r := scan.Text()
			if len(r) == 0 {
				return def
			}
			if len(vars) == 0 {
				return r
			}
			for _, v := range vars {
				if v == r {
					return r
				}
			}
			output("Bad answer! Try again", vars, def)
		}
	}
}

// InputBool console bool input request
func InputBool(msg string, def bool) bool {
	v := "n"
	if def {
		v = "y"
	}
	v = Input(msg, yesNo, v)
	return v == "y"
}

func writeWithColor(c, msg string, args []interface{}) {
	if !strings.HasSuffix(msg, eof) {
		msg += eof
	}
	fmt.Printf(c+msg+cRESET, args...)
}

// Rawf console message writer without level info
func Rawf(msg string, args ...interface{}) {
	writeWithColor(cRESET, msg, args)
}

// Infof console message writer for info level
func Infof(msg string, args ...interface{}) {
	writeWithColor(cRESET, "[INF] "+msg, args)
}

// Warnf console message writer for warning level
func Warnf(msg string, args ...interface{}) {
	writeWithColor(cYELLOW, "[WAR] "+msg, args)
}

// Errorf console message writer for error level
func Errorf(msg string, args ...interface{}) {
	writeWithColor(cRED, "[ERR] "+msg, args)
}

// ShowDebug init show debug
func ShowDebug(ok bool) {
	var v uint32 = 0
	if ok {
		v = 1
	}
	atomic.StoreUint32(&debugLevel, v)
}

// Debugf console message writer for debug level
func Debugf(msg string, args ...interface{}) {
	if atomic.LoadUint32(&debugLevel) > 0 {
		writeWithColor(cBLUE, "[DEB] "+msg, args)
	}
}

// FatalIfErr console message writer if err is not nil
func FatalIfErr(err error, msg string, args ...interface{}) {
	if err != nil {
		Fatalf(errors.Wrapf(err, msg, args...).Error())
	}
}

func WarnIfErr(err error, msg string, args ...interface{}) {
	if err != nil {
		Warnf(errors.Wrapf(err, msg, args...).Error())
	}
}

// Fatalf console message writer with exit code 1
func Fatalf(msg string, args ...interface{}) {
	writeWithColor(cRED, "[ERR] "+msg, args)
	os.Exit(1)
}
