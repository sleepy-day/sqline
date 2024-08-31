package components

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

type RadioFunc func(ev tcell.Event)

type RadioSelect struct {
	left, top, right, bottom int
	selected                 int
	style, selHlStyle        *tcell.Style
	hlStyle                  tcell.Style
	label                    []rune
	opts                     []ListItem[string]
	radioFn                  RadioFunc
	focus                    bool
}

func CreateRadioSelect(left, top, right, bottom int, label []rune, opts []ListItem[string], style, hlStyle *tcell.Style) *RadioSelect {
	if bottom-top < 4 {
		panic("height for RS is less than 4, not enough space")
	}

	return &RadioSelect{
		left:       left,
		top:        top,
		right:      right,
		bottom:     bottom,
		style:      style,
		selHlStyle: hlStyle,
		hlStyle:    tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack),
		selected:   -1,
		opts:       opts,
		label:      label,
	}
}

func (rs *RadioSelect) Reset() {
	rs.selected = -1
	rs.focus = false
}

func (rs *RadioSelect) HandleInput(ev *tcell.EventKey) {
	key := ev.Rune()
	sel, err := strconv.Atoi(string(key))
	if err != nil {
		return
	}

	if sel > len(rs.opts) || sel <= 0 {
		return
	}

	rs.selected = sel - 1
}

func (rs *RadioSelect) GetSelection() string {
	if rs.selected <= 0 {
		return ""
	}

	return rs.opts[rs.selected].Value
}

func (rs *RadioSelect) Render(screen tcell.Screen) {
	for i, v := range rs.label {
		if rs.focus {
			screen.SetContent(rs.left+i, rs.top, v, nil, rs.hlStyle)
		} else {
			screen.SetContent(rs.left+i, rs.top, v, nil, *rs.style)
		}
	}

	var width int
	for j, opt := range rs.opts {
		radioWidth := rs.left + width

		for i, ch := range opt.Label {
			if i == 0 {
				screen.SetContent(radioWidth+i, rs.top+1, tcell.RuneULCorner, nil, *rs.style)
				screen.SetContent(radioWidth+i, rs.top+2, tcell.RuneVLine, nil, *rs.style)
				screen.SetContent(radioWidth+i, rs.top+3, tcell.RuneLLCorner, nil, *rs.style)

			}

			if rs.selected == j {
				screen.SetContent(radioWidth+i+1, rs.top+2, ch, nil, *rs.selHlStyle)
			} else {
				screen.SetContent(radioWidth+i+1, rs.top+2, ch, nil, *rs.style)
			}
			screen.SetContent(radioWidth+i+1, rs.top+1, tcell.RuneHLine, nil, *rs.style)
			screen.SetContent(radioWidth+i+1, rs.top+3, tcell.RuneHLine, nil, *rs.style)

			if i == len(opt.Label)-1 {
				screen.SetContent(radioWidth+i+2, rs.top+1, tcell.RuneURCorner, nil, *rs.style)
				screen.SetContent(radioWidth+i+2, rs.top+2, tcell.RuneVLine, nil, *rs.style)
				screen.SetContent(radioWidth+i+2, rs.top+3, tcell.RuneLRCorner, nil, *rs.style)
			}
		}

		width += len(opt.Label) + 3
	}
}

func (rs *RadioSelect) Focus() {
	rs.focus = true
}

func (rs *RadioSelect) LoseFocus() {
	rs.focus = false
}
