package main

import (
	"github.com/vchimishuk/asp/config"
	"github.com/vchimishuk/asp/format"
	"github.com/vchimishuk/chubby"
)

type StatusWindow struct {
	panel       *PanelWindow
	playingFmtr format.Formatter
	pausedFmtr  format.Formatter
	stoppedFmtr format.Formatter
}

func NewStatusWindow(w, y, x int) (*StatusWindow, error) {
	panel, err := NewPanelWindow(w, y, x)
	if err != nil {
		return nil, err
	}
	// TODO: Separate color.
	panel.SetBackground(config.ColorPair(config.ColorTitle))

	return &StatusWindow{
		panel:       panel,
		playingFmtr: format.NewFormatter(" {-*%:%a - %t}{*%:[%o/%l]} "),
		pausedFmtr:  format.NewFormatter(" {-*%:%a - %t}{*%:[%o/%l]} "),
		stoppedFmtr: format.NewFormatter(""),
	}, nil
}

func (w *StatusWindow) Update(state chubby.State, data map[string]string) {
	var fmtr format.Formatter

	if state == chubby.StatePlaying {
		fmtr = w.playingFmtr
	} else if state == chubby.StatePaused {
		fmtr = w.pausedFmtr
	} else if state == chubby.StateStopped {
		fmtr = w.stoppedFmtr
	}

	w.panel.SetText(fmtr.Format(data, w.panel.Width()))
}
