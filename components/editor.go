package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/sleepy-day/sqline/util"
)

var (
	tabWidth       int = len(tabs) - 1
	indent         int = 0
	twoSpacedTab       = []rune("  ")
	fourSpacedTab      = []rune("    ")
	eightSpacedTab     = []rune("         ")
	tabs               = fourSpacedTab
)

type editorMode byte

type ExecSQLFunc func([]rune) error

const (
	normal editorMode = iota
	insert
	visual
)

type Editor struct {
	gap *util.GapBuffer

	left, top               int
	right, bottom           int
	height, width           int
	innerLeft, innerRight   int
	innerTop, innerBottom   int
	innerHeight, innerWidth int
	curX, curY              int
	hlStartPos, hlEndPos    int
	hlStartLn, hlEndLn      int

	lineOffset    int
	lines         [][]rune
	lineLengths   []int
	prevX         int
	tabsBehind    int
	refreshScreen bool

	style, hlStyle *tcell.Style
	execSQLFunc    ExecSQLFunc
	mode           editorMode
}

func CreateEditor(left, top, right, bottom int, buf []byte, style, hlStyle *tcell.Style) *Editor {
	var gap *util.GapBuffer
	if buf != nil {
		gap, _ = util.CreateGapBuffer(buf, 4000)
	} else {
		// TODO: Error handling
		gap, _ = util.CreateGapBuffer(nil, 4000)
	}

	editor := &Editor{
		left:        left,
		top:         top,
		right:       right,
		bottom:      bottom,
		innerLeft:   left + 1,
		innerTop:    top + 1,
		innerRight:  right - 1,
		innerBottom: bottom - 1,
		style:       style,
		hlStyle:     hlStyle,
		gap:         gap,
		lineOffset:  0,
		hlStartPos:  -1,
		hlStartLn:   -1,
		hlEndLn:     -1,
		hlEndPos:    -1,
		mode:        normal,
	}

	editor.innerWidth = editor.innerRight - editor.innerLeft
	editor.innerHeight = editor.innerBottom - editor.innerTop
	editor.gap.ShiftGap(0)
	editor.lines = gap.GetLines(0, editor.innerHeight)
	editor.lineLengths = make([]int, len(editor.lines))

	for i, v := range editor.lines {
		editor.lineLengths[i] = len(v) - 1
	}

	return editor
}

func (edit *Editor) savePrevX() {
	if edit.prevX != -1 {
		return
	}

	if edit.atEndOfLine() {
		edit.prevX = -2
	} else {
		edit.prevX = edit.curX
	}
}

func (edit *Editor) InNormalMode() bool {
	return edit.mode == normal
}

func (edit *Editor) HandleInput(ev tcell.Event) {
InputSwitch:
	switch event := ev.(type) {
	case *tcell.EventKey:
		if edit.mode == visual {
			switch {
			case event.Key() == tcell.KeyUp:
				edit.moveUp()
			case event.Key() == tcell.KeyDown:
				edit.moveDown()
			case event.Key() == tcell.KeyLeft:
				edit.moveLeft()
			case event.Key() == tcell.KeyRight:
				edit.moveRight()
			case event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModShift:
				edit.execSQL()
			case event.Key() == tcell.KeyEsc:
				edit.mode = normal
				edit.hlStartPos = -1
				edit.hlEndPos = -1
				edit.hlStartLn = -1
				edit.hlEndLn = -1

				break InputSwitch
			}

			edit.hlEndPos = edit.curX
			edit.hlEndLn = edit.curY + edit.lineOffset
		}

		if edit.mode == normal {
			switch {
			case event.Key() == tcell.KeyUp:
				edit.moveUp()
			case event.Key() == tcell.KeyDown:
				edit.moveDown()
			case event.Key() == tcell.KeyLeft:
				edit.moveLeft()
			case event.Key() == tcell.KeyRight:
				edit.moveRight()
			case event.Rune() == 'i':
				edit.mode = insert
				break InputSwitch
			case event.Rune() == 'V' && edit.mode == normal:
				edit.mode = visual
				edit.hlStartPos = edit.curX
				edit.hlStartLn = edit.curY + edit.lineOffset
				edit.hlEndPos = edit.curX
				edit.hlEndLn = edit.curY + edit.lineOffset
			}
		}

		if edit.mode == insert {
			switch {
			case event.Key() == tcell.KeyUp:
				edit.moveUp()
			case event.Key() == tcell.KeyDown:
				edit.moveDown()
			case event.Key() == tcell.KeyLeft:
				edit.moveLeft()
			case event.Key() == tcell.KeyRight:
				edit.moveRight()
			case event.Key() == tcell.KeyEnter:
				edit.insertNewLine()
			case event.Key() == tcell.KeyTab:
				edit.insertTab()
			case event.Key() == tcell.KeyBackspace2 || event.Key() == tcell.KeyBackspace:
				if !edit.atStartOfFile() {
					edit.delete(true)
				}
			case event.Key() == tcell.KeyDelete:
				if !edit.atEndOfFile() {
					edit.delete(false)
				}
			case event.Key() == tcell.KeyEsc:
				edit.mode = normal
				break InputSwitch
			default:
				edit.insertChar(event.Rune())
			}
		}
	}
}

