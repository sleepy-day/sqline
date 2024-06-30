package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"

	. "github.com/sleepy-day/sqline/shared"
)

type Input struct {
	name      string
	label     string
	labelName string
	x, y, w   int
	editFn    EditFunc
}

func NewInput(name, label string, x, y, w int, editFn EditFunc, fn KeybindFunc) *Input {
	Assert(name != "", "NewInput(): name empty")
	Assert(label != "", "NewInput(): label empty")
	Assert(editFn != nil, "NewInput(): editFn is nil")

	return &Input{
		name:      name,
		labelName: name + "_label",
		label:     label,
		x:         x,
		y:         y,
		w:         w,
		editFn:    editFn,
	}
}

func (inp *Input) Layout(g *gocui.Gui) (*gocui.View, error) {
	view, err := g.SetView(inp.labelName, inp.x, inp.y, inp.x+inp.w, inp.y+2, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}
	view.Frame = false
	fmt.Fprint(view, inp.label)

	_, labelY := view.Size()
	x := inp.x + 1
	y := inp.y + labelY + 1

	view, err = g.SetView(inp.name, x, y, x+inp.w, y+2, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}

	view.Editable = true
	view.Editor = gocui.EditorFunc(inp.editFn)

	return view, nil
}

func (inp *Input) CleanUp(g *gocui.Gui) {
	g.DeleteView(inp.name)
}
