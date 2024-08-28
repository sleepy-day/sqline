package components

import (
	"github.com/gdamore/tcell/v2"
)

type InfoBox struct {
	left, top, right, bottom int
	message                  []rune
	style                    *tcell.Style
}

func CreateInfoBox(left, top, right, bottom int, style *tcell.Style) *InfoBox {
	return &InfoBox{
		left:   left,
		top:    top,
		right:  right,
		bottom: bottom,
		style:  style,
	}
}

func (ibox *InfoBox) Render(screen tcell.Screen) {
	if len(ibox.message) == 0 {
		return
	}

	x, y := ibox.left, ibox.top

	for i := 0; i < len(ibox.message); i++ {
		if x >= ibox.right {
			x = ibox.left
			y++
		}

		if y >= ibox.bottom {
			break
		}

		screen.SetContent(x, y, ibox.message[i], nil, *ibox.style)

		x++
	}
}

func (ibox *InfoBox) SetMessage(msg string) {
	ibox.message = []rune(msg)
}

func (ibox *InfoBox) Reset() {
	ibox.message = []rune{}
}
