package app

import (
	"os"

	"github.com/gdamore/tcell/v2"
	. "github.com/sleepy-day/sqline/shared"
	"github.com/sleepy-day/sqline/texteditor"
)

var (
	defStyle   tcell.Style = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	maxX, maxY int

	editor *texteditor.Editor
)

func quit(screen tcell.Screen) {
	maybePanic := recover()
	screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

func Run() {
	screen, err := tcell.NewScreen()
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

	var ev tcell.Event
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
			default:
				editor.HandleInput(ev)
			}
		}
		editor.Render(screen)
		screen.Show()
	}
}
