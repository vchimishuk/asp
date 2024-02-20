package main

import (
	"fmt"

	"github.com/vchimishuk/asp/config"
)

type MessageWindow struct {
	panel *PanelWindow
}

func NewMessageWindow(w, y, x int) (*MessageWindow, error) {
	panel, err := NewPanelWindow(w, y, x)
	if err != nil {
		return nil, err
	}
	panel.SetColor(config.ColorMessage)

	return &MessageWindow{
		panel: panel,
	}, nil
}

func (w *MessageWindow) Delete() {
	w.panel.Delete()
}

// TODO: Introduce custom fomratter.
func (w *MessageWindow) Update(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	if len(s) > w.panel.Width() {
		s = s[:w.panel.Width()]
	}
	w.panel.SetText(s)
}

func (w *MessageWindow) Clear() {
	w.panel.Clear()
}
