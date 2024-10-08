package components

import "github.com/gdamore/tcell/v2"

type Button struct {
	left, top int
	width     int
	text      []rune
	style     *tcell.Style
	hlStyle   tcell.Style
	focus     bool
}

func CreateButton(left, top int, text []rune, style *tcell.Style) *Button {
	return &Button{
		left:    left,
		top:     top,
		width:   len(text) + 2,
		text:    text,
		style:   style,
		hlStyle: tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack),
	}
}

func (b *Button) Render(screen tcell.Screen) {
	style := b.style
	if b.focus {
		style = &b.hlStyle
	}

	for i := range b.width {
		btnLeft := b.left + i

		if i == 0 {
			screen.SetContent(btnLeft, b.top, tcell.RuneULCorner, nil, *style)
			screen.SetContent(btnLeft, b.top+1, tcell.RuneVLine, nil, *style)
			screen.SetContent(btnLeft, b.top+2, tcell.RuneLLCorner, nil, *style)
		} else if i == b.width-1 {
			screen.SetContent(btnLeft, b.top, tcell.RuneURCorner, nil, *style)
			screen.SetContent(btnLeft, b.top+1, tcell.RuneVLine, nil, *style)
			screen.SetContent(btnLeft, b.top+2, tcell.RuneLRCorner, nil, *style)
		} else {
			screen.SetContent(btnLeft, b.top, tcell.RuneHLine, nil, *style)
			screen.SetContent(btnLeft, b.top+1, b.text[i-1], nil, *style)
			screen.SetContent(btnLeft, b.top+2, tcell.RuneHLine, nil, *style)
		}
	}
}

func (b *Button) Focus() {
	b.focus = true
}

func (b *Button) LoseFocus() {
	b.focus = false
}
