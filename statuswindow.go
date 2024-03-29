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
	panel.SetColor(config.ColorStatus)

	return &StatusWindow{
		panel:       panel,
		playingFmtr: format.NewFormatter(config.FormatStatusPlaying),
		pausedFmtr:  format.NewFormatter(config.FormatStatusPaused),
		stoppedFmtr: format.NewFormatter(""),
	}, nil
}

func (w *StatusWindow) Delete() {
	w.panel.Delete()
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
