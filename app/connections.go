package app

import (
	"github.com/awesome-gocui/gocui"
	"github.com/sleepy-day/sqline/app/widgets"
	. "github.com/sleepy-day/sqline/shared"
)

type ConnListPage struct {
	Height, Width int
	List          *widgets.List[ConnInfo]
}

func CreateConnListPage(g *gocui.Gui, width, height int, opts []ConnInfo) *ConnListPage {
	page := &ConnListPage{
		Height: height,
		Width:  width,
	}

	guiX, guiY := g.Size()

	left := Scale(0.5, guiX-width)
	right := left + width
	top := Scale(0.5, guiY-height)
	bottom := top + height

	page.List = widgets.NewList("conn_list", "", left, top, right, bottom, opts, nil)

	return page
}

func (page *ConnListPage) SetOptions(opts []ConnInfo) {
	page.List.SetOptions(opts)
}

func (page *ConnListPage) Open(g *gocui.Gui) error {
	_, err := page.List.Layout(g)
	if err != nil {
		return err
	}

	return nil
}

func (page *ConnListPage) Close(g *gocui.Gui) {
	page.List.CleanUp(g)
}
