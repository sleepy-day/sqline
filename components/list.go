package components

import (
	"github.com/gdamore/tcell/v2"
)

type ListItem struct {
	Label []rune
	Value string
}

type List struct {
	left, top     int
	right, bottom int
	listItems     []ListItem
	selected      int
	style         *tcell.Style
	hlStyle       tcell.Style
	offset        int
}

func CreateList(left, top, right, bottom int, listItems []ListItem, style *tcell.Style) *List {
	list := &List{
		left:      left,
		top:       top,
		right:     right,
		bottom:    bottom,
		listItems: listItems,
		style:     style,
		hlStyle:   tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite),
	}

	if listItems == nil {
		list.selected = -1
	} else {
		list.selected = 0
	}

	return list
}

func (list *List) SelectedItem() *ListItem {
	if len(list.listItems) > 0 {
		return &list.listItems[list.selected]
	}

	return nil
}

func (list *List) Add(item *ListItem) {
	list.listItems = append(list.listItems, *item)
}

func (list *List) HandleInput(ev tcell.Event) {
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

func (list *List) Render(screen tcell.Screen) {
	for i := range list.right - list.left {
		if i == 0 {
			screen.SetContent(list.left, list.top, tcell.RuneULCorner, nil, *list.style)
			screen.SetContent(list.left, list.bottom, tcell.RuneLLCorner, nil, *list.style)
			continue
		}

		screen.SetContent(list.left+i, list.top, tcell.RuneHLine, nil, *list.style)
		screen.SetContent(list.left+i, list.bottom, tcell.RuneHLine, nil, *list.style)
	}

	for i := range list.bottom - list.top {
		if i == 0 || i == list.bottom-list.top {
			continue
		}

		screen.SetContent(list.left, list.top+i, tcell.RuneVLine, nil, *list.style)
		screen.SetContent(list.right, list.top+i, tcell.RuneVLine, nil, *list.style)
	}

	if len(list.listItems) == 0 {
		return
	}

	for i, j := list.offset, 0; i < len(list.listItems) && i < list.offset+list.bottom-list.top-1; i, j = i+1, j+1 {
		if len(list.listItems[i].Label) > 1 {
			if list.selected == i {
				screen.SetContent(list.left+1, list.top+j+1, list.listItems[i].Label[0], list.listItems[i].Label[1:], list.hlStyle)
				continue
			}

			screen.SetContent(list.left+1, list.top+j+1, list.listItems[i].Label[0], list.listItems[i].Label[1:], *list.style)
			continue
		}

		if list.selected == i {
			screen.SetContent(list.left+1, list.top+j+1, list.listItems[i].Label[0], nil, list.hlStyle)
			continue
		}

		screen.SetContent(list.left+1, list.top+j+1, list.listItems[i].Label[0], nil, *list.style)
	}
}
