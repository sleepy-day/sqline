package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/gdamore/tcell/v2/termbox"
	"github.com/sleepy-day/sqline/database"
)

type dbButton struct {
	Label    string
	Value    string
	ViewName string
}

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

func dbButtons() []dbButton {
	return []dbButton{
		{Label: "Postgres", Value: "postgres", ViewName: "psql_button"},
		{Label: "Sqlite", Value: "sqlite3", ViewName: "sqlite_button"},
	}
}

func connInput(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	}
}

func connSwitchSelection(g *gocui.Gui, _ *gocui.View) error {
	if s_mode != m_connect {
		return nil
	}

	view := g.CurrentView()
	index := -1
	for i, v := range connViews {
		if view.Name() == v {
			index = i
			break
		}
	}

	if index < 0 {
		panic("Invalid view switch on connection view.")
	}

	index++
	if index == len(connViews) {
		g.SetCurrentView(connViews[0])
		return nil
	}

	g.SetCurrentView(connViews[index])
	return nil
}

func unknownView(err error) bool {
	return !errors.Is(err, gocui.ErrUnknownView)
}

func openAddView(g *gocui.Gui, _ *gocui.View) error {
	if s_mode != m_normal {
		return nil
	}

	s_mode = m_connect

	maxX, maxY := g.Size()
	x, y := scale(0.2, maxX), scale(0.25, maxY)
	x2, y2 := scale(0.8, maxX), scale(0.75, maxY)
	if _, err := g.SetView("add_database", x, y, x2, y2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
	}

	x, y = scale(0.21, maxX), scale(0.30, maxY)
	x2, y2 = scale(0.50, maxX), scale(0.30, maxY)+2
	if v, err := g.SetView("driver_label", x, y, x2, y2, 0); err != nil {
		if unknownView(err) {
			return err
		}

		v.Frame = false
		fmt.Fprintf(v, "Database Driver:")
	}

	x, y = scale(0.22, maxX), scale(0.35, maxY)
	x2, y2 = scale(0.30, maxX), scale(0.35, maxY)+2
	for _, v := range dbButtons() {
		view, err := g.SetView(v.ViewName, x, y, x2, y2, 0)
		if err != nil && unknownView(err) {
			return err
		}

		view.SelBgColor = gocui.Attribute(termbox.ColorWhite)
		view.SelFgColor = gocui.Attribute(termbox.ColorBlack)
		fmt.Fprintf(view, spaceText(v.Label, x2-x))

		x = x2 + 1
		x2 += scale(0.08, maxX) + 1
	}

	if err := g.SetKeybinding("driver_label", '1', gocui.ModNone, psqlSelect); err != nil {
		return err
	}

	if err := g.SetKeybinding("driver_label", '2', gocui.ModNone, sqliteSelect); err != nil {
		return err
	}

	x, y = scale(0.21, maxX), scale(0.40, maxY)
	x2, y2 = scale(0.50, maxX), scale(0.40, maxY)+2
	if v, err := g.SetView("name_label", x, y, x2, y2, 0); err != nil {
		if unknownView(err) {
			return err
		}

		v.Frame = false
		fmt.Fprintf(v, "Name:")
	}

	x, y = scale(0.22, maxX), scale(0.43, maxY)
	x2, y2 = scale(0.58, maxX), scale(0.43, maxY)+2
	if v, err := g.SetView("name_input", x, y, x2, y2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Editable = true
		v.Editor = gocui.EditorFunc(connInput)
	}

	x, y = scale(0.21, maxX), scale(0.48, maxY)
	x2, y2 = scale(0.50, maxX), scale(0.48, maxY)+2
	if v, err := g.SetView("conn_str_label", x, y, x2, y2, 0); err != nil {
		if unknownView(err) {
			return err
		}

		v.Frame = false
		fmt.Fprintf(v, "Connection String:")
	}

	x, y = scale(0.22, maxX), scale(0.53, maxY)
	x2, y2 = scale(0.58, maxX), scale(0.53, maxY)+2
	if v, err := g.SetView("connect_input", x, y, x2, y2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Editable = true
		v.Editor = gocui.EditorFunc(connInput)
	}

	x, y = scale(0.32, maxX), scale(0.65, maxY)
	x2, y2 = scale(0.72, maxX), scale(0.65, maxY)+2
	if _, err := g.SetView("conn_status", x, y, x2, y2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
	}

	x, y = scale(0.25, maxX), scale(0.58, maxY)
	x2, y2 = scale(0.33, maxX), scale(0.58, maxY)+2
	if v, err := g.SetView("test_button", x, y, x2, y2, 0); err != nil {
		if unknownView(err) {
			return err
		}

		fmt.Fprintf(v, "Test")
	}

	x = x2 + 1
	x2 += scale(0.08, maxX) + 1
	if v, err := g.SetView("save_button", x, y, x2, y2, 0); err != nil {
		if unknownView(err) {
			return err
		}

		fmt.Fprint(v, "Save")
	}

	if err := g.SetKeybinding("test_button", gocui.KeyEnter, gocui.ModNone, testConnection); err != nil {
		return err
	}

	if err := g.SetKeybinding("save_button", gocui.KeyEnter, gocui.ModNone, saveConnection); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, connSwitchSelection); err != nil {
		return err
	}

	_, err := g.SetCurrentView("driver_label")
	return err
}

