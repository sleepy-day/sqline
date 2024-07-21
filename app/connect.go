package app

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/sleepy-day/sqline/app/widgets"
	"github.com/sleepy-day/sqline/database"
	. "github.com/sleepy-day/sqline/shared"
)

type dbButton struct {
	Label    string
	Value    string
	ViewName string
}

type NewConnPage struct {
	Height, Width   int
	SelectableViews []string
	Modal           *widgets.Modal
	Driver          *widgets.RadioSelect
	Name            *widgets.Input
	ConnStr         *widgets.Input
	TestConn        *widgets.Button
	SaveConn        *widgets.Button
	Info            *widgets.Info
}

// TODO: move init into New functions

func CreateNewConnPage(g *gocui.Gui, width, height int) *NewConnPage {
	page := &NewConnPage{
		Height: height,
		Width:  width,
	}

	guiX, guiY := g.Size()

	left := Scale(0.5, guiX-width)
	right := left + width
	top := Scale(0.5, guiY-height)
	bottom := top + height

	page.Modal = widgets.NewModal("new_conn", "", left, top, right, bottom)

	page.Modal.Layout(g)
	defer page.Modal.CleanUp(g)

	modX, modY := page.Modal.Size()

	left += Scale(0.05, modX)
	top += Scale(0.05, modY)
	page.Driver = widgets.NewRadioSelect(g, "driver_select", "Select a driver:", left, top, driverOpts())

	page.Driver.Layout(g)
	defer page.Driver.CleanUp(g)

	page.SelectableViews = append(page.SelectableViews, "driver_select")

	_, prevY := page.Driver.Size()

	top += prevY
	w := Scale(0.8, width)
	page.Name = widgets.NewInput("new_conn_name", "Name:", left, top, w, nil, nil)

	page.Name.Layout(g)
	defer page.Name.CleanUp(g)

	page.SelectableViews = append(page.SelectableViews, "new_conn_name")

	_, prevY = page.Name.Size()

	top += prevY
	page.ConnStr = widgets.NewInput("conn_str_input", "Connection String:", left, top, w, nil, nil)

	page.ConnStr.Layout(g)
	defer page.ConnStr.CleanUp(g)

	page.SelectableViews = append(page.SelectableViews, "conn_str_input")

	_, prevY = page.ConnStr.Size()

	top += prevY
	left += Scale(0.25, width)
	page.TestConn = widgets.NewButton("test_conn_button", "Test Connection", "test", left, top, 0, page.testConnection())

	page.TestConn.Layout(g)
	defer page.TestConn.CleanUp(g)

	page.SelectableViews = append(page.SelectableViews, "test_conn_button")

	prevX, _ := page.TestConn.Size()
	left += prevX + 2
	page.SaveConn = widgets.NewButton("save_conn_button", "Save", "save", left, top, 0, page.saveConnection())

	page.SaveConn.Layout(g)
	defer page.SaveConn.CleanUp(g)

	page.SelectableViews = append(page.SelectableViews, "save_conn_button")

	_, prevY = page.SaveConn.Size()

	top += prevY + 1
	left -= Scale(0.40, modX)
	right = Scale(0.60, modX)
	page.Info = widgets.NewInfo("info_box", left, top, right)

	g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, SwapSelection)

	return page
}

func (page *NewConnPage) Open(g *gocui.Gui) error {
	_, err := page.Modal.Layout(g)
	if err != nil {
		return err
	}

	driver, err := page.Driver.Layout(g)
	if err != nil {
		return err
	}

	_, err = page.Name.Layout(g)
	if err != nil {
		return err
	}

	_, err = page.ConnStr.Layout(g)
	if err != nil {
		return err
	}

	_, err = page.TestConn.Layout(g)
	if err != nil {
		return err
	}

	_, err = page.SaveConn.Layout(g)
	if err != nil {
		return err
	}

	_, err = page.Info.Layout(g)
	if err != nil {
		return err
	}

	g.SetCurrentView(driver.Name())

	return nil
}

func (page *NewConnPage) Close(g *gocui.Gui) {
	page.Modal.CleanUp(g)
	page.Driver.CleanUp(g)
	page.Name.CleanUp(g)
	page.ConnStr.CleanUp(g)
	page.TestConn.CleanUp(g)
	page.SaveConn.CleanUp(g)
	page.Info.CleanUp(g)
}

func (page *NewConnPage) SwapSelection(g *gocui.Gui) {
	view := g.CurrentView()
	index := -1
	for i, v := range page.SelectableViews {
		if view.Name() == v {
			index = i
			break
		}
	}

	Assert(index >= 0, "SwapSelection(): next view was not found")

	index++
	if index == len(page.SelectableViews) {
		index = 0
	}

	g.SetCurrentView(page.SelectableViews[index])
	return
}

func SwapSelection(g *gocui.Gui, _ *gocui.View) error {
	if s_mode != m_connect {
		return nil
	}

	NewConn.SwapSelection(g)

	return nil
}

func driverOpts() []widgets.Option {
	return []widgets.Option{
		{Name: "psql_select", Label: "Postgres", Value: "postgres"},
		{Name: "sqlite_select", Label: "Sqlite", Value: "sqlite3"},
		{Name: "mysql_select", Label: "MySql", Value: "mysql"},
	}
}

func (page *NewConnPage) testConnection() widgets.KeybindFunc {
	return func(g *gocui.Gui, _ *gocui.View) error {
		status, err := g.View(page.Info.Name())
		if err != nil {
			return err
		}
		status.Clear()

		connStrInput, err := g.View(page.ConnStr.Name())
		if err != nil {
			return err
		}

		lines := connStrInput.BufferLines()
		if len(lines) == 0 {
			fmt.Fprintf(status, "Invalid input")
			return nil
		}

		driver := page.Driver.Selected()

		err = database.TestConnection(driver, lines[0])
		if err != nil {
			fmt.Fprintf(status, err.Error())
			return nil
		}

		fmt.Fprintf(status, "Test connection successful.")
		return nil
	}
}

func (page *NewConnPage) saveConnection() widgets.KeybindFunc {
	return func(g *gocui.Gui, _ *gocui.View) error {
		status, err := g.View(page.Info.Name())
		if err != nil {
			return err
		}
		status.Clear()

		connStrInput, err := g.View(page.ConnStr.Name())
		if err != nil {
			return err
		}

		lines := connStrInput.BufferLines()
		if len(lines) == 0 || lines[0] == "" {
			fmt.Fprintf(status, "Invalid connection string input")
			return nil
		}

		connStr := lines[0]

		nameInput, err := g.View(page.Name.Name())
		if err != nil {
			return err
		}

		lines = nameInput.BufferLines()
		if len(lines) == 0 || lines[0] == "" {
			fmt.Fprintf(status, "Invalid name input")
			return nil
		}

		name := lines[0]
		driver := page.Driver.Selected()

		for _, v := range savedConns.Conns {
			if strings.Compare(v.Driver, driver) != 0 && strings.Compare(v.ConnStr, connStr) != 0 {
				fmt.Fprintf(status, "Connection already exists as [%s]", v.Name)
				return nil
			}
		}

		savedConns.Conns = append(savedConns.Conns, ConnInfo{Name: name, Driver: driver, ConnStr: connStr})
		err = saveConns(savedConns)
		if err != nil {
			panic(err.Error())
		}

		status.Clear()
		fmt.Fprint(status, "Connection saved.")

		return nil
	}
}

func oldstartConnection(g *gocui.Gui, v *gocui.View) error {
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
