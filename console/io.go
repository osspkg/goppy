/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package console

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"go.osspkg.com/errors"
)

const (
	colorReset  = "\u001B[0m"
	colorBlack  = "\u001B[30m"
	colorRed    = "\u001B[31m"
	colorGreen  = "\u001B[32m"
	colorYellow = "\u001B[33m"
	colorBlue   = "\u001B[34m"
	colorPurple = "\u001B[35m"
	colorCyan   = "\u001B[36m"

	newLine   = "\n"
	clearLine = "\033[2K"

	cursorUp   = "\033[A"
	cursorDown = "\033[B"
	cursorHide = "\033[?25l"
	cursorShow = "\033[?25h"
)

var (
	yesNo             = []string{"y", "n"}
	debugLevel uint32 = 0
)

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

func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	FatalIfErr(cmd.Run(), "failed to clear screen")
}

func IsInteractiveTerminal() bool {
	fileInfo, _ := os.Stdin.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) == 0
}

func disableInputBuffering() {
	fmt.Print(cursorHide)
	FatalIfErr(errors.Wrap(
		exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run(),
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run(),
	), "failed to disable input buffering")
}

func enableInputBuffering() {
	FatalIfErr(exec.Command("stty", "-F", "/dev/tty", "echo").Run(),
		"failed to enable input buffering")
	fmt.Print(cursorShow)
}

func getTerminalWidth() (int, int) {
	cmd := exec.Command("stty", "-F", "/dev/tty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80, 24 // fallback
	}
	parts := strings.Fields(string(out))
	if len(parts) < 2 {
		return 80, 24
	}
	h, _ := strconv.Atoi(parts[0])
	w, _ := strconv.Atoi(parts[1])
	return w, h
}

type InteractiveMenu struct {
	Title       string
	Items       []string
	CallBack    func(...string)
	MultiChoice bool
	MaxCols     int
}

