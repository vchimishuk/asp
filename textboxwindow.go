package main

// TODO: If enterred text is wider that window we should not breaks,
//       but let the user scroll and navigate over the text with
//       LEFT/RIGHT/^F/^B keys.
// TODO: Unicode input support.

import (
	"strings"
	"unicode"

	ncurses "github.com/gbin/goncurses"
)

const (
	// TODO: CamelCase constants.
	KEY_ACK  ncurses.Key = 0x06 // ^F
	KEY_BELL ncurses.Key = 0x07 // ^G
	KEY_ENQ  ncurses.Key = 0x05 // ^E
	KEY_EOT  ncurses.Key = 0x04 // ^D
	KEY_ESC  ncurses.Key = 0x1B
	KEY_ETB  ncurses.Key = 0x17 // ^W
	KEY_SOH  ncurses.Key = 0x01 // ^A
	KEY_STX  ncurses.Key = 0x02 // ^B
	KEY_VT   ncurses.Key = 0x0B // ^K
)

type TextboxWindow struct {
	window *ncurses.Window
}

func NewTextboxWindow(w, y, x int) (*TextboxWindow, error) {
	window, err := ncurses.NewWindow(1, w, y, x)
	if err != nil {
		return nil, err
	}
	// window.SetBackground(ncurses.ColorPair(1))
	window.Keypad(true)

	return &TextboxWindow{window: window}, nil
}

// TODO: This method has a lot of bugs and pretty hard to understand.
//       We need to re-implement it in a clear straightforward way.
func (w *TextboxWindow) Input(prompt string) string {
	prompt += " "

	ncurses.Cursor(1)
	x := len(prompt) // Cursor position in window.
	pos := 0         // Cursor position in string.
	buf := ""

	for {
		maxX := w.maxX()
		s := max(0, pos-x-len(prompt))
		e := min(len(buf), pos+len(prompt)+maxX-x)

		w.window.MovePrint(0, 0, prompt)
		if x == maxX && len(buf) > maxX-1-len(prompt) {
			w.window.MovePrint(0, len(prompt), buf[s+1:e])
			w.window.Print(" ")
		} else {
			visible := buf[s:e]
			w.window.MovePrint(0, len(prompt), visible)
			if len(prompt)+len(visible) < maxX {
				w.window.Print(strings.Repeat(" ", maxX-len(prompt)-len(visible)))
			}
			w.window.Move(0, x)
		}
		w.window.Refresh()

		ch := w.window.GetChar()

		if ch == 0 {
			continue
		} else if ch == ncurses.KEY_RETURN {
			break
		} else if ch == KEY_SOH { // ^A
			pos = 0
			x = len(prompt)
		} else if ch == KEY_STX || ch == ncurses.KEY_LEFT { // ^B
			pos = max(0, pos-1)
			x = max(len(prompt), x-1)
		} else if ch == KEY_EOT { // ^D
			// TODO:
		} else if ch == KEY_ENQ { // ^E
			pos = len(buf)
			x = min(len(buf)+len(prompt), w.maxX())
		} else if ch == KEY_ACK || ch == ncurses.KEY_RIGHT { // ^F
			pos = min(len(buf), pos+1)
			x = min(len(buf)+len(prompt), min(w.maxX(), x+1))
		} else if ch == KEY_BELL || ch == KEY_ESC { // ^G, ESC
			buf = ""
			break
		} else if ch == ncurses.KEY_BACKSPACE {
			if pos > 0 {
				buf = buf[:pos-1] + buf[pos:]
				pos--
				x--
			}
		} else if ch == KEY_ETB { // ^W
			// TODO:
		} else if ch == KEY_VT { // ^K
			// TODO:
			buf = buf[:pos]
		} else if unicode.IsPrint(rune(ch)) {
			buf = buf[:pos] + string(ch) + buf[pos:]
			pos++
			x = min(w.maxX(), x+1)
		}
	}

	ncurses.Cursor(0)
	w.erase()
	w.window.Refresh()

	return buf
}

// // draw draws text inputted by user in such way that cursor stands at
// // pos position which is positioned at x window coordinate.
// func (w *TextboxWindow) draw(buf string, pos int, x int) {
// 	maxX := w.maxX()
// 	s := max(0, pos-x)
// 	e := min(len(buf), pos+maxX-x)

// 	if x == maxX && len(buf) > maxX-1 {
// 		w.window.MovePrint(0, 0, buf[s+1:e])
// 		w.window.Print(" ")
// 	} else {
// 		w.window.MovePrint(0, 0, buf[s:e])
// 		w.window.Move(0, x)
// 	}
// }

func (w *TextboxWindow) erase() {
	w.window.MovePrint(0, 0, strings.Repeat(" ", w.maxX()))

}

func (w *TextboxWindow) maxX() int {
	_, x := w.window.MaxYX()

	return x
}
