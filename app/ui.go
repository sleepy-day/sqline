package app

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
)

var connViews []string = []string{
	"driver_label",
	"name_input",
	"connect_input",
	"test_button",
	"save_button",
}

func scale(factor float32, value int) int {
	return int(factor * float32(value))
}

func delAddView(g *gocui.Gui) {
	if s_mode != m_connect {
		return
	}

	g.DeleteView("add_database")
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	s := scale

	_, err := g.SetView("no_selection", -2, -2, -1, -1, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}

	_, err = g.SetView("databases", -1, -1, s(0.2, maxX), s(0.2, maxY), 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}

	_, err = g.SetView("schemas", -1, s(0.2, maxY), s(0.2, maxX), s(0.4, maxY), 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}

	_, err = g.SetView("tables", -1, s(0.4, maxY), s(0.2, maxX), maxY, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}

	_, err = g.SetView("cmdline", s(0.2, maxX), maxY-20, maxX, maxY, 1)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}

	var x, y int
	var editor gocui.Editor = gocui.EditorFunc(sqlEditor)
	if v, err := g.SetView("editor", s(0.20, maxX)+5, -1, maxX, maxY-20, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Editor = editor
		v.Editable = true
		v.Wrap = true

		x, y = g.Size()
	}

	lineNoView, err := g.SetView("line_numbers", s(0.2, maxX), -1, s(0.20, maxX)+5, maxY-20, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		fmt.Fprintln(lineNoView, fmt.Sprintf("%d\n%d", x, y))
	}

	return nil
}
