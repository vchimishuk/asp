package main

import (
	"strings"
	"unicode/utf8"

	ncurses "github.com/gbin/goncurses"
)

type PanelWindow struct {
	window *ncurses.Window
	text   string
}

func NewPanelWindow(w, y, x int) (*PanelWindow, error) {
	window, err := ncurses.NewWindow(1, w, y, x)
	if err != nil {
		return nil, err
	}

	return &PanelWindow{window: window}, nil
}

func (w *PanelWindow) SetColor(color ncurses.Char) {
	w.window.SetBackground(color)
}

func (w *PanelWindow) SetText(text string) {
	w.text = text
	w.refresh()
}

func (w *PanelWindow) Clear() {
	w.window.Erase()
	w.window.Refresh()
}

func (w *PanelWindow) refresh() {
	w.window.MovePrint(0, 0, w.text)
	l := utf8.RuneCountInString(w.text)
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
