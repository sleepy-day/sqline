package app

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/sleepy-day/sqline/components"
	. "github.com/sleepy-day/sqline/shared"
	"github.com/sleepy-day/sqline/texteditor"
)

var (
	defStyle   tcell.Style = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	maxX, maxY int

	editor *texteditor.Editor
	screen tcell.Screen
)

func quit(screen tcell.Screen) {
	maybePanic := recover()
	screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

func Run() {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	err = screen.Init()
	if err != nil {
		panic(err)
	}

	screen.SetStyle(defStyle)
	screen.EnablePaste()
	screen.Clear()
	defer quit(screen)

	maxX, maxY = screen.Size()

	var buf []byte
	if len(os.Args) > 1 {
		buf, err = os.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	if buf == nil {
		buf = []byte{}
	}

	editor = texteditor.CreateEditor(Scale(0.15, maxX), 0, maxX, maxY, buf, &defStyle)

	items := []components.ListItem{
		{Label: []rune("Postgres"), Value: "postgres"},
		{Label: []rune("Sqlite"), Value: "sqlite"},
		{Label: []rune("MySql"), Value: "mysql"},
		{Label: []rune("Sql Server"), Value: "mssql"},
	}

	list := components.CreateList(0, 0, Scale(0.15, maxX), Scale(0.2, maxY), items, &defStyle)

	var ev tcell.Event
	edit := true
	for {
		ev = screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()
			maxX, maxY = screen.Size()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC:
				return
			case tcell.KeyF5:
				edit = !edit
			default:
				if edit {
					editor.HandleInput(ev)
				} else {
					list.HandleInput(ev)
				}
			}
		}
		screen.Fill(' ', defStyle)
		screen.Sync()
		list.Render(screen)
		editor.Render(screen)
		screen.Show()
	}
}

func DrawConnectors(topX, topY, bottomX, bottomY int) {
	screen.SetContent(topX, topY, tcell.RuneTTee, nil, defStyle)
	screen.SetContent(bottomX, bottomY, tcell.RuneLTee, nil, defStyle)
}
