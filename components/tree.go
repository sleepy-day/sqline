package components

import (
	"github.com/gdamore/tcell/v2"
)

var (
	FirstSpacing  = []rune(" ")
	SecondSpacing = []rune("  ")
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
	width, height int
	label         []rune
	selected      int
	offset        int
	treeItems     []*TreeItem
	style         *tcell.Style
	hlStyle       tcell.Style
	visibleItems  []*TreeItem
}

func CreateTree(left, top, right, bottom int, treeItems []*TreeItem, label []rune, style *tcell.Style) *Tree {
	tree := &Tree{
		left:      left,
		top:       top,
		right:     right,
		bottom:    bottom,
		width:     right - left,
		height:    bottom - top,
		label:     label,
		treeItems: treeItems,
		style:     style,
		hlStyle:   tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite),
		selected:  -1,
	}

	if treeItems == nil {
		tree.selected = -1
	} else {
		tree.selected = 0
	}

	return tree
}

func (tree *Tree) Resize(left, top, right, bottom int) {
	tree.left = left
	tree.top = top
	tree.right = right
	tree.bottom = bottom
	tree.width = right - left
	tree.height = bottom - top
}

func (tree *Tree) SetItems(items []*TreeItem) {
	tree.treeItems = items
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

func (tree *Tree) HandleInput(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyUp:
		if tree.selected == 0 {
			break
		}
		tree.selected--
		if tree.selected < tree.offset {
			tree.offset--
		}
	case tcell.KeyDown:
		if tree.selected == len(tree.visibleItems)-1 {
			break
		}
		tree.selected++

		if tree.selected >= tree.bottom-tree.top-1 && tree.offset < tree.bottom-tree.top-tree.offset-1 {
			tree.offset++
		}
	case tcell.KeyEnter:
		if len(tree.visibleItems[tree.selected].Children) > 0 {
			tree.visibleItems[tree.selected].expanded = !tree.visibleItems[tree.selected].expanded
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
		for j := range tree.bottom - tree.top {
			screen.SetContent(tree.left+i, tree.top+j, ' ', nil, *tree.style)
		}
	}

	for i := range tree.right - tree.left + 1 {
		if i == 0 {
			screen.SetContent(tree.left, tree.top, tcell.RuneULCorner, nil, *tree.style)
			screen.SetContent(tree.left, tree.bottom, tcell.RuneLLCorner, nil, *tree.style)
			continue
		} else if i == tree.right-tree.left {
			screen.SetContent(tree.left+i, tree.top, tcell.RuneURCorner, nil, *tree.style)
			screen.SetContent(tree.left+i, tree.bottom, tcell.RuneLRCorner, nil, *tree.style)
			continue
		}

		topCh := tcell.RuneHLine
		if i < len(tree.label)+1 && i > 0 {
			topCh = tree.label[i-1]
		}

		screen.SetContent(tree.left+i, tree.top, topCh, nil, *tree.style)
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

		var label []rune
		if tree.visibleItems[i].Level == 1 {
			label = append(label, FirstSpacing...)
		} else if tree.visibleItems[i].Level == 2 {
			label = append(label, SecondSpacing...)
		}

		label = append(label, ch)
		label = append(label, tree.visibleItems[i].Label...)

		style := tree.style
		if tree.selected == i {
			style = &tree.hlStyle
		}

		for k, ch := range label {
			if k == tree.width-1 {
				break
			}
			screen.SetContent(tree.left+offset+k+1, tree.top+j+1, ch, nil, *style)
		}
	}
}