func psqlSelect(g *gocui.Gui, v *gocui.View) error {
	buttons := []string{
		"psql_button",
		"sqlite_button",
	}

	for _, val := range buttons {
		view, err := g.View(val)
		if err != nil {
			return err
		}

		if val != "psql_button" {
			view.Highlight = false
			continue
		}

		view.Highlight = true
	}

	driver = "postgres"
	return nil
}

func sqliteSelect(g *gocui.Gui, _ *gocui.View) error {
	buttons := []string{
		"psql_button",
		"sqlite_button",
	}

	for _, val := range buttons {
		view, err := g.View(val)
		if err != nil {
			return err
		}

		if val != "sqlite_button" {
			view.Highlight = false
			continue
		}

		view.Highlight = true
	}

	driver = "sqlite3"
	return nil
}

func testConnection(g *gocui.Gui, _ *gocui.View) error {
	status, err := g.View("conn_status")
	if err != nil {
		return err
	}
	status.Clear()

	connStrInput, err := g.View("connect_input")
	if err != nil {
		return err
	}

	lines := connStrInput.BufferLines()
	if len(lines) == 0 {
		fmt.Fprintf(status, "Invalid input")
		return nil
	}

	err = database.TestConnection(driver, lines[0])
	if err != nil {
		fmt.Fprintf(status, err.Error())
		return nil
	}

	fmt.Fprintf(status, "Test connection successful.")
	return nil
}

func saveConnection(g *gocui.Gui, _ *gocui.View) error {
	status, err := g.View("conn_status")
	if err != nil {
		return err
	}
	status.Clear()

	connStrInput, err := g.View("connect_input")
	if err != nil {
		return err
	}

	lines := connStrInput.BufferLines()
	if len(lines) == 0 || lines[0] == "" {
		fmt.Fprintf(status, "Invalid connection string input")
		return nil
	}

	connStr := lines[0]

	nameInput, err := g.View("name_input")
	if err != nil {
		return err
	}

	lines = nameInput.BufferLines()
	if len(lines) == 0 || lines[0] == "" {
		fmt.Fprintf(status, "Invalid name input")
		return nil
	}

	name := lines[0]

	for _, v := range savedConns.Conns {
		if strings.Compare(v.Driver, driver) != 0 && strings.Compare(v.ConnStr, connStr) != 0 {
			fmt.Fprintf(status, "Connection already exists as [%s]", v.Name)
			return nil
		}
	}

	savedConns.Conns = append(savedConns.Conns, ConnInfo{Name: name, Driver: driver, ConnStr: connStr})

	saveConns()

	return nil
}

func startConnection(g *gocui.Gui, v *gocui.View) error {
	status, err := g.View("conn_status")
	if err != nil {
		return err
	}

	lines := v.BufferLines()
	if len(lines) == 0 {
		fmt.Fprintf(status, "Invalid input")
		return nil
	}

	activeDb = database.CreatePg()
	status.Clear()

	err = activeDb.Initialize(lines[0])
	if err != nil {
		fmt.Fprintf(status, err.Error())
		return nil
	}

	drv, connStr := activeDb.Info()
	for _, v := range savedConns.Conns {
		if strings.Compare(v.Driver, drv) != 0 && strings.Compare(v.ConnStr, connStr) != 0 {
			fmt.Fprintf(status, "Connection already exists as [%s]", v.Name)
			return nil
		}
	}

	fmt.Fprintf(status, "Test connection successful.")
	return nil
}

func spaceText(text string, width int) string {
	length := len([]rune(text))
	if width == length {
		return text
	}

	spaces := width - length
	for {
		text += " "
		spaces--
		if spaces == 0 {
			return text
		}

		text = " " + text
		spaces--
		if spaces == 0 {
			return text
		}
	}
}
