package app

import (
	"errors"
	"fmt"
	"log"

	"github.com/awesome-gocui/gocui"
	"github.com/sleepy-day/sqline/database"

	"github.com/gdamore/tcell/v2"
	_ "github.com/gdamore/tcell/v2/encoding"
)

var (
	// Current state of the application
	s_mode int = m_normal

	savedConns *Connections
	activeDb   database.Database

	NewConn  *NewConnPage
	ConnList *ConnListPage
)

const (
	m_normal = iota
	m_insert
	m_connect
	m_connectList
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

func StartAlt() {
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(fmt.Sprintf("%v\n", err))
	}

	err = screen.Init()
	if err != nil {
		panic(fmt.Sprintf("%v\n", err))
	}

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)

	screen.SetStyle(defStyle)

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

	if savedConns == nil {
		panic("eee")
	}

	NewConn = CreateNewConnPage(g, 80, 25)
	ConnList = CreateConnListPage(g, 60, 50, savedConns.Conns)

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'a', gocui.ModNone, openConnPage); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'c', gocui.ModNone, openConnList); err != nil {
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

func openConnList(g *gocui.Gui, _ *gocui.View) error {
	if s_mode != m_normal {
		return nil
	}

	s_mode = m_connectList
	return ConnList.Open(g)
}

func openConnPage(g *gocui.Gui, _ *gocui.View) error {
	if s_mode != m_normal {
		return nil
	}

	s_mode = m_connect
	return NewConn.Open(g)
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

func closePage(g *gocui.Gui) {
	switch s_mode {
	case m_connect:
		NewConn.Close(g)
	case m_connectList:
		ConnList.Close(g)
	}

	return
}

func normalMode(g *gocui.Gui, _ *gocui.View) error {
	closePage(g)

	g.SetCurrentView("no_selection")
	s_mode = m_normal
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
