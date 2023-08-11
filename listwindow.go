package main

import (
	"fmt"
	"strings"

	ncurses "github.com/gbin/goncurses"
	"github.com/vchimishuk/asp/config"
)

type ListItem interface {
	Format(width int) string
	IsSelected(val string) bool
}

type ListWindow struct {
	window   *ncurses.Window
	items    []ListItem
	input    func(string) string
	selected string
	// First visible item index.
	offset int
	// Cursor index. Cursor is an selected row which is
	// manipulated by user with keyboard.
	cursor    int
	searchStr string
}

func NewListWindow(parent *ncurses.Window, input func(string) string) (*ListWindow, error) {
	h, w := parent.MaxYX()
	window := parent.Sub(h-1, w, 1, 0)

	return &ListWindow{
		window:    window,
		input:     input,
		items:     nil,
		offset:    -1,
		selected:  "",
		cursor:    -1,
		searchStr: "",
	}, nil
}

func (w *ListWindow) Command(cmd config.Cmd) {
	switch cmd {
	case config.CmdEnd:
		w.End()
	case config.CmdHome:
		w.Home()
	case config.CmdNext:
		w.Next()
	case config.CmdPageDown:
		w.PageDown()
	case config.CmdPageUp:
		w.PageUp()
	case config.CmdPrev:
		w.Prev()
	case config.CmdSearch:
		w.search()
	case config.CmdSearchNext:
		w.searchNext()
	case config.CmdSearchPrev:
		w.searchPrev()
	}
}

func (w *ListWindow) Clear() {
	w.items = nil
	w.offset = -1
	w.cursor = -1

	w.refresh()
}

func (w *ListWindow) Add(items ...ListItem) {
	w.items = append(w.items, items...)

	if w.cursor == -1 && len(w.items) > 0 {
		w.offset = 0
		w.cursor = 0
	}

	w.refresh()
}

func (w *ListWindow) Selected() string {
	return w.selected
}

func (w *ListWindow) SetSelected(s string) {
	old := w.selected
	w.selected = s

	if old != s {
		w.refresh()
	}
}

// func (w *ListWindow) Selected() ListItem {
// 	if w.selected != -1 {
// 		return w.items[w.selected]
// 	} else {
// 		return nil
// 	}
// }

// func (w *ListWindow) SetSelected(i int) error {
// 	if i < 0 || i >= len(w.items) {
// 		return fmt.Errorf("Can't select %d item of %d", i, len(w.items))
// 	}
// 	w.selected = i
// 	w.redraw()

// 	return nil
// }

func (w *ListWindow) Cursor() ListItem {
	if w.cursor != -1 {
		return w.items[w.cursor]
	} else {
		return nil
	}
}

func (w *ListWindow) SetCursor(i int) error {
	if i < 0 || i >= len(w.items) {
		return fmt.Errorf("%d out of range 0..%d", i, len(w.items))
	}

	w.cursor = i
	h := w.height()
	w.offset = min(max(0, w.cursor-h/2), max(0, len(w.items)-h))
	w.refresh()

	return nil
}

func (w *ListWindow) Next() {
	if len(w.items) > 0 && w.cursor < len(w.items)-1 {
		w.cursor++
		h := w.height()
		if w.cursor > w.offset+h-1 {
			w.offset = min(w.cursor-h/2, len(w.items)-h)
		}
		w.refresh()
	}
}

func (w *ListWindow) Prev() {
	if len(w.items) > 0 && w.cursor > 0 {
		w.cursor--
		if w.cursor < w.offset {
			w.offset = max(w.cursor-w.height()/2, 0)
		}
		w.refresh()
	}
}

// TODO: mc-style page up/down listing.
func (w *ListWindow) PageDown() {
	h := w.height()
	last := len(w.items) - 1
	lastVis := min(last, w.offset+h-1)

	if w.cursor == lastVis && last > lastVis {
		w.offset = min(last-h+1, w.offset+h)
		w.cursor = min(last, w.cursor+h)
	} else {
		w.cursor = lastVis
	}

	w.refresh()
}

func (w *ListWindow) PageUp() {
	h := w.height()

	if w.cursor == w.offset {
		w.offset = max(0, w.offset-h)
		w.cursor = w.offset
	} else {
		w.cursor = w.offset
	}

	w.refresh()
}

func (w *ListWindow) Home() {
	if len(w.items) > 0 {
		w.offset = 0
		w.cursor = 0
		w.refresh()
	}
}

func (w *ListWindow) End() {
	if len(w.items) > 0 {
		w.cursor = len(w.items) - 1
		w.offset = max(0, w.cursor-w.height()+1)
		w.refresh()
	}
}

// // Position returns vertical scrollbar position.
// func (w *ListWindow) Position() int {
// 	h := w.height()
// 	last := len(w.items) - 1
// 	lastVis := min(last, w.offset+h-1)

// 	if lastVis < last {
// 		return lastVis * 100 / last
// 	} else {
// 		return 100
// 	}
// }

func (w *ListWindow) search() {
	s := strings.ToLower(strings.TrimSpace(w.input("Search:")))

	if s != "" {
		w.searchStr = s
		w.searchNext()
	}
}

func (w *ListWindow) searchNext() {
	if w.searchStr == "" {
		return
	}

	_, width := w.window.MaxYX()
	for i := w.cursor + 1; i < len(w.items); i++ {
		s := strings.ToLower(w.items[i].Format(width))
		if strings.Contains(s, w.searchStr) {
			w.SetCursor(i)
			return
		}
	}
	for i := 0; i < w.cursor; i++ {
		s := strings.ToLower(w.items[i].Format(width))
		if strings.Contains(s, w.searchStr) {
			w.SetCursor(i)
			return
		}
	}
}

func (w *ListWindow) searchPrev() {
	if w.searchStr == "" {
		return
	}

	_, width := w.window.MaxYX()
	for i := w.cursor - 1; i >= 0; i-- {
		s := strings.ToLower(w.items[i].Format(width))
		if strings.Contains(s, w.searchStr) {
			w.SetCursor(i)
			return
		}
	}
	for i := len(w.items) - 1; i > w.cursor; i-- {
		s := strings.ToLower(w.items[i].Format(width))
		if strings.Contains(s, w.searchStr) {
			w.SetCursor(i)
			return
		}
	}
}

func (w *ListWindow) height() int {
	y, _ := w.window.MaxYX()

	return y
}

func (w *ListWindow) refresh() {
	height, width := w.window.MaxYX()
	l := len(w.items)

	for i := 0; i < height; i++ {
		var attr ncurses.Char
		var s string
		ii := w.offset + i

		if w.offset == -1 || ii > l-1 {
			attr = config.ColorNormal
			s = strings.Repeat(" ", width)
		} else {
			attr = config.ColorList
			sel := w.items[ii].IsSelected(w.selected)

			if sel || ii == w.cursor {

				if sel && ii == w.cursor {
					attr = config.ColorCursorSelected
				} else if sel {
					attr = config.ColorListSelected
				} else { // ii == w.cursor
					attr = config.ColorCursor
				}
			}
			s = w.items[ii].Format(width)
		}

		w.window.AttrOn(attr)
		w.window.MovePrint(i, 0, s)
		w.window.AttrOff(attr)
	}

	w.window.Refresh()
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
