package components

import (
	"github.com/gdamore/tcell/v2"
)

type TableDataFunc func([][][]rune, []rune)

type Table struct {
	window, popUpWindow *Window
	style               *tcell.Style
	hlStyle             tcell.Style
	oddRowStyle         tcell.Style
	evenRowStyle        tcell.Style

	left, top, right, bottom     int
	pLeft, pTop, pRight, pBottom int
	popUpScroll                  int
	popUpWidth                   int
	maxWidth                     int
	tableHeight                  int

	sCol, sRow    int
	anchorCol     int
	lastAnchorCol int
	anchorRow     int
	colWidths     []int
	currentCell   []rune

	expanded bool
	prepared bool
	scroll   bool
	refresh  bool

	data      [][][]rune
	resultMsg []rune
}

func CreateTable(left, top, right, bottom, pLeft, pTop, pRight, pBottom, maxWidth int, data [][][]rune, style *tcell.Style) *Table {
	t := &Table{
		left:         left,
		top:          top,
		right:        right,
		bottom:       bottom,
		maxWidth:     maxWidth,
		tableHeight:  bottom - top - 1,
		expanded:     false,
		sCol:         -1,
		sRow:         -1,
		anchorCol:    0,
		style:        style,
		oddRowStyle:  tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorWhite),
		evenRowStyle: tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite),
		hlStyle:      tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite),
		window:       CreateWindow(left, top, right, bottom, 0, 0, true, true, nil, style),
		popUpWindow:  CreateWindow(pLeft, pTop, pRight, pBottom, 0, 0, true, true, nil, style),
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

	t.popUpWidth = t.popUpWindow.GetUsableWidth()

	return t
}

func (t *Table) HandleInput(ev tcell.Event) {
	if len(t.data) == 0 {
		return
	}

	if t.sRow == -1 {
		t.sRow = 0
	}
	if t.sCol == -1 {
		t.sCol = 0
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

			if t.sRow+t.anchorRow >= len(t.data)-1 || t.expanded {
				break EventLoop
			}

			if t.sRow < t.tableHeight-1 {
				t.sRow++
			} else if t.sRow == t.tableHeight-1 && t.sRow+t.anchorRow < len(t.data) {
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
			t.currentCell = t.data[t.sRow+t.anchorRow][t.sCol]
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

	lastAnchorCol, width, finalCol := 0, 0, false
	colSepLines := make([]int, len(t.data[0]))

	colAdjust, padOffset := 0, 0
	if t.sCol == len(t.data[0])-1 && t.scroll {
		colAdjust = 1
		padOffset = -10
	}

	colWidth, cols := 0, 0
	for _, v := range t.colWidths[t.anchorCol:] {
		cols++
		colWidth += v

		if colWidth > t.right-t.left {
			break
		}
	}

	visibleRows := t.data[t.anchorRow:]
	for i := t.anchorCol + colAdjust; i < t.anchorCol+cols && i < len(t.data[0]) && !finalCol; i++ {
		rowLeft := t.left + width + padOffset + 1
		if i == t.anchorCol+colAdjust {
			rowLeft++
		}

		for j := 0; j < len(visibleRows) && j < t.tableHeight; j++ {
			style := t.evenRowStyle
			if j == t.sRow && i == t.sCol {
				style = t.hlStyle
			} else if (j+t.anchorRow)%2 == 1 {
				style = t.oddRowStyle
			}

			for u, ch := range visibleRows[j][i] {
				if rowLeft+u < t.left {
					continue
				}

				if rowLeft+u == t.right {
					finalCol = true
					break
				}

				if u <= t.maxWidth {
					screen.SetContent(rowLeft+u, t.top+j+1, ch, nil, style)
				} else {
					break
				}
			}

			cellLen := len(visibleRows[j][i])
			for k := range t.colWidths[i] - cellLen {
				if rowLeft+cellLen+k < t.left {
					continue
				} else if rowLeft+cellLen+k >= t.right {
					break
				}

				screen.SetContent(rowLeft+cellLen+k, t.top+j+1, ' ', nil, style)
			}
		}

		width += t.colWidths[i] + 4
		colSepLines[i] = t.colWidths[i] + 4
		if t.lastAnchorCol == 0 {
			lastAnchorCol++
		}
	}

	if t.lastAnchorCol == 0 {
		t.lastAnchorCol = lastAnchorCol
	}

	if t.expanded {
		t.popUpWindow.Render(screen)
		left, top, right, bottom := t.popUpWindow.GetUsableDimensions()
		width := right - left
		height := bottom - top - 1

		cell := t.data[t.sRow+t.anchorRow][t.sCol]
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

func (t *Table) lastColSelected() bool {
	if len(t.data) == 0 {
		return false
	}

	return t.sCol == len(t.data[0])-1
}

func (t *Table) colAnchoredFirst() bool {
	return t.anchorCol == 0
}

func (t *Table) rowAnchoredFirst() bool {
	return t.anchorRow == 0
}

func (t *Table) lastColVisible() bool {
	if len(t.data) == 0 {
		return false
	}

	return t.lastAnchorCol >= len(t.data[0])-1
}

func (t *Table) lastRowVisible() bool {
	return t.anchorRow+t.bottom-t.top >= len(t.data)
}

func (t *Table) TableFunc() TableDataFunc {
	return func(table [][][]rune, resultMsg []rune) {
		if table == nil && resultMsg == nil {
			return
		}

		if resultMsg != nil && table == nil {
			table = [][][]rune{
				[][]rune{
					[]rune("Results"),
					resultMsg,
				},
			}
		}

		t.data = table
		t.resultMsg = resultMsg

		t.colWidths = make([]int, len(table[0]))

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

		if totalWidth > t.right-t.left {
			t.scroll = true
		}

		t.refresh = true
		t.sCol = -1
		t.sRow = -1
		t.anchorRow = 0
		t.anchorCol = 0
	}
}
