package main

// TODO: No need to have seaparate TextBox window.
type CommandWindow struct {
	textbox *TextboxWindow
}

func NewCommandWindow(w, y, x int) (*CommandWindow, error) {
	tb, err := NewTextboxWindow(w, y, x)
	if err != nil {
		return nil, err
	}

	return &CommandWindow{textbox: tb}, nil
}

func (w *CommandWindow) Refresh() {
	w.textbox.Refresh()
}

func (w *CommandWindow) Input(prompt string) string {
	return w.textbox.Input(prompt)
}
