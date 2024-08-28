package views

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	comp "github.com/sleepy-day/sqline/components"
	"github.com/sleepy-day/sqline/db"
	. "github.com/sleepy-day/sqline/shared"
)

type MainViewState byte

const (
	NoFocus MainViewState = iota
	Editor
	DbList
	SchemaList
	TblList
	DataTable
)

type MainView struct {
	showDB, showSchema        bool
	left, top, right, bottom  int
	height, width             int
	sideListHeight, sideWidth int
	style, hlStyle            *tcell.Style
	editor                    *comp.Editor
	dbList                    *comp.List[db.DbInfo]
	schemaList                *comp.List[db.SchemaInfo]
	tblList                   *comp.Tree
	dataTable                 *comp.Table
	status                    *comp.StatusBar
	State                     MainViewState
}

func CreateMainView(left, top, right, bottom int, showDB, showSchema bool, buf []byte, style, hlStyle *tcell.Style) *MainView {
	view := &MainView{
		left:           left,
		top:            top,
		right:          right - 1,
		bottom:         bottom - 1,
		height:         bottom - top - 1,
		width:          right - left - 1,
		style:          style,
		hlStyle:        hlStyle,
		showDB:         showDB,
		showSchema:     showSchema,
		State:          NoFocus,
		sideWidth:      30,
		sideListHeight: Scale(0.2, bottom-1),
	}

	sideBarWidth := 30
	mainSideStart := sideBarWidth + 1
	dbSchemaHeight := Scale(0.2, view.bottom)
	viewBottom := view.bottom - 1

	if showDB && showSchema {
		view.dbList = comp.CreateList[db.DbInfo](view.left, view.top, sideBarWidth, dbSchemaHeight, nil, []rune("Databases"), style)
		view.schemaList = comp.CreateList[db.SchemaInfo](view.left, dbSchemaHeight+1, sideBarWidth, dbSchemaHeight*2, nil, []rune("Schemas"), style)
	} else if showDB {
		view.dbList = comp.CreateList[db.DbInfo](view.left, view.top, sideBarWidth, dbSchemaHeight, nil, []rune("Databases"), style)
	} else if showSchema {
		view.schemaList = comp.CreateList[db.SchemaInfo](view.left, view.top, sideBarWidth, dbSchemaHeight, nil, []rune("Schemas"), style)
	}

	if showDB && showSchema {
		view.tblList = comp.CreateTree(view.left, (dbSchemaHeight*2)+1, sideBarWidth, viewBottom, nil, []rune("Tables"), style)
	} else if showDB || showSchema {
		view.tblList = comp.CreateTree(view.left, dbSchemaHeight+1, sideBarWidth, viewBottom, nil, []rune("Tables"), style)
	} else {
		view.tblList = comp.CreateTree(view.left, view.top, sideBarWidth, viewBottom, nil, []rune("Tables"), style)
	}

	tableHeight := 16
	view.editor = comp.CreateEditor(mainSideStart, view.top, view.right, viewBottom-tableHeight, nil, style, hlStyle)
	view.dataTable = comp.CreateTable(mainSideStart, view.bottom-tableHeight, view.right, viewBottom, 30, nil, style)
	view.status = comp.CreateStatusBar(view.left, view.bottom, view.right, 5, []rune("NoMode"), style, hlStyle)

	return view
}

func (view *MainView) SetVisibleComponents(showDB, showSchema bool, screen tcell.Screen) {
	view.showDB = showDB
	view.showSchema = showSchema

	bottom := view.bottom - 1
	if view.showDB && view.showSchema {
		view.dbList.Resize(view.left, view.top, view.sideWidth, view.sideListHeight)
		view.schemaList.Resize(view.left, view.sideListHeight+1, view.sideListHeight, (view.sideListHeight*2)+1)
		view.tblList.Resize(view.left, (view.sideListHeight*2)+1, view.sideWidth, bottom)
	} else if view.showDB {
		view.dbList.Resize(view.left, view.top, view.sideWidth, view.sideListHeight)
		view.tblList.Resize(view.left, view.sideListHeight+1, view.sideWidth, bottom)
	} else if view.showSchema {
		view.schemaList.Resize(view.left, view.top, view.sideWidth, view.sideListHeight)
		view.tblList.Resize(view.left, view.sideListHeight+1, view.sideWidth, bottom)
	} else {
		view.tblList.Resize(view.left, view.top, view.sideWidth, bottom)
	}

	screen.Fill(' ', *view.style)
	view.Render(screen)
	screen.Sync()
	screen.Show()
}

func (view *MainView) SetDatabaseList(dbInfo []db.DbInfo) {
	if len(dbInfo) == 0 {
		view.dbList.SetList([]comp.ListItem[db.DbInfo]{})
	}

	items := []comp.ListItem[db.DbInfo]{}
	for _, v := range dbInfo {
		items = append(items, comp.ListItem[db.DbInfo]{
			Label: []rune(v.Name),
			Value: v,
		})
	}

	view.dbList.SetList(items)
}

func (view *MainView) SetSchemaList(schemaInfo []db.SchemaInfo) {
	if len(schemaInfo) == 0 {
		view.schemaList.SetList([]comp.ListItem[db.SchemaInfo]{})
	}

	items := []comp.ListItem[db.SchemaInfo]{}
	for _, v := range schemaInfo {
		items = append(items, comp.ListItem[db.SchemaInfo]{
			Label: []rune(v.Name),
			Value: v,
		})
	}

	view.schemaList.SetList(items)
}

func (view *MainView) SetTableList(tables []db.Table) {
	if len(tables) == 0 {
		view.tblList.SetItems([]*comp.TreeItem{})
	}

	var items []*comp.TreeItem
	for _, v := range tables {
		table := &comp.TreeItem{
			Label: []rune(v.Name),
			Value: v.Name,
			Level: 0,
		}

		for _, col := range v.Columns {
			table.Children = append(table.Children, &comp.TreeItem{
				Label: []rune(fmt.Sprintf("%s - %s", col.Name, col.Type)),
				Child: true,
				Level: 1,
				Value: col.Name,
			})
		}

		items = append(items, table)
	}

	view.tblList.SetItems(items)
}

func (view *MainView) Render(screen tcell.Screen) {
	view.editor.Render(screen)
	view.tblList.Render(screen)
	view.dataTable.Render(screen)

	if view.showDB {
		view.dbList.Render(screen)
	}
	if view.showSchema {
		view.schemaList.Render(screen)
	}

	view.status.Render(screen)
}

func (view *MainView) EditorInNormalMode() bool {
	return view.editor.InNormalMode()
}

func (view *MainView) HandleInput(ev tcell.Event) {
	switch view.State {
	case NoFocus:
		break
	case Editor:
		view.editor.HandleInput(ev)
	case DbList:
		view.dbList.HandleInput(ev)
	case SchemaList:
		view.schemaList.HandleInput(ev)
	case TblList:
		view.tblList.HandleInput(ev)
	case DataTable:
		view.dataTable.HandleInput(ev)
	}
}

func (view *MainView) SetState(state MainViewState) {
	view.State = state

	switch state {
	case Editor:
		view.status.SetStatus([]rune("Editor: Normal"))
	case DbList:
		view.status.SetStatus([]rune("Databases"))
	case SchemaList:
		view.status.SetStatus([]rune("Schemas"))
	case TblList:
		view.status.SetStatus([]rune("Tables"))
	case DataTable:
		view.status.SetStatus([]rune("DataTable"))
	}
}

func (view *MainView) SetStatus(status []rune) {
	view.status.SetStatus(status)
}
