package views

import (
	"github.com/gdamore/tcell/v2"
	comp "github.com/sleepy-day/sqline/components"
	"github.com/sleepy-day/sqline/util"
)

type SelectFunc func(util.DBEntry)

type OpenConnView struct {
	open                     bool
	left, top, right, bottom int
	height, width            int
	style, hlStyle           *tcell.Style
	window                   *comp.Window
	connList                 *comp.List[util.DBEntry]
	infoBtn                  *comp.Button
	selectFunc               func(util.DBEntry)
}

func CreateOpenConnView(left, top, right, bottom int, style, hlStyle *tcell.Style, dbEntries []util.DBEntry, selectFunc SelectFunc) *OpenConnView {
	ocView := &OpenConnView{
		left:       left,
		top:        top,
		right:      right,
		bottom:     bottom,
		height:     top - bottom,
		width:      right - left,
		style:      style,
		hlStyle:    hlStyle,
		selectFunc: selectFunc,
	}

	ocView.window = comp.CreateWindow(left, top, right, bottom, 2, 2, true, []rune("Open A Saved Connection"), style)

	var conns []comp.ListItem[util.DBEntry]
	for _, v := range dbEntries {
		conns = append(conns, comp.ListItem[util.DBEntry]{
			Label: []rune(v.Name),
			Value: v,
		})
	}

	innLeft, innTop, innRight, innBottom := ocView.window.GetUsableDimensions()
	ocView.connList = comp.CreateList(innLeft, innTop-5, innRight, innBottom, conns, nil, style)

	return ocView
}

func (ocv *OpenConnView) Render(screen tcell.Screen) {
	ocv.window.Render(screen)
	ocv.connList.Render(screen)
}

func (ocv *OpenConnView) HandleInput(ev *tcell.EventKey) {
	switch {
	case ev.Key() == tcell.KeyEnter:
		conn := ocv.connList.SelectedItem().Value
		ocv.selectFunc(conn)
	default:
		ocv.connList.HandleInput(ev)
	}
}
