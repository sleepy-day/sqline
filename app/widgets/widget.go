package widgets

import (
	"github.com/awesome-gocui/gocui"
)

type KeybindFunc func(g *gocui.Gui, v *gocui.View) error

type EditFunc func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier)

type Widget interface {
	Layout(g *gocui.Gui) (*gocui.View, error)
	CleanUp(g *gocui.Gui)
}
