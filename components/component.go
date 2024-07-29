package components

import (
	"github.com/gdamore/tcell/v2"
)

type Component interface {
	Render(screen tcell.Screen)
	HandleInput(ev tcell.Event)
	Sync(maxX, maxY int)
	SetFocus()
	RemoveFocus()
}
