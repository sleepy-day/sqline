package components

import (
	"slices"

	"github.com/gdamore/tcell/v2"
)

type TextBox struct {
	left, top, width int
	buf              []rune
	focused          bool
	style            *tcell.Style
	cursorPos        int
}

func CreateTextBox(left, top, width int, style *tcell.Style) *TextBox {
	return &TextBox{
		left:    left,
		top:     top,
		width:   width,
		style:   style,
		buf:     []rune{},
		focused: false,
	}
}

func (tbox *TextBox) HandleInput(ev tcell.Event) {
	bufEmpty := len(tbox.buf) == 0
	reachedEnd := tbox.cursorPos >= len(tbox.buf)
	reachedStart := tbox.cursorPos == 0

	switch event := ev.(type) {
	case *tcell.EventKey:
		switch {
		case event.Key() == tcell.KeyLeft:
			if bufEmpty || reachedStart {
				break
			}
			tbox.cursorPos--
		case event.Key() == tcell.KeyRight:
			if bufEmpty || reachedEnd {
				break
			}
			tbox.cursorPos++
		default:
			ch := event.Rune()
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
}

func (tbox *TextBox) Render(screen tcell.Screen) {
	for i := range tbox.width {
		if i == 0 {
			screen.SetContent(tbox.left, tbox.top, tcell.RuneULCorner, nil, *tbox.style)
			screen.SetContent(tbox.left, tbox.top+1, tcell.RuneVLine, nil, *tbox.style)
			screen.SetContent(tbox.left, tbox.top+2, tcell.RuneLLCorner, nil, *tbox.style)
			continue
		} else if i == tbox.width-1 {
			screen.SetContent(tbox.left+i, tbox.top, tcell.RuneURCorner, nil, *tbox.style)
			screen.SetContent(tbox.left+i, tbox.top+1, tcell.RuneVLine, nil, *tbox.style)
			screen.SetContent(tbox.left+i, tbox.top+2, tcell.RuneLRCorner, nil, *tbox.style)
			continue
		} else if i == 1 {
			if len(tbox.buf) > 1 {
				screen.SetContent(tbox.left+i, tbox.top+1, tbox.buf[0], tbox.buf[1:], *tbox.style)
			} else if len(tbox.buf) > 0 {
				screen.SetContent(tbox.left+i, tbox.top+1, tbox.buf[0], nil, *tbox.style)
			}
		}

		if i >= len(tbox.buf) {
			screen.SetContent(tbox.left+i, tbox.top+1, ' ', nil, *tbox.style)
		}

		screen.SetContent(tbox.left+i, tbox.top, tcell.RuneHLine, nil, *tbox.style)
		screen.SetContent(tbox.left+i, tbox.top+2, tcell.RuneHLine, nil, *tbox.style)

		screen.ShowCursor(tbox.left+tbox.cursorPos+1, tbox.top+1)
	}
}
