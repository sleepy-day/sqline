package components

import (
	"github.com/gdamore/tcell/v2"
)

type TableDataFunc func([][][]rune, []rune)

type Table struct {
	expanded                 bool
	window, popUpWindow      *Window
	style                    *tcell.Style
	hlStyle                  tcell.Style
	left, top, right, bottom int
	maxWidth                 int
	rowCount                 int
	sCol, sRow               int
	anchorCol                int
	lastAnchorCol            int
	anchorRow                int
	lastAnchorRow            int
	colWidths                []int
	data                     [][][]rune
	resultMsg                []rune
	prepared                 bool
	popUpScroll              int
	currentCell              []rune
	popUpWidth               int
	scroll                   bool
}

func CreateTable(left, top, right, bottom, maxWidth int, data [][][]rune, style *tcell.Style) *Table {
	t := &Table{
		left:      left,
		top:       top,
		right:     right,
		bottom:    bottom,
		maxWidth:  maxWidth,
		rowCount:  (bottom - top - 2) / 2,
		expanded:  false,
		sCol:      0,
		sRow:      0,
		anchorCol: 0,
		style:     style,
		hlStyle:   tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite),
		window:    CreateWindow(left, top, right, bottom, 0, 0, true, nil, style),
	}

	if len(data) > 0 && len(data[0]) > 0 {
		t.data = data
		t.colWidths = make([]int, len(data[0]))

		totalWidth := 0
		for col := range t.data[0] {

			width := 0
			for row := range t.data {
				if len(t.data[row][col]) > t.maxWidth {
					width = t.maxWidth
					break
				}

				if len(t.data[row][col]) > width {
					width = len(t.data[row][col])
				}
			}

			t.colWidths[col] = width
			totalWidth += width
		}

		if totalWidth > right-left {
			t.scroll = true
		}
	}

	t.popUpWindow = CreateWindow(t.left, t.top, t.right, t.bottom, 0, 0, true, nil, style)
	t.popUpWidth = t.popUpWindow.GetUsableWidth()

	if bottom-top%2 != 0 {
		t.rowCount++
	}

	return t
}

func (t *Table) HandleInput(ev tcell.Event) {
	if len(t.data) == 0 {
		return
	}

EventLoop:
	switch event := ev.(type) {
	case *tcell.EventKey:
		switch {
		case event.Key() == tcell.KeyUp:
			if t.expanded {
				if t.popUpScroll > 0 {
					t.popUpScroll--
				}

				break
			}

			if t.sRow <= 0 {
				t.sRow = 0
				break EventLoop
			}

			t.sRow--
			if t.sRow < t.anchorRow {
				t.anchorRow--
			}
		case event.Key() == tcell.KeyDown:
			if t.expanded {
				if (t.popUpScroll * t.popUpWidth) < len(t.currentCell) {
					t.popUpScroll++
				}

				break
			}

			if t.sRow >= len(t.data)-1 || t.expanded {
				break EventLoop
			}

			t.sRow++
			if t.sRow >= t.anchorRow+t.rowCount {
				t.anchorRow++
			}
		case event.Key() == tcell.KeyLeft:
			if t.sCol <= 0 {
				t.sCol = 0
				break EventLoop
			} else if t.expanded {
				break EventLoop
			}

			t.sCol--
			if t.sCol < t.anchorCol {
				t.anchorCol--
				t.lastAnchorCol--
			}
		case event.Key() == tcell.KeyRight:
			if t.sCol >= len(t.data[0])-1 || t.expanded {
				break EventLoop
			}
			t.sCol++
			if t.sCol >= t.lastAnchorCol-1 && t.lastAnchorCol < len(t.data[0]) {
				t.lastAnchorCol++
				t.anchorCol++
			}
		case event.Key() == tcell.KeyEnter:
			t.currentCell = t.data[t.sRow][t.sCol]
			t.popUpScroll = 0
			t.expanded = true
		case event.Key() == tcell.KeyEsc:
			t.expanded = false
		}
	}
}

