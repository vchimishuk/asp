package main

import (
	"github.com/vchimishuk/asp/config"
	"github.com/vchimishuk/asp/format"
)

type TitleWindow struct {
	panel *PanelWindow
	fmtr  format.Formatter
}

func NewTitleWindow(w, y, x int) (*TitleWindow, error) {
	panel, err := NewPanelWindow(w, y, x)
	if err != nil {
		return nil, err
	}
	panel.SetColor(config.ColorTitle)

	return &TitleWindow{
		panel: panel,
		fmtr:  format.NewFormatter(" {%p}"),
	}, nil
}

func (w *TitleWindow) Update(data map[string]string) {
	w.panel.SetText(w.fmtr.Format(data, w.panel.Width()))
}
