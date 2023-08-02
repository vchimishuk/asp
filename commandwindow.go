package main

import (
	ncurses "github.com/gbin/goncurses"
)

type CommandWindow struct {
	window  *ncurses.Window
	textbox *TextboxWindow
}

func NewCommandWindow(w, y, x int) (*CommandWindow, error) {
	wnd, err := ncurses.NewWindow(1, w, y, x)
	if err != nil {
		return nil, err
	}
	tb, err := NewTextboxWindow(w, y, x)
	if err != nil {
		return nil, err
	}

	return &CommandWindow{window: wnd, textbox: tb}, nil
}

func (w *CommandWindow) Input(prompt string) string {
	return w.textbox.Input(prompt)
}
