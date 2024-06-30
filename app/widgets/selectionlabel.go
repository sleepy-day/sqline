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
	name     string
	label    string
	x, y, w  int
	selected string
	options  []Option
	buttons  []Button
}

func NewRadioSelect(name, label string, x, y int, options map[string]string) *RadioSelect {
	Assert(name != "", "NewRadioSelect(): name empty")
	Assert(label != "", "NewRadioSelect(): name empty")
	Assert(len(options) > 0, "NewRadioSelect(): options length is 0")
	Assert(len(options) <= 9, "NewRadioSelect(): options length greater than 9")

	opts := []Option{}
	for k, v := range options {
		opts = append(opts, Option{Label: k, Value: v})
	}

	return &RadioSelect{
		name:    name,
		label:   label,
		x:       x,
		y:       y,
		w:       len(label) + 1,
		options: opts,
	}
}

func (rs *RadioSelect) Layout(g *gocui.Gui) (*gocui.View, error) {
	view, err := g.SetView(rs.label, rs.x, rs.y, rs.x+rs.w, rs.y+2, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}
	view.Frame = false
	fmt.Fprint(view, rs.label)

	_, labelY := view.Size()
	y := rs.y + labelY + 1
	x := rs.x + 1

	for i, v := range rs.options {
		Assert(i < 10, "RadioSelect.Layout(): more than 9 options present")

		btn := NewButton(v.Name, v.Label, v.Value, x+1, y, 0, nil)
		btnView, err := btn.Layout(g)
		if err != nil {
			return nil, err
		}

		ch := []rune(strconv.Itoa(i))[0]

		err = g.SetKeybinding(rs.name, ch, gocui.ModNone, makeRadioSelection(btnView, rs, btn.value))
		if err != nil {
			return nil, err
		}

		x += len(btn.label) + 1
	}

	return view, nil
}

func (rs *RadioSelect) CleanUp(g *gocui.Gui) {
	for _, v := range rs.buttons {
		g.DeleteView(v.name)
	}

	g.DeleteView(rs.name)
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
