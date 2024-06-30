package app

import "github.com/awesome-gocui/gocui"

func editorInsertMode(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyEnter:
		v.EditNewLine()
	case key == gocui.KeyArrowDown:
		v.MoveCursor(0, 1)
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0)
	}
}

func editorNormalMode(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch == 'j' || key == gocui.KeyArrowUp:
		v.MoveCursor(0, 1)
	case ch == 'k' || key == gocui.KeyArrowDown:
		v.MoveCursor(0, -1)
	case ch == 'h' || key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0)
	case ch == 'l' || key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0)
	}
}

func sqlEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch s_mode {
	case m_normal:
		editorNormalMode(v, key, ch, mod)
	case m_insert:
		editorInsertMode(v, key, ch, mod)
	}
}
