package main

import (
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sleepy-day/sqline/ui"
)

func main() {
	sql := &Sqline{textarea: textarea.New(), infoBar: ui.InitInfoBar()}
	p := tea.NewProgram(sql, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
