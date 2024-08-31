package components

import (
	"github.com/gdamore/tcell/v2"
)

type StatusBar struct {
	left, top, right   int
	width, popUpHeight int
	popUp              bool
	status             []rune
	info               []rune
	style, statusStyle *tcell.Style
}

func CreateStatusBar(left, top, right, popUpHeight int, initStatus []rune, style, statusStyle *tcell.Style) *StatusBar {
	return &StatusBar{
		left:        left,
		top:         top,
		right:       right,
		width:       right - left - 1,
		popUpHeight: popUpHeight,
		status:      initStatus,
		info:        []rune{},
		popUp:       false,
		style:       style,
		statusStyle: statusStyle,
	}
}

func (sb *StatusBar) Render(screen tcell.Screen) {
	if !sb.popUp {
		statusWidth := len(sb.status) + 1

		for i := range sb.width {
			statusLeft := sb.left + i

			if i < 1 || i == statusWidth {
				screen.SetContent(statusLeft, sb.top, ' ', nil, *sb.statusStyle)
				continue
			} else if i < len(sb.status)+1 {
				screen.SetContent(statusLeft, sb.top, sb.status[i-1], nil, *sb.statusStyle)
				continue
			}

			if i > 15 && i-15 < len(sb.info) {
				screen.SetContent(statusLeft, sb.top, sb.info[i-statusWidth-5], nil, *sb.style)
				continue
			}

			screen.SetContent(statusLeft, sb.top, ' ', nil, *sb.style)
		}
	}
}

func (sb *StatusBar) SetStatus(status []rune, style tcell.Style) {
	sb.statusStyle = &style
	sb.status = status
}
