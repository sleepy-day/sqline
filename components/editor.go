package components

import (
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
	hlLine         bool
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

func (edit *Editor) getLineLength() int {
	if len(edit.lineLengths) == 0 {
		return 0
	}

	if edit.curY == len(edit.lineLengths)-1 {
		return edit.lineLengths[edit.curY]
	}

	return edit.lineLengths[edit.curY] - 1
}

func (edit *Editor) moveToLineEnd() {
	if len(edit.lineLengths) == 0 {
		return
	}

	edit.curX = edit.getLineLength()
	edit.prevX = -1
	edit.move(false)
}

func (edit *Editor) moveToLineStart() {
	edit.curX = 0
	edit.prevX = -1
}

func (edit *Editor) HandleInput(ev *tcell.EventKey) {
	if edit.mode == visual {
		switch ev.Key() {
		case tcell.KeyUp:
			edit.moveUp()
		case tcell.KeyDown:
			edit.moveDown()
		case tcell.KeyLeft:
			edit.moveLeft()
		case tcell.KeyRight:
			edit.moveRight()
		case tcell.KeyEnter:
			edit.execSQL()
		case tcell.KeyHome:
			edit.moveToLineStart()
		case tcell.KeyEnd:
			edit.moveToLineEnd()
		case tcell.KeyEsc:
			edit.mode = normal
			edit.hlLine = false
			edit.hlStartPos = -1
			edit.hlEndPos = -1
			edit.hlStartLn = -1
			edit.hlEndLn = -1

			return
		}

		edit.hlEndLn = edit.curY + edit.lineOffset
		if !edit.hlLine {
			edit.hlEndPos = edit.curX
		}
	}

	if edit.mode == normal {
		switch ev.Key() {
		case tcell.KeyUp:
			edit.moveUp()
		case tcell.KeyDown:
			edit.moveDown()
		case tcell.KeyLeft:
			edit.moveLeft()
		case tcell.KeyRight:
			edit.moveRight()
		case tcell.KeyHome:
			edit.moveToLineStart()
		case tcell.KeyEnd:
			edit.moveToLineEnd()
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'i':
				edit.mode = insert
			case 'V':
				if edit.mode != normal {
					break
				}
				edit.mode = visual
				edit.hlLine = true
				edit.hlStartPos = 0
				edit.hlEndPos = edit.getLineLength()
				edit.hlStartLn = edit.curY + edit.lineOffset
				edit.hlEndLn = edit.curY + edit.lineOffset
			case 'v':
				if edit.mode != normal {
					break
				}
				edit.mode = visual
				edit.hlStartPos = edit.curX
				edit.hlStartLn = edit.curY + edit.lineOffset
				edit.hlEndPos = edit.curX
				edit.hlEndLn = edit.curY + edit.lineOffset

			}
		}

		return
	}

	if edit.mode == insert {
		switch ev.Key() {
		case tcell.KeyUp:
			edit.moveUp()
		case tcell.KeyDown:
			edit.moveDown()
		case tcell.KeyLeft:
			edit.moveLeft()
		case tcell.KeyRight:
			edit.moveRight()
		case tcell.KeyEnter:
			edit.insertNewLine()
		case tcell.KeyTab:
			edit.insertTab()
		case tcell.KeyHome:
			edit.moveToLineStart()
		case tcell.KeyEnd:
			edit.moveToLineEnd()
		case tcell.KeyBackspace2, tcell.KeyBackspace:
			if !edit.atStartOfFile() {
				edit.delete(true)
			}
		case tcell.KeyDelete:
			if !edit.atEndOfFile() {
				edit.delete(false)
			}
		case tcell.KeyEsc:
			edit.mode = normal
		default:
			edit.insertChar(ev.Rune())
		}

		return
	}
}

func (edit *Editor) execSQL() {
	if edit.execSQLFunc == nil {
		return
	}

	stLn, stPos := edit.hlStartLn, edit.hlStartPos
	enLn, enPos := edit.hlEndLn, edit.hlEndPos
	if edit.hlLine {
		stPos = 0
		enPos = 999999
		if stLn > enLn {
			stLn = enLn
			enLn = edit.hlStartLn
		}
	} else if stLn == enLn && stPos > enPos {
		stPos = enPos
		enPos = edit.hlStartPos
	} else if stLn > enLn {
		stLn = enLn
		enLn = edit.hlStartLn
		stPos = enPos
		enPos = edit.hlStartPos - 1

		if enPos == 0 && enLn > 0 {
			enPos = edit.getLineLength()
			enLn--
		}
	}

	text, err := edit.gap.GetTextInRange(
		util.Pos{Line: stLn, Col: stPos},
		util.Pos{Line: enLn, Col: enPos},
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

	for row, v := range edit.lines {
		if row > edit.innerHeight {
			break
		}

		spaces := 0
		for col, ch := range v {
			line := row + edit.lineOffset

			startLn := edit.hlStartLn
			startPos := edit.hlStartPos
			endLn := edit.hlEndLn
			endPos := edit.hlEndPos

			if !edit.hlLine && startLn > endLn || (startLn == endLn && startPos > endPos) {
				startLn = endLn
				startPos = endPos

				endLn = edit.hlStartLn
				endPos = edit.hlStartPos - 1
				if endPos == 0 && endLn > 0 {
					endPos = edit.lineLengths[edit.curY]
					endLn--
				}
			}

			style := edit.style
			switch {
			case edit.hlLine && line >= endLn && line <= startLn:
				fallthrough
			case edit.hlLine && line <= endLn && line >= startLn:
				fallthrough
			case line == startLn && line == endLn && col >= startPos && col <= endPos:
				fallthrough
			case line == startLn && col >= startPos && startLn < endLn:
				fallthrough
			case line > startLn && line < endLn:
				fallthrough
			case line > startLn && line == endLn && col <= endPos:
				style = edit.hlStyle
			}

			if ch == '\t' {
				screen.SetContent(edit.innerLeft+col, edit.innerTop+row, tabs[0], tabs[1:], *style)
				spaces += tabWidth
				continue
			}
			screen.SetContent(edit.innerLeft+col, edit.innerTop+row, ch, nil, *style)
		}
	}
}

func (edit *Editor) SetSQLFunc(fn ExecSQLFunc) {
	edit.execSQLFunc = fn
}

func (edit *Editor) ClearSQLFunc() {
	edit.execSQLFunc = nil
}
