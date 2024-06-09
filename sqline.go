package main

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sleepy-day/sqline/ui"
)

const (
	normalMode = iota
	textMode
	createMode
	quitMode
)

var mode = textMode

type Sqline struct {
	width    int
	height   int
	infoBar  *ui.InfoBar
	textarea textarea.Model
}

func (s *Sqline) Init() tea.Cmd {
	return textarea.Blink
}

func (s *Sqline) textModeProc(message tea.Msg, cmds *[]tea.Cmd) {
	var cmd tea.Cmd
	if !s.textarea.Focused() {
		cmd = s.textarea.Focus()
		*cmds = append(*cmds, cmd)
	}

	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			*cmds = append(*cmds, tea.Quit)
			return
		case "esc":
			s.textarea.Blur()
			mode = normalMode
			return
		}
	}

	s.textarea, cmd = s.textarea.Update(message)
	*cmds = append(*cmds, cmd)
}

func (s *Sqline) normalModeProc(message tea.Msg, cmds *[]tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			*cmds = append(*cmds, tea.Quit)
			return
		case "i":
			mode = textMode
			return
		}
	}
}

func (s *Sqline) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if sizeMsg, ok := message.(tea.WindowSizeMsg); ok {
		s.SetDimensions(sizeMsg.Width, sizeMsg.Height)
	}

	switch mode {
	case normalMode:
		s.normalModeProc(message, &cmds)
	case textMode:
		s.textModeProc(message, &cmds)
	}

	return s, tea.Batch(cmds...)
}

func (s *Sqline) View() string {
	return lipgloss.JoinVertical(lipgloss.Bottom, s.textarea.View(), s.infoBar.RenderKeyMap())
}

func (s *Sqline) SetDimensions(x, y int) {
	barHeight := 1
	s.width, s.height = x, y
	s.infoBar.SetDimensions(x, barHeight)
	s.textarea.SetWidth(x)
	s.textarea.SetHeight(x - barHeight)
}