func (edit *Editor) execSQL() {
	if edit.execSQLFunc == nil {
		return
	}

	text, err := edit.gap.GetTextInRange(
		util.Pos{Line: edit.hlStartLn, Col: edit.hlStartPos},
		util.Pos{Line: edit.hlEndLn, Col: edit.hlEndLn},
	)

	if err != nil {
		// TODO: get error handling func
		return
	}

	err = edit.execSQLFunc(text)
	if err != nil {
		// here too
		return
	}
}

func (edit *Editor) insertChar(ch rune) {
	if ch <= 0 {
		return
	}

	edit.insert(ch)
	edit.curX++
	edit.move(true)
}

func (edit *Editor) highlight() {

}

func (edit *Editor) highlightLine() {

}

func (edit *Editor) insertTab() {
	edit.insert('\t')
	edit.tabsBehind++
	edit.curX++
	edit.move(true)
}

func (edit *Editor) insertNewLine() {
	edit.insert('\n')

	if edit.atEndOfView() {
		edit.curY++
	} else {
		edit.lineOffset++
	}

	edit.curX = 0
	edit.tabsBehind = 0

	edit.move(true)
}

func (edit *Editor) moveDown() {
	if edit.atEndOfFile() {
		return
	}

	edit.savePrevX()

	if edit.canMoveDown() {
		edit.curY++
		edit.setCursorToPrevX()
		edit.move(false)
		edit.tabsBehind = edit.gap.TabsBehind()
		return
	}

	if edit.canScrollDown() {
		edit.lineOffset++
		edit.setCursorToPrevX()
		edit.move(true)
		edit.tabsBehind = edit.gap.TabsBehind()
		return
	}

	edit.curX = edit.lineLengths[edit.curY]

	edit.move(false)
	edit.tabsBehind = edit.gap.TabsBehind()
	edit.prevX = -1
}

func (edit *Editor) moveUp() {
	if edit.atStartOfFile() {
		return
	}

	edit.savePrevX()

	if edit.canScrollUp() {
		edit.lineOffset--
		edit.setCursorToPrevX()
		edit.move(true)
		edit.tabsBehind = edit.gap.TabsBehind()
		return
	}

	if edit.canMoveUp() {
		edit.curY--
		edit.setCursorToPrevX()
		edit.move(false)
		edit.tabsBehind = edit.gap.TabsBehind()
		return
	}

	if !edit.atStartOfLine() {
		edit.curX = 0
		edit.move(false)
		edit.prevX = -1
		edit.tabsBehind = 0
	}
}

func (edit *Editor) moveLeft() {
	if edit.atStartOfFile() {
		return
	}

	edit.prevX = -1

	if edit.atStartOfLine() && edit.canMoveUp() {
		edit.curY--
		edit.curX = edit.lineLengths[edit.curY] - 1
		edit.move(true)
		edit.tabsBehind = edit.gap.TabsBehind()
		return
	}

	if edit.atStartOfLine() && edit.lineOffsetSet() {
		edit.lineOffset--
		edit.updateLines()
		edit.move(true)
		edit.curX = edit.lineLengths[edit.curY] - 1
		edit.tabsBehind = edit.gap.TabsBehind()
		return
	}

	if edit.gap.PeekBehind() == '\t' {
		edit.tabsBehind--
	}

	edit.curX--
	edit.move(true)
}

func (edit *Editor) moveRight() {
	if edit.atEndOfFile() {
		return
	}

	edit.prevX = -1

	//	panic(fmt.Sprintf("end: %v move: %v scroll: %v", edit.atEndOfLine(), edit.canMoveDown(), edit.canScrollDown()))

	if edit.atEndOfLine() && edit.canMoveDown() {
		edit.curY++
		edit.curX = 0
		edit.move(true)
		edit.tabsBehind = 0
		return
	}

	if edit.atEndOfLine() && edit.canScrollDown() {
		edit.lineOffset++
		edit.curX = 0
		edit.updateLines()
		edit.move(true)
		edit.tabsBehind = 0
		return
	}

	if !edit.atEndOfLine() {
		edit.curX++
		edit.move(true)

		if edit.gap.PeekBehind() == '\t' {
			edit.tabsBehind++
		}
	}

}

