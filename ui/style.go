package ui

import "github.com/charmbracelet/lipgloss"

var (
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C1C6B2")).
			Background(lipgloss.Color("#353533")).
			Padding(0, 2)

	StatusStyle = lipgloss.NewStyle().
			Inherit(StatusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)
)
