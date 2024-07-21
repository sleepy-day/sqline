package widgets

import (
	"errors"

	"github.com/awesome-gocui/gocui"
	. "github.com/sleepy-day/sqline/shared"
)

type Modal struct {
	name                     string
	label                    string
	left, top, right, bottom int
}

func NewModal(name, label string, left, top, right, bottom int) *Modal {
	Assert(left < right, "NewModal(): left has a higher value than right")
	Assert(top < bottom, "NewModal(): top has a higher value than bottom")
	Assert(name != "", "NewModal(): name is empty")

	return &Modal{
		name:   name,
		label:  label,
		left:   left,
		top:    top,
		right:  right,
		bottom: bottom,
	}
}

func (mdl *Modal) Layout(g *gocui.Gui) (*gocui.View, error) {
	view, err := g.SetView(mdl.name, mdl.left, mdl.top, mdl.right, mdl.bottom, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}

	return view, nil
}

func (mdl *Modal) Resize(left, top, right, bottom int) {
	mdl.left = left
	mdl.top = top
	mdl.right = right
	mdl.bottom = bottom
}

func (mdl *Modal) CleanUp(g *gocui.Gui) {
	g.DeleteView(mdl.name)
}

func (mdl *Modal) Dimensions() (left, top, right, bottom int) {
	left = mdl.left
	top = mdl.top
	right = mdl.right
	bottom = mdl.bottom

	return
}

func (mdl *Modal) Size() (x, y int) {
	x = mdl.right - mdl.left
	y = mdl.bottom - mdl.top

	return
}

func (mdl *Modal) UsableSize() (left, top, right, bottom int) {
	x, y := mdl.Size()

	left = left + Scale(0.1, x)
	right = right - Scale(0.1, x)
	top = top + Scale(0.05, y)
	bottom = bottom + Scale(0.05, y)

	return
}
