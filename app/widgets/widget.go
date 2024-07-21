package widgets

import (
	"github.com/awesome-gocui/gocui"
)

const (
	MinHeight = 2
	Offset    = 1
	BtnGap    = 3
)

type KeybindFunc func(g *gocui.Gui, v *gocui.View) error

type EditFunc func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier)

type Widget interface {
	Layout(g *gocui.Gui) (*gocui.View, error)
	CleanUp(g *gocui.Gui)
	Dimensions() (left, top, right, bottom int)
	Size() (x, y int)
	Resize(x, y, w int)
	HasLabel() bool
}