func (t *Table) Render(screen tcell.Screen) {
	t.window.Render(screen)
	if len(t.data) == 0 {
		return
	}

	lastAnchorCol := 0
	colLines := make([]int, len(t.data[0]))
	width := 0
	finalCol := false

	colAdjust := 0
	padOffset := 0
	if t.sCol == len(t.data[0])-1 && t.scroll {
		colAdjust = 1
		padOffset = -10
	}

	colWidth := 0
	cols := 0
	for _, v := range t.colWidths[t.anchorCol:] {
		cols++
		colWidth += v

		if colWidth > t.right-t.left {
			break
		}
	}

	for i := t.anchorCol + colAdjust; i < t.anchorCol+cols && i < len(t.data[0]) && !finalCol; i++ {

		finalRow := false
		rowLeft := t.left + width + padOffset + 1
		if i == t.anchorCol+colAdjust {
			rowLeft++
		}
		for j, rowPos := t.anchorRow, 0; j < t.anchorRow+t.rowCount && !finalRow; j, rowPos = j+1, rowPos+1 {
			for u, ch := range t.data[j][i] {
				if rowLeft+u == t.right {
					finalCol = true
					break
				}

				if t.top+(rowPos*2)+1 > t.bottom {
					finalRow = true
				}

				if u <= t.maxWidth {
					if j == t.sRow && i == t.sCol {
						screen.SetContent(rowLeft+u, t.top+(rowPos*2)+1, ch, nil, t.hlStyle)
					} else {
						screen.SetContent(rowLeft+u, t.top+(rowPos*2)+1, ch, nil, *t.style)
					}
				} else {
					break
				}
			}
		}

		width += t.colWidths[i] + 4
		colLines[i] = t.colWidths[i] + 4
		if t.lastAnchorCol == 0 {
			lastAnchorCol++
		}
	}

	if t.lastAnchorCol == 0 {
		t.lastAnchorCol = lastAnchorCol
	}

	// TODO: clean up this mess
	rangeLen := len(colLines[t.anchorCol:t.lastAnchorCol])
	borderWidth := 0
	for lnNo, w := range colLines[t.anchorCol : cols+t.anchorCol] {
		for j := range w {
			borderLeft := t.left + padOffset + borderWidth + j
			last := j == w-1
			lastColSel := t.sCol == len(t.data[0])-1
			colAnchoredFirst := t.anchorCol == 0
			rowAnchoredFirst := t.anchorRow == 0
			lastColVisible := t.lastAnchorCol >= len(t.data[0])-1
			lastRowVisible := t.anchorRow+((t.bottom-t.top)/2) >= len(t.data)
			lnNoAtStart := lnNo == 0
			lastLnNo := lnNo == rangeLen-1
			endOfRange := j == t.right-t.left-2

			for i := range t.bottom - t.top {
				even := i%2 == 0

				if last && even && lastColSel && lastLnNo && lastColVisible && !colAnchoredFirst {
					switch {
					case i == 0:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneURCorner, nil, *t.style)
					case i == t.bottom-t.top-1:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneLRCorner, nil, *t.style)
					}

					continue
				}

				if last && even && i == 0 {
					switch {
					case rowAnchoredFirst && !lastLnNo:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneTTee, nil, *t.style)
					case rowAnchoredFirst && lastLnNo:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneURCorner, nil, *t.style)
					case lastColVisible && lastLnNo:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneRTee, nil, *t.style)
					default:
						screen.SetContent(borderLeft, t.top+i, tcell.RunePlus, nil, *t.style)
					}

					continue
				}

				if last && even && lastRowVisible && i == t.bottom-t.top-1 {
					if lastColVisible && lastLnNo {
						screen.SetContent(borderLeft, t.top+i, tcell.RuneLRCorner, nil, *t.style)
					} else {
						screen.SetContent(borderLeft, t.top+i, tcell.RuneBTee, nil, *t.style)
					}

					continue
				}

				if last && even && i == t.bottom-t.top-1 {
					switch {
					case !lastRowVisible && lastColVisible && lastLnNo:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneRTee, nil, *t.style)
					case !lastRowVisible:
						screen.SetContent(borderLeft, t.top+i, tcell.RunePlus, nil, *t.style)
					default:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneBTee, nil, *t.style)
					}

					continue
				}

				if last && even {
					switch {
					case lnNo < rangeLen-1:
						screen.SetContent(borderLeft, t.top+i, tcell.RunePlus, nil, *t.style)
					case colAnchoredFirst && lastColVisible:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneRTee, nil, *t.style)
					default:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneLTee, nil, *t.style)
					}

					continue
				}

				if last {
					screen.SetContent(borderLeft, t.top+i, tcell.RuneVLine, nil, *t.style)
					continue
				}

				if lastColVisible && endOfRange && even {
					switch {
					case i == 0:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneURCorner, nil, *t.style)
					case i == t.bottom-t.top-1:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneLRCorner, nil, *t.style)
					default:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneRTee, nil, *t.style)
					}
					continue
				} else if lastColVisible && endOfRange {
					screen.SetContent(borderLeft, t.top+i, tcell.RuneVLine, nil, *t.style)
					continue
				}

				if even && lnNoAtStart && j == 0 && i == 0 {
					switch {
					case colAnchoredFirst && rowAnchoredFirst:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneULCorner, nil, *t.style)
					case rowAnchoredFirst && !colAnchoredFirst:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneTTee, nil, *t.style)
					case !colAnchoredFirst && !rowAnchoredFirst:
						screen.SetContent(borderLeft, t.top+i, tcell.RunePlus, nil, *t.style)
					case colAnchoredFirst && !rowAnchoredFirst:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneLTee, nil, *t.style)
					case !colAnchoredFirst:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneTTee, nil, *t.style)
					}
					continue
				}

				if even && lnNoAtStart && j == 0 {
					switch {
					case i == t.bottom-t.top-1 && !colAnchoredFirst && !lastRowVisible:
						screen.SetContent(borderLeft, t.top+i, tcell.RunePlus, nil, *t.style)
					case i == t.bottom-t.top-1 && colAnchoredFirst && !lastRowVisible:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneLTee, nil, *t.style)
					case i == t.bottom-t.top-1 && lastRowVisible && !colAnchoredFirst:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneBTee, nil, *t.style)
					case i == t.bottom-t.top-1:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneLLCorner, nil, *t.style)
					case t.anchorCol == 0:
						screen.SetContent(borderLeft, t.top+i, tcell.RuneLTee, nil, *t.style)
					default:
						screen.SetContent(borderLeft, t.top+i, tcell.RunePlus, nil, *t.style)
					}
					continue
				}

				if lnNo == 0 && j == 0 {
					screen.SetContent(borderLeft, t.top+i, tcell.RuneVLine, nil, *t.style)
					continue
				}

				if even && borderLeft < t.right {
					screen.SetContent(borderLeft, t.top+i, tcell.RuneHLine, nil, *t.style)
				}
			}
		}

		borderWidth += w
	}

	if t.expanded {
		t.popUpWindow.Render(screen)
		left, top, right, bottom := t.popUpWindow.GetUsableDimensions()
		width := right - left
		height := bottom - top - 1

		cell := t.data[t.sRow][t.sCol]
		x, y := left, top

		maxChars := width * height
		if len(cell) > maxChars {

			startChar := t.popUpScroll * width
			for i := startChar; i < len(cell); i++ {
				if x >= right {
					x = left
					y++
				}
				if y >= bottom {
					break
				}

				screen.SetContent(x, y, cell[i], nil, *t.style)

				x++
			}

		} else {

			for _, ch := range cell {
				if x >= right {
					x = left
					y++
				}
				if y >= bottom {
					break
				}

				screen.SetContent(x, y, ch, nil, *t.style)

				x++
			}
		}
	}
}

func (t *Table) TableFunc() TableDataFunc {
	return func(table [][][]rune, resultMsg []rune) {
		t.data = table
		t.resultMsg = resultMsg
	}
}
