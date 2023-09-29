package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Right key.Binding
	Left  key.Binding
	Enter key.Binding
	Tab   key.Binding
	P     key.Binding
	C     key.Binding
	CtrlR key.Binding
	Esc   key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.C, k.P, k.Enter, k.CtrlR}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.C, k.Enter}, // first column
	}
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k:", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j:", "move down"),
	),
	Right: key.NewBinding(
		key.WithKeys("left", "h"),
	),
	Left: key.NewBinding(
		key.WithKeys("right", "l"),
	),
	C: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c:", "checkout branch"),
	),
	P: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p:", "open pull request"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter:", "open issue"),
	),
	CtrlR: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctlr+r:", "refresh"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab:", "switch pane"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
	),
}
