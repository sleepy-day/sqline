package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/gdamore/tcell/v2/termbox"

	. "github.com/sleepy-day/sqline/shared"
)

type Button struct {
	name    string
	label   string
	value   string
	x, y, w int
	fn      KeybindFunc
}

func NewButton(name, label, value string, x, y, w int, fn KeybindFunc) *Button {
	Assert(name != "", "NewButton(): name empty")
	Assert(label != "", "NewButton(): label empty")
	Assert(value != "", "NewButton(): value empty")

	if w == 0 {
		w = len(label) + 1
	}

	return &Button{
		name:  name,
		label: label,
		value: value,
		x:     x,
		y:     y,
		w:     w,
		fn:    fn,
	}
}

func (btn *Button) Layout(g *gocui.Gui) (*gocui.View, error) {
	view, err := g.SetView(btn.name, btn.x, btn.y, btn.x+btn.w, btn.y+2, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}

	view.SelBgColor = gocui.Attribute(termbox.ColorWhite)
	view.SelFgColor = gocui.Attribute(termbox.ColorBlack)
	fmt.Fprintf(view, btn.label)

	if btn.fn == nil {
		return view, nil
	}

	err = g.SetKeybinding(btn.name, gocui.KeyEnter, gocui.ModNone, btn.fn)
	return view, err
}

func (btn *Button) CleanUp(g *gocui.Gui) {
	g.DeleteView(btn.name)
}
