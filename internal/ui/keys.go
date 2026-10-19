package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Enter    key.Binding
	Merge    key.Binding
	Yes      key.Binding
	No       key.Binding
	Refresh  key.Binding
	Help     key.Binding
	Quit     key.Binding
}

var keys = keyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k")),
	Down:     key.NewBinding(key.WithKeys("down", "j")),
	PageUp:   key.NewBinding(key.WithKeys("pgup", "b")),
	PageDown: key.NewBinding(key.WithKeys("pgdown", "f", " ")),
	Top:      key.NewBinding(key.WithKeys("home", "g")),
	Bottom:   key.NewBinding(key.WithKeys("end", "G")),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open PR in browser"),
	),
	Merge: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "approve and merge PR"),
	),
	Yes: key.NewBinding(key.WithKeys("y")),
	No:  key.NewBinding(key.WithKeys("n", "esc")),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

const helpText = `
  j/k, up/down   move selection
  b/f, pgup/pgdn move one page
  g/G, home/end  go to first/last PR
  enter          open PR in browser
  m              approve and squash-merge PR (asks to confirm)
  r              refresh
  ?              toggle this help
  q, ctrl+c      quit
`
