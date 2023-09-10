package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Enter key.Binding
	C     key.Binding
	Quit  key.Binding
	CtrlR key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.C, k.Enter, k.CtrlR}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.C, k.Enter}, // first column
	}
}

var Keys = KeyMap{
	C: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "checkout branch"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open issue"),
	),
	CtrlR: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctlr+r", "refresh"),
	),
}
