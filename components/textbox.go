package components

import (
	"slices"

	"github.com/gdamore/tcell/v2"
)

type TextBox struct {
	focused          bool
	left, top, right int
	cursorPos        int
	style            *tcell.Style
	buf              []rune
	label            []rune
}

func CreateTextBox(left, top, right int, label []rune, style *tcell.Style) *TextBox {
	return &TextBox{
		left:    left,
		top:     top,
		right:   right,
		style:   style,
		buf:     []rune{},
		label:   label,
		focused: false,
	}
}

func (tbox *TextBox) HandleInput(ev *tcell.EventKey) {
	bufEmpty := len(tbox.buf) == 0
	reachedEnd := tbox.cursorPos >= len(tbox.buf)
	reachedStart := tbox.cursorPos == 0

	switch {
	case ev.Key() == tcell.KeyLeft:
		if bufEmpty || reachedStart {
			break
		}
		tbox.cursorPos--
	case ev.Key() == tcell.KeyRight:
		if bufEmpty || reachedEnd {
			break
		}
		tbox.cursorPos++
	default:
		ch := ev.Rune()
		if ch == 0 {
			break
		}

		if reachedEnd || reachedStart && bufEmpty {
			tbox.buf = append(tbox.buf, ch)
			tbox.cursorPos++
			break
		} else if reachedStart {
			tbox.buf = append([]rune{ch}, tbox.buf...)
			tbox.cursorPos++
			break
		}

		tbox.buf = slices.Insert(tbox.buf, tbox.cursorPos, ch)
		tbox.cursorPos++
	}
}

func (tbox *TextBox) Render(screen tcell.Screen) {
	width := tbox.right - tbox.left - 6
	for i := range width {
		if i < len(tbox.label) {
			screen.SetContent(tbox.left+i, tbox.top, tbox.label[i], nil, *tbox.style)
		}

		if i == 0 {
			screen.SetContent(tbox.left, tbox.top+1, tcell.RuneULCorner, nil, *tbox.style)
			screen.SetContent(tbox.left, tbox.top+2, tcell.RuneVLine, nil, *tbox.style)
			screen.SetContent(tbox.left, tbox.top+3, tcell.RuneLLCorner, nil, *tbox.style)
			continue
		} else if i == width-1 {
			screen.SetContent(tbox.left+i, tbox.top+1, tcell.RuneURCorner, nil, *tbox.style)
			screen.SetContent(tbox.left+i, tbox.top+2, tcell.RuneVLine, nil, *tbox.style)
			screen.SetContent(tbox.left+i, tbox.top+3, tcell.RuneLRCorner, nil, *tbox.style)
			continue
		} else if i == 1 {
			if len(tbox.buf) > 1 {
				screen.SetContent(tbox.left+i, tbox.top+2, tbox.buf[0], tbox.buf[1:], *tbox.style)
			} else if len(tbox.buf) > 0 {
				screen.SetContent(tbox.left+i, tbox.top+2, tbox.buf[0], nil, *tbox.style)
			}
		}

		if i >= len(tbox.buf) {
			screen.SetContent(tbox.left+i, tbox.top+2, ' ', nil, *tbox.style)
		}

		screen.SetContent(tbox.left+i, tbox.top+1, tcell.RuneHLine, nil, *tbox.style)
		screen.SetContent(tbox.left+i, tbox.top+3, tcell.RuneHLine, nil, *tbox.style)

		screen.ShowCursor(tbox.left+tbox.cursorPos+1, tbox.top+2)
	}
}

func (tbox *TextBox) GetString() string {
	if tbox.buf == nil {
		return ""
	}

	return string(tbox.buf)
}

func (tbox *TextBox) Reset() {
	tbox.buf = []rune{}
	tbox.cursorPos = 0
}
