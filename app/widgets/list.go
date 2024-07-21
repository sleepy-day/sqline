package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	. "github.com/sleepy-day/sqline/shared"
)

type List[T ListOpt] struct {
	name                     string
	label                    string
	searchName               string
	options                  []T
	current                  int
	left, top, right, bottom int
	selectFn                 KeybindFunc
}

type ListOpt interface {
	Text() string
	Subtext() string
}

func NewList[T ListOpt](name, label string, left, top, right, bottom int, opts []T, fn KeybindFunc) *List[T] {
	Assert(name != "", "NewList(): name is empty")
	Assert(left <= right, "NewList(): right value is larger than left")
	Assert(top <= bottom, "NewList(): top value is larger than bottom")

	return &List[T]{
		name:       name,
		searchName: name + "_search",
		label:      label,
		options:    opts,
		left:       left,
		top:        top,
		right:      right,
		bottom:     bottom,
		selectFn:   fn,
	}
}

func (list *List[ListOpt]) Layout(g *gocui.Gui) (*gocui.View, error) {
	view, err := g.SetView(list.searchName, list.left, list.top, list.right, list.top+4, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}

	view, err = g.SetView(list.name, list.left, list.top+5, list.right, list.bottom, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return nil, err
	}

	view.Clear()

	for i, v := range list.options {
		if i == list.current {
			fmt.Fprintf(view, "\x1b[3%d;%dm%s\x1b[0m\n", 2, 7, v.Text())
			continue
		}

		fmt.Fprintf(view, "%s\n", v.Text())
	}

	g.SetKeybinding(list.name, gocui.KeyEnter, gocui.ModNone, list.selectFn)

	view.Editable = true
	view.Editor = gocui.EditorFunc(list.listHandler(g))

	g.SetCurrentView(list.name)

	return view, nil
}

func (list *List[T]) CleanUp(g *gocui.Gui) {
	g.DeleteView(list.searchName)
	g.DeleteView(list.name)
}

func (list *List[T]) Append(opt T) {
	list.options = append(list.options, opt)
}

func (list *List[T]) listHandler(g *gocui.Gui) EditFunc {
	return func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
		switch {
		case key == gocui.KeyArrowUp:
			list.current--
			if list.current == -1 {
				list.current = len(list.options) - 1
			}

			list.Layout(g)
		case key == gocui.KeyArrowDown:
			list.current++
			if list.current == len(list.options) {
				list.current = 0
			}
			list.Layout(g)
		}
	}
}

func (list *List[T]) SetOptions(opts []T) {
	list.options = opts
}

func searchHandler(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	}
}
