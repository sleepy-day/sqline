package texteditor

import (
	"github.com/gdamore/tcell/v2"
)

type Editor struct {
	gap *GapBuffer

	xPos, yPos int
	maxX, maxY int
	curX, curY int

	firstLn, lastLn int
	lines           [][]rune
	lineLengths     []int
	updateLines     bool
	prevX           int

	style *tcell.Style

	focused bool
}

func CreateEditor(x, y, maxX, maxY int, buf []byte, style *tcell.Style) *Editor {
	var gap *GapBuffer
	if buf != nil {
		gap = CreateGapBuffer(buf, 4000)
	} else {
		gap = CreateGapBuffer([]byte{}, 4000)
	}

	return &Editor{
		xPos:  x,
		yPos:  y,
		maxX:  maxX,
		maxY:  maxY,
		style: style,
		gap:   gap,
	}
}

func (edit *Editor) HandleInput(ev tcell.Event) {
	switch event := ev.(type) {
	case *tcell.EventKey:
		switch {
		case event.Key() == tcell.KeyUp:
			if edit.curY != 0 {
				edit.curY--
				edit.gap.MoveUp()

				if edit.gap.prevX < 0 {
					edit.gap.prevX = edit.curX
				}

				if edit.gap.prevX < edit.lineLengths[edit.curY] {
					edit.curX = edit.gap.prevX
				} else {
					if edit.lineLengths[edit.curY] == 1 {
						edit.curX = 0
					} else {
						edit.curX = edit.lineLengths[edit.curY]
					}
				}
			} else if edit.curY == 0 && edit.curX > 0 {
				edit.curX = 0
				edit.gap.MoveUp()
				edit.gap.prevX = 0
			}
		case event.Key() == tcell.KeyDown:
			if edit.curY != edit.maxY && edit.curY < len(edit.lineLengths)-1 {
				edit.curY++
				edit.gap.MoveDown()

				if edit.gap.prevX < 0 {
					edit.gap.prevX = edit.curX
				}

				if edit.gap.prevX < edit.lineLengths[edit.curY] {
					edit.curX = edit.gap.prevX
				} else if edit.lineLengths[edit.curY] == 1 {
					edit.curX = edit.lineLengths[edit.curY] - 1
				} else {
					edit.curX = edit.lineLengths[edit.curY]
				}
			} else if edit.curY == len(edit.lineLengths)-1 && edit.curX < edit.lineLengths[edit.curY] {
				edit.curX = edit.lineLengths[edit.curY]
				edit.gap.prevX = edit.curX
				edit.gap.MoveDown()
			}
		case event.Key() == tcell.KeyLeft:
			if edit.curX == 0 && edit.curY > 0 {
				edit.gap.prevX = -1
				edit.curY--
				edit.curX = edit.lineLengths[edit.curY]
				edit.gap.MoveLeft()
			} else if edit.curX != 0 {
				edit.gap.prevX = -1
				edit.curX--
				edit.gap.MoveLeft()
			}
		case event.Key() == tcell.KeyRight:
			if (edit.lineLengths[edit.curY] == 1 || edit.curX == edit.lineLengths[edit.curY]) && edit.curY < len(edit.lineLengths)-1 {
				edit.gap.prevX = -1
				edit.curX = 0
				edit.curY++
				edit.gap.MoveRight()
			} else if edit.curX < edit.lineLengths[edit.curY] {
				edit.gap.prevX = -1
				edit.curX++
				edit.gap.MoveRight()
			}
		case event.Key() == tcell.KeyEnter:
			edit.curY++
			edit.curX = 0
			edit.gap.prevX = -1
			edit.insert('\n')
		case event.Key() == tcell.KeyBackspace2:
			if edit.curX > 0 {
				edit.gap.Delete(true)
				edit.curX--
				edit.updateLines = true
			} else if edit.curY > 0 {
				edit.curY--
				edit.curX = edit.lineLengths[edit.curY]
				edit.gap.Delete(true)
				edit.updateLines = true
			}
			break
		case event.Key() == tcell.KeyDelete:
			if edit.curX != edit.lineLengths[edit.curY] && edit.curY != len(edit.lineLengths)-1 {
				edit.gap.Delete(false)
				edit.updateLines = true
			}
		default:
			rn := event.Rune()
			if rn == 0 {
				break
			}
			edit.gap.prevX = -1
			edit.insert(rn)
			edit.lineLengths[edit.curY]++
			edit.curX++
			edit.gap.prevX = edit.curX
		}
	}

	if edit.updateLines {
		edit.lines = edit.gap.GetLines(edit.firstLn, edit.gap.GetLineCount())
		tmp := make([]int, edit.lastLn)
		copy(tmp, edit.lineLengths)
		edit.lineLengths = tmp
		edit.lineLengths = make([]int, len(edit.lines))
		for i, j := edit.firstLn, 0; i <= edit.lastLn && j < len(edit.lines); i, j = i+1, j+1 {
			edit.lineLengths[i] = len(edit.lines[j])
		}

		edit.updateLines = false
	}
}

func (edit *Editor) GetCursorPos() (x, y int) {
	return edit.curX, edit.curY
}

func (edit *Editor) SetFocus() {
	edit.focused = true
}

func (edit *Editor) RemoveFocus() {
	edit.focused = false
}

func (edit *Editor) delete(backwards bool) {
	edit.gap.Delete(backwards)
	edit.updateLines = true
}

func (edit *Editor) insert(ch rune) {
	edit.gap.Insert(ch)
	edit.updateLines = true
}

func (edit *Editor) Render(screen tcell.Screen) {
	screen.Fill(' ', *edit.style)
	screen.Sync()

	for i, v := range edit.lines {
		if len(v) > 1 {
			screen.SetContent(edit.xPos, edit.yPos+i, v[0], v[1:], *edit.style)
		} else {
			screen.SetContent(edit.xPos, edit.yPos+i, v[0], nil, *edit.style)
		}
	}
}

/*
func updateBufferDisplay() {
screen.Fill(' ', defStyle)
screen.Sync()
for i, v := range lines {
for j, ch := range v {
if ch == '\n' {
screen.SetContent(j, i, '$', nil, defStyle)
} else {
screen.SetContent(j, i, ch, nil, defStyle)
}
}
//		if len(v) > 1 {
//			screen.SetContent(0, i, v[0], v[1:], defStyle)
//		} else {
//			screen.SetContent(0, i, v[0], nil, defStyle)
//		}
}
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

*/