func (edit *Editor) setCursorToPrevX() {
	if edit.prevX == -1 {
		return
	}

	lineLength := edit.lineLengths[edit.curY] - 1
	if edit.onLastLine() {
		lineLength++
	}

	if edit.prevX == -2 {
		edit.curX = lineLength
	} else if edit.prevX >= 0 && lineLength < edit.prevX {
		edit.curX = lineLength
	} else {
		edit.curX = edit.prevX
	}
}

func (edit *Editor) atStartOfFile() bool {
	return edit.curX == 0 && edit.curY+edit.lineOffset == 0
}

func (edit *Editor) atEndOfFile() bool {
	if len(edit.lineLengths) == 0 {
		return true
	}

	return edit.curX == edit.lineLengths[edit.curY]+1 && edit.curY+edit.lineOffset == edit.gap.Lines()
}

func (edit *Editor) atStartOfLine() bool {
	return edit.curX == 0
}

func (edit *Editor) atEndOfLine() bool {
	if len(edit.lineLengths) == 0 {
		return true
	} else if edit.curY == len(edit.lineLengths)-1 {
		return edit.curX == edit.lineLengths[edit.curY]
	}

	return edit.curX == edit.lineLengths[edit.curY]-1
}

func (edit *Editor) canScrollDown() bool {
	return edit.curY+edit.lineOffset < edit.gap.Lines()
}

func (edit *Editor) canScrollUp() bool {
	return edit.curY == 0 && edit.lineOffset > 0
}

func (edit *Editor) lineOffsetSet() bool {
	return edit.lineOffset > 0
}

func (edit *Editor) canMoveUp() bool {
	return edit.curY > 0
}

func (edit *Editor) canMoveDown() bool {
	return edit.curY < edit.innerHeight && edit.curY < len(edit.lineLengths)-1
}

func (edit *Editor) atEndOfView() bool {
	return edit.curY < edit.innerHeight
}

func (edit *Editor) onLastLine() bool {
	if edit.gap.Lines() == 0 {
		return true
	}

	return edit.curY+edit.lineOffset == edit.gap.Lines()
}

func (edit *Editor) updateLines() {
	edit.lines = edit.gap.GetLines(edit.lineOffset, edit.lineOffset+edit.innerHeight)
	edit.lineLengths = make([]int, len(edit.lines))
	for i, j := edit.lineOffset, 0; i <= edit.lineOffset+edit.innerHeight && j < len(edit.lines); i, j = i+1, j+1 {
		edit.lineLengths[j] = len(edit.lines[j])

	}

	edit.refreshScreen = true
}

func (edit *Editor) move(refresh bool) {
	offset, _ := edit.gap.FindOffset(util.Pos{Line: edit.lineOffset + edit.curY, Col: edit.curX})
	edit.gap.ShiftGap(offset)

	if refresh {
		edit.updateLines()
	}
}

func (edit *Editor) Sync(maxX, maxY int) {
	edit.right, edit.bottom = maxX, maxY
}

func (edit *Editor) GetCursorPos() (x, y int) {
	return edit.curX, edit.curY
}

func (edit *Editor) delete(backwards bool) {
	if edit.atStartOfFile() && backwards {
		return
	} else if edit.atEndOfFile() && !backwards {
		return
	}

	calcTabs := false
	prevLength := 0
	if backwards && edit.gap.PeekBehind() == '\t' {
		edit.tabsBehind--
	} else if backwards && edit.gap.PeekBehind() == '\n' {
		calcTabs = true
		prevLength = edit.lineLengths[edit.curY-1] - 1
	}

	edit.gap.Delete(backwards)
	edit.updateLines()

	switch {
	case backwards && edit.atStartOfLine() && edit.canMoveUp():
		edit.curY--
		edit.curX = prevLength
	case backwards && edit.atStartOfLine() && edit.canScrollUp():
		edit.lineOffset--
		edit.curX = prevLength
	case backwards:
		edit.curX--
	}

	edit.move(true)

	if calcTabs {
		edit.tabsBehind = edit.gap.TabsBehind()
	}
}

func (edit *Editor) insert(ch rune) {
	edit.gap.Insert(ch, util.Pos{Line: edit.lineOffset + edit.curY, Col: edit.curX})
}

