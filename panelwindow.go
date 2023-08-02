package main

import (
	"strings"

	ncurses "github.com/gbin/goncurses"
)

type PanelWindow struct {
	window *ncurses.Window
	// fmtr     format.Formatter
	// position int
	text string
}

func NewPanelWindow(w, y, x int) (*PanelWindow, error) {
	// window := parent.Sub(h, w, y, x)
	window, err := ncurses.NewWindow(1, w, y, x)
	if err != nil {
		return nil, err
	}

	return &PanelWindow{window: window}, nil
	// fmtr:   format.NewFormatter(" {-60%:%t}{*%:(%n%%)} "),
}

func (w *PanelWindow) SetBackground(color ncurses.Char) {
	w.window.SetBackground(color)
}

// TODO: Make formatter to toake map[string]interface{} instead of
//       map[string]string, and let it to do Itoa convertion.
//
// TODO: Create single method Update(map[string]{}interface) instead of
//       SetPanel and SetPosition().
//
// TODO: Split this PanelWindow into separate text and time windows?

func (w *PanelWindow) SetText(text string) {
	w.text = text
	w.refresh()
}

// func (w *PanelWindow) SetPosition(pos int) {
// 	w.position = pos
// 	w.update()
// }

func (w *PanelWindow) refresh() {
	// data := map[string]string{
	// 	"t": w.text,
	// 	"n": strconv.Itoa(w.position)}
	// s := w.fmtr.Format(data, width)
	// w.window.Erase()
	w.window.MovePrint(0, 0, w.text)
	l := len(w.text)
	if l < w.Width() {
		w.window.MovePrint(0, l, strings.Repeat(" ", w.Width()-l))
	}
	w.window.Refresh()
}

func (w *PanelWindow) Width() int {
	_, x := w.window.MaxYX()

	return x
}

func (w *PanelWindow) Height() int {
	y, _ := w.window.MaxYX()

	return y
}
