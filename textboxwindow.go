package main

// TODO: Implement undo action (^/).
// TODO: Unicode input support.
// TODO: Add ^U key.

import (
	"strings"
	"unicode"

	ncurses "github.com/gbin/goncurses"
)

const (
	KEY_ACK  ncurses.Key = 0x06 // ^F
	KEY_BELL             = 0x07 // ^G
	KEY_BS               = 0x08 // ^H
	KEY_ENQ              = 0x05 // ^E
	KEY_EOT              = 0x04 // ^D
	KEY_ESC              = 0x1B
	KEY_ETB              = 0x17 // ^W
	KEY_SOH              = 0x01 // ^A
	KEY_STX              = 0x02 // ^B
	KEY_VT               = 0x0B // ^K
)

type TextboxWindow struct {
	window *ncurses.Window
}

func NewTextboxWindow(w, y, x int) (*TextboxWindow, error) {
	window, err := ncurses.NewWindow(1, w, y, x)
	if err != nil {
		return nil, err
	}
	window.Keypad(true)

	return &TextboxWindow{window: window}, nil
}

func (w *TextboxWindow) Input(prompt string) string {
	prompt += " "

	ncurses.Cursor(1)
	// Width of edit area.
	width := w.maxX() - len(prompt)
	// Cursor position in window.
	x := 0
	// First index of uffer visible on screen.
	o := 0
	buf := ""

loop:
	for {
		s := buf[o:min(o+width, len(buf))]
		// Buffer offset, cursor position in buffer.
		bo := o + x

		w.window.MovePrint(0, 0, prompt)
		w.window.Print(s)
		if len(s) < width {
			w.window.Print(strings.Repeat(" ", width-len(s)))
		}
		w.window.Move(0, len(prompt)+x)
		w.window.Refresh()

		ch := w.window.GetChar()

		if ch == 0 {
			continue
		} else if ch == ncurses.KEY_RETURN {
			break
		} else if ch == KEY_SOH || ch == ncurses.KEY_HOME {
			o = 0
			x = 0
		} else if ch == KEY_STX || ch == ncurses.KEY_LEFT {
			x--
		} else if ch == KEY_EOT || ch == ncurses.KEY_DC {
			if bo < len(buf) {
				buf = buf[:bo] + buf[bo+1:]
			}
		} else if ch == KEY_ENQ || ch == ncurses.KEY_END {
			o = max(0, len(buf)-width+1)
			x = len(buf) - o
		} else if ch == KEY_ACK || ch == ncurses.KEY_RIGHT {
			if x < len(buf) {
				x++
			}
		} else if ch == KEY_BELL {
			buf = ""
			break
		} else if ch == KEY_ESC {
			switch w.window.GetChar() {
			case 'b':
				i := wordBegin(buf, bo)
				d := bo - i
				if d > x {
					o = max(0, i-1)
					x = min(1, o)
				} else {
					x -= d
				}
			case 'd':
				i := wordEnd(buf, bo)
				buf = buf[:bo] + buf[i:]
			case 'f':
				i := wordEnd(buf, bo)
				d := i - bo
				x += d
				if x >= width {
					x = width - 1
					o = i - x
				}
			default:
				buf = ""
				break loop
			}
		} else if ch == ncurses.KEY_BACKSPACE || ch == KEY_BS {
			if bo > 0 {
				buf = buf[:bo-1] + buf[bo:]
				x--
			}
		} else if ch == KEY_ETB {
			i := wordBegin(buf, bo)
			buf = buf[:i] + buf[bo:]
			d := bo - i
			if d > x {
				o = max(0, i-1)
				x = min(1, o)
			} else {
				x -= d
			}
		} else if ch == KEY_VT {
			buf = buf[:bo]
		} else if unicode.IsPrint(rune(ch)) {
			buf = buf[:bo] + string(ch) + buf[bo:]
			x++
		}

		if x == 0 && o != 0 {
			x++
			o--
		}
		if x < 0 {
			x = 0
		}
		if x == width || (x == width-1 && len(buf)-o > width) {
			x--
			o = min(o+1, len(buf)-width+1)
		}
	}

	ncurses.Cursor(0)
	w.erase()
	w.window.Refresh()

	return buf
}

func (w *TextboxWindow) erase() {
	w.window.MovePrint(0, 0, strings.Repeat(" ", w.maxX()))

}

func (w *TextboxWindow) maxX() int {
	_, x := w.window.MaxYX()

	return x
}

func wordBegin(s string, pos int) int {
	trim := true
	done := false
	for pos > 0 && !done {
		c := rune(s[pos-1])
		if !unicode.IsSpace(c) {
			if !isWord(c) {
				if trim {
					done = true
				} else {
					break
				}
			}
			trim = false
		} else if !trim {
			break
		}
		pos--
	}

	return pos
}

func wordEnd(s string, pos int) int {
	trim := false
	first := true

	for pos < len(s) {
		c := rune(s[pos])
		if unicode.IsSpace(c) {
			trim = true
		} else if isWord(c) {
			if trim {
				break
			}
		} else {
			if first {
				trim = true
			} else {
				break
			}
		}
		first = false
		pos++
	}

	return pos
}

func isWord(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_'
}
