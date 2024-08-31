package components

import (
	"github.com/gdamore/tcell/v2"
)

type ListItem[T any] struct {
	Label []rune
	Value T
}

type List[T any] struct {
	left, top     int
	right, bottom int
	selected      int
	offset        int
	style         *tcell.Style
	hlStyle       tcell.Style
	listItems     []ListItem[T]
	window        *Window
}

func CreateList[T any](left, top, right, bottom int, listItems []ListItem[T], title []rune, style *tcell.Style) *List[T] {
	list := &List[T]{
		left:      left,
		top:       top,
		right:     right,
		bottom:    bottom,
		listItems: listItems,
		style:     style,
		hlStyle:   tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite),
		window:    CreateWindow(left, top, right, bottom, 0, 0, true, title, style),
	}

	if listItems == nil {
		list.selected = -1
	} else {
		list.selected = 0
	}

	return list
}

func (list *List[T]) SetList(items []ListItem[T]) {
	list.listItems = items
}

func (list *List[T]) SelectedItem() *ListItem[T] {
	if len(list.listItems) > 0 {
		return &list.listItems[list.selected]
	}

	return nil
}

func (list *List[T]) Resize(left, top, right, bottom int) {
	list.left = left
	list.top = top
	list.right = right
	list.bottom = bottom
	list.window.Resize(left, top, right, bottom)
}

func (list *List[T]) Add(item *ListItem[T]) {
	list.listItems = append(list.listItems, *item)
}

func (list *List[T]) HandleInput(ev tcell.Event) {
	switch event := ev.(type) {
	case *tcell.EventKey:
		switch {
		case event.Key() == tcell.KeyUp:
			if list.selected == 0 {
				break
			}
			list.selected--
			if list.selected < list.offset {
				list.offset--
			}
		case event.Key() == tcell.KeyDown:
			if list.selected == len(list.listItems)-1 {
				break
			}
			list.selected++

			if list.selected >= list.bottom-list.top-1 && list.offset < list.bottom-list.top-list.offset-1 {
				list.offset++
			}
		}
	}
}

func (list *List[T]) Render(screen tcell.Screen) {
	/*
		for i := range list.right - list.left + 1 {
			if i == 0 {
				screen.SetContent(list.left, list.top, tcell.RuneULCorner, nil, *list.style)
				screen.SetContent(list.left, list.bottom, tcell.RuneLLCorner, nil, *list.style)
				continue
			} else if i == list.right-list.left {
				screen.SetContent(list.left+i, list.top, tcell.RuneURCorner, nil, *list.style)
				screen.SetContent(list.left+i, list.bottom, tcell.RuneLRCorner, nil, *list.style)
				continue
			}

			screen.SetContent(list.left+i, list.bottom, tcell.RuneHLine, nil, *list.style)
			if i-1 < len(list.title) && i > 0 {
				screen.SetContent(list.left+i, list.top, list.title[i-1], nil, *list.style)
				continue
			}
			screen.SetContent(list.left+i, list.top, tcell.RuneHLine, nil, *list.style)
		}

		for i := range list.bottom - list.top {
			if i == 0 || i == list.bottom-list.top {
				continue
			}

			screen.SetContent(list.left, list.top+i, tcell.RuneVLine, nil, *list.style)
			screen.SetContent(list.right, list.top+i, tcell.RuneVLine, nil, *list.style)
		}
	*/

	list.window.Render(screen)

	style := list.style
	for row := range list.bottom - list.top - 1 {
		for col := range list.right - list.left - 1 {
			screen.SetContent(list.left+col+1, list.top+row+1, ' ', nil, *style)
		}
	}

	if len(list.listItems) == 0 {
		return
	}

	for i, j := list.offset, 0; i < len(list.listItems) && i < list.offset+list.bottom-list.top-1; i, j = i+1, j+1 {
		if list.selected == i {
			style = &list.hlStyle
		} else {
			style = list.style
		}

		for k, ch := range list.listItems[i].Label {
			screen.SetContent(list.left+k+1, list.top+j+1, ch, nil, *style)
		}
	}
}
