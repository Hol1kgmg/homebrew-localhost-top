package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Top     key.Binding
	Bottom  key.Binding
	Search  key.Binding
	Kill    key.Binding
	Force   key.Binding
	Reload  key.Binding
	Detail  key.Binding
	Open    key.Binding
	LANLink key.Binding
	QRCode  key.Binding
	Sort    key.Binding
	Command key.Binding
	Quit    key.Binding
	Help    key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

var keys = keyMap{
	Up:      key.NewBinding(key.WithKeys("k", "up")),
	Down:    key.NewBinding(key.WithKeys("j", "down")),
	Top:     key.NewBinding(key.WithKeys("g")),
	Bottom:  key.NewBinding(key.WithKeys("G")),
	Search:  key.NewBinding(key.WithKeys("/")),
	Kill:    key.NewBinding(key.WithKeys("K")),
	Force:   key.NewBinding(key.WithKeys("X")),
	Reload:  key.NewBinding(key.WithKeys("r")),
	Detail:  key.NewBinding(key.WithKeys("enter", "l")),
	Open:    key.NewBinding(key.WithKeys("o")),
	LANLink: key.NewBinding(key.WithKeys("L")),
	QRCode:  key.NewBinding(key.WithKeys("Q")),
	Sort:    key.NewBinding(key.WithKeys("s")),
	Command: key.NewBinding(key.WithKeys(":")),
	Quit:    key.NewBinding(key.WithKeys("q")),
	Help:    key.NewBinding(key.WithKeys("?")),
	Confirm: key.NewBinding(key.WithKeys("y")),
	Cancel:  key.NewBinding(key.WithKeys("n", "esc")),
}