func (edit *Editor) Render(screen tcell.Screen) {
	screen.ShowCursor(edit.innerLeft+edit.curX+(edit.tabsBehind*(tabWidth)), edit.top+edit.curY+1)

	if edit.refreshScreen {
		for i := range edit.innerWidth {
			for j := range edit.innerHeight + 1 {
				screen.SetContent(edit.innerLeft+i, edit.innerTop+j, ' ', nil, *edit.style)
			}
		}

		edit.refreshScreen = false
	}

	for i := range edit.right - edit.left + 1 {
		if i == 0 {
			screen.SetContent(edit.left, edit.top, tcell.RuneULCorner, nil, *edit.style)
			screen.SetContent(edit.left, edit.bottom, tcell.RuneLLCorner, nil, *edit.style)
			continue
		} else if i == edit.right-edit.left {
			screen.SetContent(edit.left+i, edit.top, tcell.RuneURCorner, nil, *edit.style)
			screen.SetContent(edit.left+i, edit.bottom, tcell.RuneLRCorner, nil, *edit.style)
			continue
		}
		screen.SetContent(edit.left+i, edit.top, tcell.RuneHLine, nil, *edit.style)
		screen.SetContent(edit.left+i, edit.bottom, tcell.RuneHLine, nil, *edit.style)
	}

	for i := range edit.bottom - edit.top {
		if i == 0 {
			continue
		}
		screen.SetContent(edit.left, edit.top+i, tcell.RuneVLine, nil, *edit.style)
		screen.SetContent(edit.right, edit.top+i, tcell.RuneVLine, nil, *edit.style)
	}

	for i, v := range edit.lines {
		spaces := 0

		for j, ch := range v {
			style := edit.style

			col := j
			line := i + edit.lineOffset

			startLn := edit.hlStartLn
			startPos := edit.hlStartPos
			endLn := edit.hlEndLn
			endPos := edit.hlEndPos

			if startLn > endLn || (startLn == endLn && startPos > endPos) {
				startLn = endLn
				startPos = endPos

				endLn = edit.hlStartLn
				endPos = edit.hlStartPos - 1
				if endPos == 0 && endLn > 0 {
					endPos = edit.lineLengths[edit.curY]
					endLn--
				}
			}

			switch {
			case line == startLn && line == endLn && col >= startPos && col <= endPos:
				fallthrough
			case line == startLn && col >= startPos:
				fallthrough
			case line > startLn && line < endLn:
				fallthrough
			case line > startLn && line == endLn && col <= endPos:
				style = edit.hlStyle
			}

			if ch == '\t' {
				screen.SetContent(edit.innerLeft+j, edit.innerTop+i, tabs[0], tabs[1:], *style)
				spaces += tabWidth
				continue
			}
			screen.SetContent(edit.innerLeft+j, edit.innerTop+i, ch, nil, *style)
		}
	}

	lineLength := []rune("lnlength: X")
	if len(edit.lineLengths) > 0 {
		lineLength = []rune(fmt.Sprintf("lnlength: %d", edit.lineLengths[edit.curY]))
	}
	lineCount := []rune(fmt.Sprintf("lncnt: %d", edit.gap.Lines()))
	xPos := []rune(fmt.Sprintf("xPos: %d", edit.curX))
	yPos := []rune(fmt.Sprintf("yPos: %d", edit.curY))
	atEOL := []rune(fmt.Sprintf("atEOL: %v", edit.atEndOfLine()))
	atSOL := []rune(fmt.Sprintf("atSOL: %v", edit.atStartOfLine()))
	atEOF := []rune(fmt.Sprintf("atEOF: %v", edit.atEndOfFile()))
	atSOF := []rune(fmt.Sprintf("atSOF: %v", edit.atStartOfFile()))
	mvDown := []rune(fmt.Sprintf("canMvDown: %v", edit.canMoveDown()))
	mvUp := []rune(fmt.Sprintf("canMvUp: %v", edit.canMoveUp()))
	lastLn := []rune(fmt.Sprintf("lastLn: %v", edit.onLastLine()))

	left := edit.innerRight - 15
	top := edit.innerBottom - 11
	screen.SetContent(left, top, lineCount[0], lineCount[1:], *edit.hlStyle)
	screen.SetContent(left, top+1, xPos[0], xPos[1:], *edit.hlStyle)
	screen.SetContent(left, top+2, yPos[0], yPos[1:], *edit.hlStyle)
	screen.SetContent(left, top+3, lineLength[0], lineLength[1:], *edit.hlStyle)
	screen.SetContent(left, top+4, atEOL[0], atEOL[1:], *edit.hlStyle)
	screen.SetContent(left, top+5, atSOL[0], atSOL[1:], *edit.hlStyle)
	screen.SetContent(left, top+6, atEOF[0], atEOF[1:], *edit.hlStyle)
	screen.SetContent(left, top+7, atSOF[0], atSOF[1:], *edit.hlStyle)
	screen.SetContent(left, top+8, mvDown[0], mvDown[1:], *edit.hlStyle)
	screen.SetContent(left, top+9, mvUp[0], mvUp[1:], *edit.hlStyle)
	screen.SetContent(left, top+10, lastLn[0], lastLn[1:], *edit.hlStyle)
}

func (edit *Editor) SetSQLFunc(fn ExecSQLFunc) {
	edit.execSQLFunc = fn
}

func (edit *Editor) ClearSQLFunc() {
	edit.execSQLFunc = nil
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
