package components

import (
	"github.com/gdamore/tcell/v2"
)

type TreeItem struct {
	expanded bool
	Label    []rune
	Children []*TreeItem
	Value    string
	Child    bool
	Level    int
}

type Tree struct {
	left, top     int
	right, bottom int
	selected      int
	offset        int
	treeItems     []*TreeItem
	style         *tcell.Style
	hlStyle       tcell.Style
	visibleItems  []*TreeItem
}

func CreateTree(left, top, right, bottom int, treeItems []*TreeItem, style *tcell.Style) *Tree {
	tree := &Tree{
		left:      left,
		top:       top,
		right:     right,
		bottom:    bottom,
		treeItems: treeItems,
		style:     style,
		hlStyle:   tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite),
	}

	if treeItems == nil {
		tree.selected = -1
	} else {
		tree.selected = 0
	}

	return tree
}

func (tree *Tree) SelectedItem() *TreeItem {
	if len(tree.treeItems) > 0 {
		return tree.treeItems[tree.selected]
	}

	return nil
}

func (tree *Tree) Add(item *TreeItem) {
	tree.treeItems = append(tree.treeItems, item)
}

func (tree *Tree) HandleInput(ev tcell.Event) {
	switch event := ev.(type) {
	case *tcell.EventKey:
		switch {
		case event.Key() == tcell.KeyUp:
			if tree.selected == 0 {
				break
			}
			tree.selected--
			if tree.selected < tree.offset {
				tree.offset--
			}
		case event.Key() == tcell.KeyDown:
			if tree.selected == len(tree.visibleItems)-1 {
				break
			}
			tree.selected++

			if tree.selected >= tree.bottom-tree.top-1 && tree.offset < tree.bottom-tree.top-tree.offset-1 {
				tree.offset++
			}
		case event.Key() == tcell.KeyEnter:
			if len(tree.visibleItems[tree.selected].Children) > 0 {
				tree.visibleItems[tree.selected].expanded = !tree.visibleItems[tree.selected].expanded
			}
		}
	}
}

func (tree *Tree) Render(screen tcell.Screen) {
	tree.visibleItems = []*TreeItem{}
	for _, v := range tree.treeItems {
		tree.visibleItems = append(tree.visibleItems, v)

		if len(v.Children) <= 0 || !v.expanded {
			continue
		}

		for _, v := range v.Children {
			tree.visibleItems = append(tree.visibleItems, v)
			if len(v.Children) > 0 && v.expanded {
				tree.visibleItems = append(tree.visibleItems, v.Children...)
			}
		}

		if len(tree.visibleItems) >= tree.bottom-tree.top+tree.offset {
			break
		}
	}

	for i := range tree.right - tree.left {
		if i == 0 {
			screen.SetContent(tree.left, tree.top, tcell.RuneULCorner, nil, *tree.style)
			screen.SetContent(tree.left, tree.bottom, tcell.RuneLLCorner, nil, *tree.style)
			continue
		}

		screen.SetContent(tree.left+i, tree.top, tcell.RuneHLine, nil, *tree.style)
		screen.SetContent(tree.left+i, tree.bottom, tcell.RuneHLine, nil, *tree.style)
	}

	for i := range tree.bottom - tree.top {
		if i == 0 || i == tree.bottom-tree.top {
			continue
		}

		screen.SetContent(tree.left, tree.top+i, tcell.RuneVLine, nil, *tree.style)
		screen.SetContent(tree.right, tree.top+i, tcell.RuneVLine, nil, *tree.style)
	}

	if len(tree.treeItems) == 0 {
		return
	}

	offset := 0
	for i, j := tree.offset, 0; i < len(tree.visibleItems) && i < tree.offset+tree.bottom-tree.top-1; i, j = i+1, j+1 {
		offset = 0
		var ch rune

		if tree.visibleItems[i].expanded {
			ch = '-'
		} else if len(tree.visibleItems[i].Children) > 0 {
			ch = '+'
		} else {
			ch = ' '
		}

		if tree.visibleItems[i].Child {
			offset = tree.visibleItems[i].Level
		}

		if tree.selected == i {
			screen.SetContent(tree.left+1+offset, tree.top+j+1, ch, tree.visibleItems[i].Label, tree.hlStyle)
			continue
		}

		screen.SetContent(tree.left+1+offset, tree.top+j+1, ch, tree.visibleItems[i].Label, *tree.style)
	}
}