//nolint:gocyclo
func (m InteractiveMenu) Run() {
	if len(m.Items) == 0 || m.CallBack == nil {
		return
	}

	if !IsInteractiveTerminal() {
		Fatalf("interactive terminal disabled")
	}

	ClearScreen()
	disableInputBuffering()
	defer enableInputBuffering()

	selected := make(map[int]bool, len(m.Items))
	current := 0
	offsetRow := 0

	maxRunes := 0
	for _, item := range m.Items {
		rCount := utf8.RuneCountInString(item)
		if rCount > maxRunes {
			maxRunes = rCount
		}
	}
	cellWidth := maxRunes + 7
	m.MaxCols = max(m.MaxCols, 1)
	m.Title = strings.Trim(m.Title, ":\n\r")

	for {
		if m.MultiChoice {
			fmt.Printf("\r%s%s (arrows to navigate, space to select, 'q' to exit):%s\n", colorGreen, m.Title, colorReset)
		} else {
			fmt.Printf("\r%s%s (arrows to navigate, 'q' to exit):%s\n", colorGreen, m.Title, colorReset)
		}

		termWidth, termHeight := getTerminalWidth()
		reservedRows := 4
		visibleRows := termHeight - reservedRows
		if visibleRows <= 0 {
			visibleRows = 1
		}

		cols := min(max(termWidth/cellWidth, 1), m.MaxCols)
		totalRows := (len(m.Items) + cols - 1) / cols
		currentRowInGrid := current % totalRows
		if currentRowInGrid < offsetRow {
			offsetRow = currentRowInGrid
		} else if currentRowInGrid >= offsetRow+visibleRows {
			offsetRow = currentRowInGrid - visibleRows + 1
		}
		rowsToRender := visibleRows
		if totalRows < visibleRows {
			rowsToRender = totalRows
		}

		for r := 0; r < rowsToRender; r++ {
			actualRow := r + offsetRow
			line := "\r"
			for c := 0; c < cols; c++ {
				i := actualRow + c*totalRows
				if i >= len(m.Items) {
					continue
				}

				color := colorReset
				marker := "[ ]"
				if selected[i] {
					color = colorCyan
					marker = "[x]"
				}
				prefix := "  "
				if i == current {
					color = colorRed
					prefix = "→ "
				}
				if !m.MultiChoice {
					marker = ""
				}

				padding := cellWidth - (utf8.RuneCountInString(prefix) +
					utf8.RuneCountInString(marker) + utf8.RuneCountInString(m.Items[i]))
				line += color + prefix + marker + m.Items[i] + colorReset + strings.Repeat(" ", max(padding, 0))
			}
			fmt.Println(line)
		}
		if m.MultiChoice {
			fmt.Printf("\r%sSelected: %d/%d %s\n", colorGreen, len(selected), len(m.Items), colorReset)
		} else {
			fmt.Printf("\r%sTotal %d %s\n", colorGreen, len(m.Items), colorReset)
		}

		var buf [3]byte
		_, err := os.Stdin.Read(buf[:])
		FatalIfErr(err, "failed to read from stdin")

		switch {
		case buf[0] == 3 || buf[0] == 'q': // Ctrl+C
			return

		case buf[0] == 13 || buf[0] == 10: // Enter
			if !m.MultiChoice {
				m.CallBack(m.Items[current])
				return
			}

			result := make([]string, 0, len(m.Items))
			for i, s := range m.Items {
				if selected[i] {
					result = append(result, s)
				}
			}
			m.CallBack(result...)
			return

		case buf[0] == 27 && buf[1] == 91 && buf[2] == 65: // Up
			if current%totalRows > 0 {
				current--
			}

		case buf[0] == 27 && buf[1] == 91 && buf[2] == 66: // Down
			if current%totalRows < totalRows-1 && current < len(m.Items)-1 {
				current++
			}

		case buf[0] == 27 && buf[1] == 91 && buf[2] == 68: // Left
			if current >= totalRows {
				current -= totalRows
			}

		case buf[0] == 27 && buf[1] == 91 && buf[2] == 67: // Right
			if current+totalRows < len(m.Items) {
				current += totalRows
			}

		case m.MultiChoice && buf[0] == 32: // Space
			selected[current] = !selected[current]

		default:
		}

		for i := 0; i <= rowsToRender+1; i++ {
			fmt.Print(cursorUp + clearLine)
		}
	}
}

func Select(msg string, vars []string, def string) string {
	scan := bufio.NewScanner(os.Stdin)

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
				if strings.EqualFold(v, r) {
					return v
				}
			}
			output("Bad answer! Try again", vars, def)
		}
	}
}

func SelectBool(msg string, def bool) bool {
	v := "n"
	if def {
		v = "y"
	}
	v = Select(msg, yesNo, v)
	return v == "y"
}

func writeWithColor(c, msg string, args []interface{}) {
	if !strings.HasSuffix(msg, newLine) {
		msg += newLine
	}
	fmt.Printf(c+msg+colorReset, args...)
}

func Rawf(msg string, args ...interface{}) {
	writeWithColor(colorReset, msg, args)
}

func Infof(msg string, args ...interface{}) {
	writeWithColor(colorReset, "[INF] "+msg, args)
}

func Warnf(msg string, args ...interface{}) {
	writeWithColor(colorYellow, "[WAR] "+msg, args)
}

func Errorf(msg string, args ...interface{}) {
	writeWithColor(colorRed, "[ERR] "+msg, args)
}

func ShowDebug(ok bool) {
	var v uint32 = 0
	if ok {
		v = 1
	}
	atomic.StoreUint32(&debugLevel, v)
}

func Debugf(msg string, args ...interface{}) {
	if atomic.LoadUint32(&debugLevel) > 0 {
		writeWithColor(colorBlue, "[DEB] "+msg, args)
	}
}

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

func RawIfErr(err error, msg string, args ...interface{}) {
	if err != nil {
		Rawf(errors.Wrapf(err, msg, args...).Error())
	}
}

func Fatalf(msg string, args ...interface{}) {
	writeWithColor(colorRed, "[ERR] "+msg, args)
	os.Exit(1)
}
