package texteditor

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

var (
	screen     tcell.Screen
	x, y       int               = 0, 0
	prevX      int               = 0
	maxX, maxY int               = 0, 0
	buffer     []byte            = []byte("test\ntesttest\n      tester5\n\n\n testing once more\n")
	decoder    *encoding.Decoder = charmap.ISO8859_1.NewDecoder()

	lineStart, lineEnd int = 0, 0
	lines              [][]rune
	lineLengths        []int

	gap *GapBuffer

	updateLines bool = false

	defStyle tcell.Style = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
)

func Start(buf []byte) {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		panic(err.Error())
	}

	err = screen.Init()
	if err != nil {
		panic(err.Error())
	}

	gap = CreateGapBuffer(buf, 2000)
	lines = gap.GetLines(1, gap.GetLineCount())

	for _, v := range lines {
		lineLengths = append(lineLengths, len(v))
	}

	screen.SetStyle(defStyle)

	for i, v := range lines {
		var rest []rune = nil
		if len(v) > 1 {
			rest = v[1:]
		}
		screen.SetContent(0, i, v[0], rest, defStyle)
	}

	lineEnd = len(lines)

	maxX, maxY = screen.Size()

	screen.SetCursorStyle(tcell.CursorStyleBlinkingBlock)
	screen.ShowCursor(x, y)

	for {
		screen.Show()

		ev := screen.PollEvent()

		switch event := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()
			_, maxY = screen.Size()
		case *tcell.EventKey:
			debugInfo()
			switch {
			case event.Key() == tcell.KeyCtrlC:
				screen.Fini()
				return
			case event.Key() == tcell.KeyUp:
				if y != 0 {
					y--
					gap.MoveUp()

					if prevX < 0 {
						prevX = x
					}

					if prevX < lineLengths[y] {
						x = prevX
					} else {
						if lineLengths[y] == 1 {
							x = 0
						} else {
							x = lineLengths[y]
						}
					}
				} else if y == 0 && x > 0 {
					x = 0
					gap.MoveUp()
					prevX = 0
				}
			case event.Key() == tcell.KeyDown:
				if y != maxY && y < len(lineLengths)-1 {
					y++
					gap.MoveDown()

					if prevX < 0 {
						prevX = x
					}

					if prevX < lineLengths[y] {
						x = prevX
					} else {
						x = lineLengths[y] - 1
					}
				} else if y == len(lineLengths)-1 && x < lineLengths[y] {
					x = lineLengths[y] - 1
					prevX = x
					gap.MoveDown()
				}
			case event.Key() == tcell.KeyLeft:
				if x == 0 && y > 0 {
					prevX = -1
					y--
					x = lineLengths[y]
					gap.MoveLeft()
				} else if x != 0 {
					prevX = -1
					x--
					gap.MoveLeft()
				}
			case event.Key() == tcell.KeyRight:
				if (lineLengths[y] == 1 || x == lineLengths[y]) && y < len(lineLengths)-1 {
					prevX = -1
					x = 0
					y++
					gap.MoveRight()
				} else if x < lineLengths[y] {
					prevX = -1
					x++
					gap.MoveRight()
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
				prevX = -1
				insert(rn)
				lineLengths[y]++
				x++
				prevX = x
			}
		}
		update()
		debugInfo()
	}
}

func insert(ch rune) {
	gap.Insert(ch)
	updateLines = true
}

func update() {
	if updateLines {
		lines = gap.GetLines(lineStart, lineEnd)
		tmp := make([]int, lineEnd)
		copy(tmp, lineLengths)
		lineLengths = tmp
		for i, j := lineStart, 0; i <= lineEnd && j < len(lines); i, j = i+1, j+1 {
			lineLengths[i] = len(lines[j])
		}

		updateLines = false

		updateBufferDisplay()
	}

	screen.ShowCursor(x, y)
}

func debugInfo() {
	chars := gap.Chars()
	gapStart, gapEnd := gap.GapStartAndEnd()
	lastMove := gap.LastMove()
	prevX := gap.PrevX()
	gap.CalcPrevLineStart()
	gap.CalcNextLineStart()
	prevLine, nextLine, followingLine, linePos := gap.LinePositions()

	debugX := maxX - 50
	debugY := maxY - 6

	prvXRunes := []rune(fmt.Sprintf("prvx: %03d", prevX))
	lnPosRunes := []rune(fmt.Sprintf("lnp: %03d", linePos))
	prvRunes := []rune(fmt.Sprintf("prv: %03d", prevLine))
	nxtRunes := []rune(fmt.Sprintf("nxt: %03d", nextLine))
	flwRunes := []rune(fmt.Sprintf("flw: %03d", followingLine))
	lmRunes := []rune(fmt.Sprintf("lm: %03d", lastMove))
	crRunes := append([]rune("cr:"), gap.NewLineBehindCursor(), ' ', ' ', ' ')
	chRunes := []rune(fmt.Sprintf("ch: %03d", chars))
	stRunes := []rune(fmt.Sprintf("st: %03d", gapStart))
	enRunes := []rune(fmt.Sprintf("en: %03d", gapEnd))

	screen.SetContent(debugX, debugY, lmRunes[0], lmRunes[1:], defStyle)
	screen.SetContent(debugX, debugY+1, crRunes[0], crRunes[1:], defStyle)
	screen.SetContent(debugX, debugY+2, chRunes[0], chRunes[1:], defStyle)
	screen.SetContent(debugX, debugY+3, stRunes[0], stRunes[1:], defStyle)
	screen.SetContent(debugX, debugY+4, enRunes[0], enRunes[1:], defStyle)
	screen.SetContent(debugX, debugY+5, prvRunes[0], prvRunes[1:], defStyle)
	screen.SetContent(debugX+10, debugY+5, nxtRunes[0], nxtRunes[1:], defStyle)
	screen.SetContent(debugX+20, debugY+5, flwRunes[0], flwRunes[1:], defStyle)
	screen.SetContent(debugX+30, debugY+5, lnPosRunes[0], lnPosRunes[1:], defStyle)
	screen.SetContent(debugX+38, debugY+5, prvXRunes[0], prvXRunes[1:], defStyle)
}

func updateBufferDisplay() {
	for i, v := range lines {
		var rest []rune = nil
		if len(v) > 1 {
			rest = v[1:]
		}
		screen.SetContent(0, i, v[0], rest, defStyle)
	}
}
