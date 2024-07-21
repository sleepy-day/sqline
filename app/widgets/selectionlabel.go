package widgets

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/awesome-gocui/gocui"
	. "github.com/sleepy-day/sqline/shared"
)

type Option struct {
	Name  string
	Label string
	Value string
}

type RadioSelect struct {
	name                     string
	label                    string
	x, y, w                  int
	left, top, right, bottom int
	selected                 string
	options                  []Option
	buttons                  []Button
}

func NewRadioSelect(g *gocui.Gui, name, label string, x, y int, options []Option) *RadioSelect {
	Assert(name != "", "NewRadioSelect(): name empty")
	Assert(label != "", "NewRadioSelect(): name empty")
	Assert(len(options) > 0, "NewRadioSelect(): options length is 0")
	Assert(len(options) <= 9, "NewRadioSelect(): options length greater than 9")

	return &RadioSelect{
		name:    name,
		label:   label,
		x:       x,
		y:       y,
		w:       x + len(label) + Offset,
		options: options,
	}
}

func (rs *RadioSelect) Layout(g *gocui.Gui) (*gocui.View, error) {
	view, err := g.SetView(rs.name, rs.x, rs.y, rs.w, rs.y+MinHeight, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		panic("error setting up Radio Select view")
	}
	view.Frame = false
	fmt.Fprint(view, rs.label)

	_, labelY := view.Size()
	btnY := rs.y + labelY + Offset
	btnX := rs.x + Offset

	rs.left, rs.top, rs.right, rs.bottom = view.Dimensions()
	for i, v := range rs.options {
		Assert(i < 10, "RadioSelect.Layout(): more than 9 options present")

		btn := NewButton(v.Name, v.Label, v.Value, btnX, btnY, 0, nil)
		rs.buttons = append(rs.buttons, *btn)

		btnView, err := btn.Layout(g)
		if err != nil {
			panic("error setting up button for radio select view")
		}

		ch := []rune(strconv.Itoa(i + 1))[0]

		err = g.SetKeybinding(rs.name, ch, gocui.ModNone, makeRadioSelection(btnView, rs, btn.value))
		if err != nil {
			panic("error setting up keybinding for radio select button")
		}

		btnX += len(btn.label) + BtnGap

		left, top, right, bottom := btnView.Dimensions()

		if rs.left > left {
			rs.left = left
		}
		if rs.top > top {
			rs.top = top
		}
		if rs.right < right {
			rs.right = right
		}
		if rs.bottom < bottom {
			rs.bottom = bottom
		}
	}

	return view, nil
}

func (rs *RadioSelect) CleanUp(g *gocui.Gui) {
	for _, v := range rs.buttons {
		g.DeleteView(v.name)
	}

	rs.deleteRadioBindings(g)
	g.DeleteView(rs.name)
}

func (rs *RadioSelect) Dimensions() (left, top, right, bottom int) {
	return rs.left, rs.top, rs.right, rs.bottom
}

func (rs *RadioSelect) Size() (x, y int) {
	return rs.right - rs.left, rs.bottom - rs.top
}

func (rs *RadioSelect) Selected() string {
	return rs.selected
}

func (rs *RadioSelect) Resize(x, y, w int) {
	if w == 0 {
		w = x + len(rs.label) + Offset
	}
	rs.x, rs.y, rs.w = x, y, w
}

func (rs *RadioSelect) HasLabel() bool {
	return true
}

func (rs *RadioSelect) deleteRadioBindings(g *gocui.Gui) {
	for i := range 10 {
		ch := []rune(strconv.Itoa(i + 1))[0]
		g.DeleteKeybinding(rs.name, ch, gocui.ModNone)
	}
}

func (rs *RadioSelect) Name() string {
	return rs.name
}

func makeRadioSelection(view *gocui.View, rs *RadioSelect, selected string) KeybindFunc {
	return func(g *gocui.Gui, v *gocui.View) error {
		if view == nil {
			return nil
		}

		for _, btn := range rs.buttons {
			btnView, err := g.View(btn.name)
			if err != nil {
				return nil
			}

			btnView.Highlight = false
		}

		view.Highlight = true
		rs.selected = selected

		return nil
	}
}
