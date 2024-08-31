package views

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	comp "github.com/sleepy-day/sqline/components"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
)

const (
	driverRadio NCVSelected = iota
	nameInput
	connStrInput
	testButton
	saveButton
)

var (
	driverTypes []comp.ListItem[string] = []comp.ListItem[string]{
		{Label: []rune("Postgres"), Value: "postgres"},
		{Label: []rune("Sqlite"), Value: "sqlite3"},
		{Label: []rune("MySql"), Value: "mysql"},
		{Label: []rune("MSSql"), Value: "mssql"},
	}
)

type TestFunc func(connStr, driver string) error
type SaveFunc func(name, connStr, driver string)
type NCVSelected byte

type NewConnView struct {
	left, top, right, bottom int
	height, width            int
	selected                 NCVSelected
	style, hlStyle           *tcell.Style
	window                   *comp.Window
	driverRadio              *comp.RadioSelect
	nameInput                *comp.TextBox
	connStrInput             *comp.TextBox
	testButton               *comp.Button
	saveButton               *comp.Button
	infoBox                  *comp.InfoBox
	testFunc                 TestFunc
	saveFunc                 SaveFunc
}

func CreateNewConnView(left, top, right, bottom int, style, hlStyle *tcell.Style, testFunc TestFunc, saveFunc SaveFunc) *NewConnView {
	ncView := &NewConnView{
		left:     left,
		top:      top,
		right:    right,
		bottom:   bottom,
		height:   top - bottom,
		width:    right - left,
		style:    style,
		hlStyle:  hlStyle,
		testFunc: testFunc,
		saveFunc: saveFunc,
		selected: driverRadio,
	}

	ncView.window = comp.CreateWindow(left, top, right, bottom, 2, 2, true, true, []rune("Add New Connection"), style)

	// TODO: Error handling for there being enough space in the window
	inpLeft, inpTop, inpRight, inpBottom := ncView.window.RequestRows(4)
	ncView.driverRadio = comp.CreateRadioSelect(inpLeft, inpTop, inpRight, inpBottom, []rune("Driver:"), driverTypes, style, hlStyle)

	inpLeft, inpTop, inpRight, _ = ncView.window.RequestRows(4)
	ncView.nameInput = comp.CreateTextBox(inpLeft, inpTop, inpRight, []rune("Name:"), style)

	inpLeft, inpTop, inpRight, _ = ncView.window.RequestRows(4)
	ncView.connStrInput = comp.CreateTextBox(inpLeft, inpTop, inpRight, []rune("Connection String:"), style)

	inpLeft, inpTop, inpRight, _ = ncView.window.RequestRows(3)
	ncView.testButton = comp.CreateButton(inpLeft, inpTop, []rune("Test"), style, hlStyle)

	inpLeft += 7
	ncView.saveButton = comp.CreateButton(inpLeft, inpTop, []rune("Save"), style, hlStyle)

	inpLeft, inpTop, inpRight, inpBottom = ncView.window.RequestRows(3)
	ncView.infoBox = comp.CreateInfoBox(inpLeft, inpTop, inpRight, inpBottom, style)

	return ncView
}

func (ncv *NewConnView) Render(screen tcell.Screen) {
	ncv.window.Render(screen)
	ncv.driverRadio.Render(screen)
	ncv.nameInput.Render(screen)
	ncv.connStrInput.Render(screen)
	ncv.testButton.Render(screen)
	ncv.saveButton.Render(screen)
	ncv.infoBox.Render(screen)
}

func (ncv *NewConnView) HandleInput(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyTab {
		switch ncv.selected {
		case driverRadio:
			ncv.selected = nameInput
		case nameInput:
			ncv.selected = connStrInput
		case connStrInput:
			ncv.selected = testButton
		case testButton:
			ncv.selected = saveButton
		case saveButton:
			ncv.selected = nameInput
		}
		return
	}

	switch {
	case ncv.selected == driverRadio:
		ncv.driverRadio.HandleInput(ev)
	case ncv.selected == nameInput:
		ncv.nameInput.HandleInput(ev)
	case ncv.selected == connStrInput:
		ncv.connStrInput.HandleInput(ev)
	case ncv.selected == testButton && ev.Key() == tcell.KeyEnter:
		connStr := ncv.connStrInput.GetString()
		if connStr == "" {
			ncv.infoBox.SetMessage("Connection string is empty")
			break
		}

		driver := ncv.driverRadio.GetSelection()
		if driver == "" {
			ncv.infoBox.SetMessage("No driver selected")
			break
		}

		err := ncv.testFunc(connStr, driver)
		if err != nil {
			ncv.infoBox.SetMessage(fmt.Sprintf("Error: %s", err.Error()))
		} else {
			ncv.infoBox.SetMessage("Test Successful")
		}
	case ncv.selected == saveButton && ev.Key() == tcell.KeyEnter:
		name := ncv.nameInput.GetString()
		if name == "" {
			ncv.infoBox.SetMessage("Name field is empty")
			break
		}

		connStr := ncv.connStrInput.GetString()
		if connStr == "" {
			ncv.infoBox.SetMessage("Connection string is empty")
			break
		}

		driver := ncv.driverRadio.GetSelection()
		if driver == "" {
			ncv.infoBox.SetMessage("No driver selected")
			break
		}

		ncv.saveFunc(name, connStr, driver)
		ncv.infoBox.SetMessage("Connection saved")
	}
}

func (ncv *NewConnView) Reset() {
	ncv.driverRadio.Reset()
	ncv.nameInput.Reset()
	ncv.connStrInput.Reset()
	ncv.testButton.Reset()
	ncv.saveButton.Reset()
	ncv.infoBox.Reset()
}
