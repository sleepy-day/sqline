package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"

	. "github.com/sleepy-day/sqline/shared"
)

type Input struct {
	name                     string
	label                    string
	labelName                string
	x, y, w, labelW          int
	left, top, right, bottom int
	editFn                   EditFunc
	frame, labelFrame        bool
}

func NewInput(name, label string, x, y, w int, editFn EditFunc, fn KeybindFunc) *Input {
	Assert(name != "", "NewInput(): name empty")
	Assert(label != "", "NewInput(): label empty")

	if editFn == nil {
		editFn = defaultInputHandler
	}

	return &Input{
		name:       name,
		labelName:  name + "_label",
		label:      label,
		labelFrame: false,
		frame:      true,
		x:          x,
		y:          y,
		w:          x + w,
		labelW:     x + len(label) + Offset,
		editFn:     editFn,
	}
}

func (inp *Input) Layout(g *gocui.Gui) (*gocui.View, error) {
	view, err := g.SetView(inp.labelName, inp.x, inp.y, inp.labelW, inp.y+MinHeight, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}
	view.Frame = inp.labelFrame
	view.Clear()
	fmt.Fprint(view, inp.label)

	inp.left, inp.top, inp.right, inp.bottom = view.Dimensions()

	if inp.labelFrame {
		inp.top--
	}

	_, labelY := view.Size()
	x := inp.x + Offset
	y := inp.y + labelY + Offset

	view, err = g.SetView(inp.name, x, y, inp.w, y+MinHeight, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}

	view.Frame = inp.frame
	view.Editable = true
	view.Editor = gocui.EditorFunc(inp.editFn)

	left, top, right, bottom := view.Dimensions()

	if inp.frame {
		bottom++
	}

	if inp.left > left {
		inp.left = left
	}
	if inp.top > top {
		inp.top = top
	}
	if inp.right < right {
		inp.right = right
	}
	if inp.bottom < bottom {
		inp.bottom = bottom
	}

	return view, nil
}

func (inp *Input) Resize(x, y, w int) {
	inp.x, inp.y = x, y
	inp.w, inp.labelW = x+w, x+len(inp.label)+1
}

func (inp *Input) CleanUp(g *gocui.Gui) {
	g.DeleteView(inp.labelName)
	g.DeleteView(inp.name)
}

func (inp *Input) Dimensions() (left, top, right, bottom int) {
	return inp.left, inp.top, inp.right, inp.bottom
}

func (inp *Input) Size() (x, y int) {
	return inp.right - inp.left, inp.bottom - inp.top
}

func (inp *Input) HasLabel() bool {
	return true
}

func defaultInputHandler(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	}
}

func (inp *Input) Name() string {
	return inp.name
}
