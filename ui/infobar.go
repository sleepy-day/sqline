package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type InfoBar struct {
	Width  int
	Height int
	Style  lipgloss.Style
	help   help.Model
	keys   keyMap
}

type keyMap struct {
	Up              key.Binding
	Down            key.Binding
	Left            key.Binding
	Right           key.Binding
	Add             key.Binding
	Connect         key.Binding
	OpenConnections key.Binding
	Help            key.Binding
	Quit            key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "up"),
		key.WithHelp("up/k", "Move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "down"),
		key.WithHelp("down/j", "Move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "left"),
		key.WithHelp("left/h -", "Move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "right"),
		key.WithHelp("right/l -", "Move right"),
	),
	Add: key.NewBinding(
		key.WithKeys("a -", "Add Database"),
		key.WithHelp("a -", "Add Database"),
	),
	Connect: key.NewBinding(
		key.WithKeys("c -", "Connect to Database"),
		key.WithHelp("c -", "Connect to Database"),
	),
	OpenConnections: key.NewBinding(
		key.WithKeys("V -", "View Open Connections"),
		key.WithHelp("V -", "View Open Connections"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("? -", "Toggle Help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q -", "Quit"),
	),
}

func InitInfoBar() *InfoBar {
	return &InfoBar{
		Style: StatusBarStyle,
		help: help.Model{
			ShortSeparator: " | ",
			FullSeparator:  " | ",
		},
		keys: keys,
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Connect, k.OpenConnections, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Help, k.Quit},
	}
}

func (i *InfoBar) RenderKeyMap() string {
	return i.Style.Render(i.help.View(i.keys))
}

func (i *InfoBar) SetDimensions(x, y int) {
	i.help.Width = x
	i.Style = i.Style.Width(x).Height(y)
}
