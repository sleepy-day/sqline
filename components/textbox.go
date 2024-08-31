package components

import (
	"slices"

	"github.com/gdamore/tcell/v2"
)

type TextBox struct {
	focused          bool
	left, top, right int
	cursorPos        int
	offset           int
	style            *tcell.Style
	hlStyle          tcell.Style
	buf              []rune
	label            []rune
	focus            bool
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
		hlStyle: tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack),
	}
}

func (tbox *TextBox) HandleInput(ev *tcell.EventKey) {
	bufEmpty := len(tbox.buf) == 0
	reachedEnd := tbox.cursorPos >= len(tbox.buf)
	reachedStart := tbox.cursorPos == 0 && tbox.offset == 0

	switch {
	case ev.Key() == tcell.KeyLeft:
		if bufEmpty {
			break
		} else if tbox.cursorPos == 0 && tbox.offset > 0 {
			tbox.offset--
		} else {
			tbox.cursorPos--
		}
	case ev.Key() == tcell.KeyRight:
		if bufEmpty || reachedEnd {
			break
		}
		tbox.cursorPos++
	case ev.Key() == tcell.KeyBackspace2 || ev.Key() == tcell.KeyBackspace:
		if reachedStart {
			break
		}

		if tbox.cursorPos+tbox.offset == len(tbox.buf) {
			tbox.buf = tbox.buf[:len(tbox.buf)-1]
		} else {
			tbox.buf = append(tbox.buf[:tbox.offset+tbox.cursorPos-2], tbox.buf[tbox.offset+tbox.cursorPos-1:]...)
		}

		if tbox.offset > 0 {
			tbox.offset--
		} else {
			tbox.cursorPos--
		}
	default:
		ch := ev.Rune()
		if ch == 0 {
			break
		}

		if reachedEnd || (reachedStart && bufEmpty) {
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

	if tbox.cursorPos > tbox.right-tbox.left {
		tbox.cursorPos--
		tbox.offset++
	}
}

func (tbox *TextBox) Render(screen tcell.Screen) {
	width := tbox.right - tbox.left - 6
	for i := range width {
		if i < len(tbox.label) {
			if tbox.focus {
				screen.SetContent(tbox.left+i, tbox.top, tbox.label[i], nil, tbox.hlStyle)
			} else {
				screen.SetContent(tbox.left+i, tbox.top, tbox.label[i], nil, *tbox.style)
			}
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
		}

		if i+1 < len(tbox.buf)+2 {
			screen.SetContent(tbox.left+i, tbox.top+2, tbox.buf[i-1], nil, *tbox.style)
		} else {
			screen.SetContent(tbox.left+i, tbox.top+2, ' ', nil, *tbox.style)
		}

		screen.SetContent(tbox.left+i, tbox.top+1, tcell.RuneHLine, nil, *tbox.style)
		screen.SetContent(tbox.left+i, tbox.top+3, tcell.RuneHLine, nil, *tbox.style)

	}

	if tbox.focus {
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
	tbox.offset = 0
	tbox.focus = false
}

func (tbox *TextBox) Focus() {
	tbox.focus = true
}

func (tbox *TextBox) LoseFocus() {
	tbox.focus = false
}
