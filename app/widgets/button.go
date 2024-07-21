package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/gdamore/tcell/v2/termbox"

	. "github.com/sleepy-day/sqline/shared"
)

type Button struct {
	name                     string
	label                    string
	value                    string
	x, y, w                  int
	left, top, right, bottom int
	fn                       KeybindFunc
}

func NewButton(name, label, value string, x, y, w int, fn KeybindFunc) *Button {
	Assert(name != "", "NewButton(): name empty")
	Assert(label != "", "NewButton(): label empty")
	Assert(value != "", "NewButton(): value empty")

	if w == 0 {
		w = x + len(label) + Offset
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
	view, err := g.SetView(btn.name, btn.x, btn.y, btn.w, btn.y+2, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}

	view.SelBgColor = gocui.Attribute(termbox.ColorWhite)
	view.SelFgColor = gocui.Attribute(termbox.ColorBlack)
	fmt.Fprintf(view, btn.label)

	btn.left, btn.top, btn.right, btn.bottom = view.Dimensions()

	if btn.fn == nil {
		return view, nil
	}

	err = g.SetKeybinding(btn.name, gocui.KeyEnter, gocui.ModNone, btn.fn)
	if err != nil {
		return nil, err
	}

	return view, nil
}

func (btn *Button) Resize(x, y, w int) {
	if w == 0 {
		w = x + len(btn.label) + Offset
	}
	btn.x, btn.y, btn.w = x, y, w
}

func (btn *Button) CleanUp(g *gocui.Gui) {
	g.DeleteView(btn.name)
}

func (btn *Button) Dimensions() (left, top, right, bottom int) {
	return btn.left, btn.top, btn.right, btn.bottom
}

func (btn *Button) Size() (x, y int) {
	return btn.right - btn.left, btn.bottom - btn.top
}

func (btn *Button) HasLabel() bool {
	return false
}
