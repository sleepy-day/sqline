package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type Info struct {
	name                     string
	left, top, right, bottom int
}

func NewInfo(name string, left, top, width int) *Info {
	return &Info{
		name:   name,
		left:   left,
		top:    top,
		right:  left + width,
		bottom: top + MinHeight,
	}
}

func (info *Info) Layout(g *gocui.Gui) (*gocui.View, error) {
	view, err := g.SetView(info.name, info.left, info.top, info.right, info.bottom, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}

	return view, nil
}

func (info *Info) CleanUp(g *gocui.Gui) {
	g.DeleteView(info.name)
}

func (info *Info) Write(g *gocui.Gui, value string) {
	view, err := g.View(info.name)
	if err != nil {
		return
	}

	view.Clear()
	fmt.Fprint(view, value)
}

func (info *Info) Size() (x, y int) {
	x = info.right - info.left
	y = info.bottom - info.top

	return
}

func (info *Info) Dimensions() (left, top, right, bottom int) {
	left = info.left
	top = info.top
	right = info.right
	bottom = info.bottom

	return
}

func (info *Info) Name() string {
	return info.name
}
