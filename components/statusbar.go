package components

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

type StatusBar struct {
	left, top, right   int
	width, popUpHeight int
	status             []rune
	info               []rune
	style, statusStyle *tcell.Style
	mu                 *sync.Mutex
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
		style:       style,
		statusStyle: statusStyle,
		mu:          &sync.Mutex{},
	}
}

func (sb *StatusBar) Render(screen tcell.Screen) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	statusWidth := len(sb.status) + 1
	infoStart := 16

	for i := range sb.width {
		statusLeft := sb.left + i

		if i < 1 || i == statusWidth {
			screen.SetContent(statusLeft, sb.top, ' ', nil, *sb.statusStyle)
			continue
		} else if i < len(sb.status)+1 {
			screen.SetContent(statusLeft, sb.top, sb.status[i-1], nil, *sb.statusStyle)
			continue
		}

		if i >= infoStart && i-infoStart < len(sb.info) {
			screen.SetContent(statusLeft, sb.top, sb.info[i-infoStart], nil, *sb.style)
			continue
		}

		screen.SetContent(statusLeft, sb.top, ' ', nil, *sb.style)
	}
}

func (sb *StatusBar) SetStatus(status []rune, style tcell.Style) {
	sb.statusStyle = &style
	sb.status = status
}

func (sb *StatusBar) SetInfo(msg []rune) {
	sb.info = msg
}

func (sb *StatusBar) ClearInfo() {
	sb.info = []rune{}
}

func (sb *StatusBar) SetErr(msg []rune) {
	sb.info = msg
	go func() {
		time.Sleep(15 * time.Second)
		sb.mu.Lock()
		defer sb.mu.Unlock()
		sb.info = []rune{}
	}()
}
