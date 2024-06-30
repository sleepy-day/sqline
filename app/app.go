package app

import (
	"errors"
	"log"

	"github.com/awesome-gocui/gocui"
	"github.com/sleepy-day/sqline/database"
)

var (
	// Current state of the application
	s_mode int = m_normal

	savedConns *Connections
	activeDb   database.Database
	driver     string
)

const (
	m_normal = iota
	m_insert
	m_connect
)

type Sqline struct {
}

func mainViews() []string {
	return []string{
		"databases",
		"schemas",
		"tables",
		"cmdline",
		"editor",
		"line_numbers",
	}
}

func Start() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	g.InputEsc = true
	g.SetManagerFunc(layout)

	err = keybindings(g)
	if err != nil {
		log.Panicln(err)
	}

	savedConns = loadConns()

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'a', gocui.ModNone, openAddView); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'i', gocui.ModNone, insertMode); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, normalMode); err != nil {
		return err
	}

	return nil
}

func insertMode(g *gocui.Gui, _ *gocui.View) error {
	if s_mode != m_normal {
		return nil
	}

	s_mode = m_insert
	_, err := g.SetCurrentView("editor")
	return err
}

func connectMode(_ *gocui.Gui, _ *gocui.View) error {
	if s_mode == m_normal {
		s_mode = m_connect
	}

	return nil
}

func normalMode(g *gocui.Gui, _ *gocui.View) error {
	if s_mode == m_connect {
		g.DeleteView("add_database")
	}

	g.SetCurrentView("no_selection")
	s_mode = m_normal
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
