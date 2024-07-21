package texteditor

import (
	"slices"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

var (
	screen  tcell.Screen
	x, y    int               = 0, 0
	buffer  []byte            = []byte("test\ntesttest\n      tester5\n\n\n testing once more\n")
	decoder *encoding.Decoder = charmap.ISO8859_1.NewDecoder()

	text [][]rune = make([][]rune, 500)

	defStyle tcell.Style = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
)

func Start() {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		panic(err.Error())
	}

	err = screen.Init()
	if err != nil {
		panic(err.Error())
	}

	screen.SetStyle(defStyle)

	maxX, maxY := screen.Size()

	screen.SetCursorStyle(tcell.CursorStyleBlinkingBlock)
	screen.ShowCursor(x, y)

	for {
		if screen.HasPendingEvent() {
			switch event := screen.PollEvent().(type) {
			case *tcell.EventResize:
				screen.Sync()
				maxX, maxY = screen.Size()
			case *tcell.EventKey:
				switch {
				case event.Key() == tcell.KeyCtrlC:
					screen.Fini()
					return
				case event.Key() == tcell.KeyUp:
					if y != 0 {
						y--
					}
				case event.Key() == tcell.KeyDown:
					if y != maxY && y < len(text) {
						y++
					}
				case event.Key() == tcell.KeyLeft:
					if x != 0 {
						x--
					}
				case event.Key() == tcell.KeyRight:
					if x != maxX && x < len(text[y]) {
						x++
					}
				case event.Key() == tcell.KeyEnter:
					y++
					x = 0
				case event.Key() == tcell.KeyCtrlC:
					screen.Fini()
					return
				case event.Key() == tcell.KeyBackspace:

				default:
					rn := event.Rune()
					if rn == 0 {
						continue
					}
					insert(rn)
					if x == maxX {
						x = 0
						y++
					} else {
						x++
					}
				}
			}
			update()
		}
	}
}

func delete() {
	if x <= len(text[y]) {
		text[y] = slices.Delete(text[y], x-1, x-1)
	}
	x--
}

func insert(ch rune) {
	if x == len(text[y]) {
		text[y] = append(text[y], ch)
		return
	}

	if x < len(text[y]) {
		text[y] = slices.Insert(text[y], x, ch)
	}
}

func update() {
	for yPos, line := range text {
		for xPos, rn := range line {
			screen.SetContent(xPos, yPos, rn, nil, defStyle)
		}
	}
	screen.ShowCursor(x, y)
	screen.Show()
}
