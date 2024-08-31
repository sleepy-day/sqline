package components

import "github.com/gdamore/tcell/v2"

type WindowInputPos struct {
	left, top, right, bottom int
}

type Window struct {
	left, top, right, bottom   int
	height, width              int
	style                      *tcell.Style
	usedRows                   int
	horizontalPad, verticalPad int
	border                     bool
	title                      []rune
}

func CreateWindow(left, top, right, bottom, horizontalPad, verticalPad int, border bool, title []rune, style *tcell.Style) *Window {
	return &Window{
		left:          left,
		top:           top,
		right:         right,
		bottom:        bottom,
		height:        bottom - top,
		width:         right - left,
		horizontalPad: horizontalPad + 1,
		verticalPad:   verticalPad + 1,
		style:         style,
		usedRows:      0,
		border:        border,
		title:         title,
	}
}

func (window *Window) RequestRows(n int) (left, top, right, bottom int) {
	if window.top+window.verticalPad+window.usedRows+n <= window.bottom {
		left = window.left + window.horizontalPad
		right = window.right + window.horizontalPad
		top = window.top + window.verticalPad + window.usedRows
		bottom = top + n

		window.usedRows += n
		return
	}

	return -1, -1, -1, -1
}

func (window *Window) Resize(left, top, right, bottom int) {
	window.left = left
	window.top = top
	window.right = right
	window.bottom = bottom
	window.width = right - left
	window.height = bottom - top
}

func (window *Window) GetUsableWidth() int {
	if window.border {
		return window.width
	}

	return window.width - 2
}

func (window *Window) GetUsableHeight() int {
	if window.border {
		return window.height
	}

	return window.height - 2
}

func (window *Window) GetUsableDimensions() (left, top, right, bottom int) {
	left = window.left + 1
	top = window.top + 1
	right = window.right - 1
	bottom = window.bottom - 1
	return
}

func (window *Window) Render(screen tcell.Screen) {
	if !window.border {
		return
	}

	for row := range window.width + 1 {
		rowLeft := row + window.left

		if row == 0 {
			screen.SetContent(rowLeft, window.top, tcell.RuneULCorner, nil, *window.style)
			screen.SetContent(rowLeft, window.bottom, tcell.RuneLLCorner, nil, *window.style)
			continue
		} else if row == window.width {
			screen.SetContent(rowLeft, window.top, tcell.RuneURCorner, nil, *window.style)
			screen.SetContent(rowLeft, window.bottom, tcell.RuneLRCorner, nil, *window.style)
			continue
		}

		if len(window.title) > 0 && row-1 < len(window.title) {
			screen.SetContent(rowLeft, window.top, window.title[row-1], nil, *window.style)
		} else {
			screen.SetContent(rowLeft, window.top, tcell.RuneHLine, nil, *window.style)
		}

		screen.SetContent(rowLeft, window.bottom, tcell.RuneHLine, nil, *window.style)
	}

	for col := range window.height + 1 {
		colTop := col + window.top

		if col == 0 || col == window.height {
			continue
		}

		screen.SetContent(window.left, colTop, tcell.RuneVLine, nil, *window.style)
		screen.SetContent(window.right, colTop, tcell.RuneVLine, nil, *window.style)
	}
}
