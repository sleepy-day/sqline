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
	EditorVisual
	EditorInsert
	DbList
	SchemaList
	TblList
	DataTable
	DataTableExpanded
	Indexes

	OpenConnStatus = "OpenConn"
	NewConnStatus  = "NewConn"
)

var (
	NoModeStatusStyle       tcell.Style = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	EditorStatusStyle       tcell.Style = tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite)
	EditorVisualStatusStyle tcell.Style = tcell.StyleDefault.Background(tcell.ColorFuchsia).Foreground(tcell.ColorWhite)
	EditorInsertStatusStyle tcell.Style = tcell.StyleDefault.Background(tcell.ColorLime).Foreground(tcell.ColorBlack)
	DatabaseStatusStyle     tcell.Style = tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack)
	SchemaStatusStyle       tcell.Style = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	TableStatusStyle        tcell.Style = tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite)
	DataTableStatusStyle    tcell.Style = tcell.StyleDefault.Background(tcell.ColorOrange).Foreground(tcell.ColorWhite)
	IndexesStatusStyle      tcell.Style = tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite)
	NewConnStatusStyle      tcell.Style = tcell.StyleDefault.Background(tcell.ColorMintCream).Foreground(tcell.ColorBlack)
	OpenConnStatusStyle     tcell.Style = tcell.StyleDefault.Background(tcell.ColorCoral).Foreground(tcell.ColorWhite)
)

type MainView struct {
	showDB, showSchema        bool
	left, top, right, bottom  int
	height, width             int
	sideListHeight, sideWidth int
	sideBoxHeight             int
	style, hlStyle            *tcell.Style
	editor                    *comp.Editor
	dbList                    *comp.List[db.DbInfo]
	schemaList                *comp.List[db.SchemaInfo]
	indexTree                 *comp.Tree
	tableTree                 *comp.Tree
	dataTable                 *comp.Table
	status                    *comp.StatusBar
	State                     MainViewState
}

func CreateMainView(left, top, right, bottom, pLeft, pTop, pRight, pBottom int, showDB, showSchema bool, buf []byte, style, hlStyle *tcell.Style) *MainView {
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
		sideWidth:      40,
		sideListHeight: Scale(0.2, bottom-1),
		sideBoxHeight:  Scale(0.2, bottom-1),
	}

	mainSideStart := view.sideWidth + 1
	viewBottom := view.bottom - 1

	if showDB && showSchema {
		view.dbList = comp.CreateList[db.DbInfo](view.left, view.top, view.sideWidth, view.sideBoxHeight, nil, []rune("Databases"), style)
		view.schemaList = comp.CreateList[db.SchemaInfo](view.left, view.sideBoxHeight+1, view.sideWidth, view.sideBoxHeight*2, nil, []rune("Schemas"), style)
	} else if showDB {
		view.dbList = comp.CreateList[db.DbInfo](view.left, view.top, view.sideWidth, view.sideBoxHeight, nil, []rune("Databases"), style)
	} else if showSchema {
		view.schemaList = comp.CreateList[db.SchemaInfo](view.left, view.top, view.sideWidth, view.sideBoxHeight, nil, []rune("Schemas"), style)
	}

	if showDB && showSchema {
		view.tableTree = comp.CreateTree(view.left, (view.sideBoxHeight*2)+1, view.sideWidth, viewBottom-view.sideBoxHeight, nil, []rune("Tables"), style)
	} else if showDB || showSchema {
		view.tableTree = comp.CreateTree(view.left, view.sideBoxHeight+1, view.sideWidth, viewBottom-view.sideBoxHeight, nil, []rune("Tables"), style)
	} else {
		view.tableTree = comp.CreateTree(view.left, view.top, view.sideWidth, viewBottom-view.sideBoxHeight, nil, []rune("Tables"), style)
	}

	view.indexTree = comp.CreateTree(view.left, viewBottom-view.sideBoxHeight+1, view.sideWidth, viewBottom, nil, []rune("Indexes"), style)

	tableHeight := 16
	view.editor = comp.CreateEditor(mainSideStart, view.top, view.right, viewBottom-tableHeight, nil, style, hlStyle)
	view.dataTable = comp.CreateTable(mainSideStart, view.bottom-tableHeight, view.right, viewBottom, pLeft, pTop, pRight, pBottom, 30, nil, style)
	view.status = comp.CreateStatusBar(view.left, view.bottom, view.right, 5, []rune("Normal"), style, &NoModeStatusStyle)

	return view
}

func (view *MainView) SetVisibleComponents(showDB, showSchema bool, screen tcell.Screen) {
	view.showDB = showDB
	view.showSchema = showSchema

	bottom := view.bottom - 1
	if view.showDB && view.showSchema {
		view.dbList.Resize(view.left, view.top, view.sideWidth, view.sideListHeight)
		view.schemaList.Resize(view.left, view.sideListHeight+1, view.sideListHeight, (view.sideListHeight*2)+1)
		view.tableTree.Resize(view.left, (view.sideListHeight*2)+1, view.sideWidth, bottom-view.sideListHeight)
		view.indexTree.Resize(view.left, bottom-view.sideListHeight+1, view.sideWidth, bottom)
	} else if view.showDB {
		view.dbList.Resize(view.left, view.top, view.sideWidth, view.sideListHeight)
		view.tableTree.Resize(view.left, view.sideListHeight+1, view.sideWidth, bottom-view.sideListHeight)
		view.indexTree.Resize(view.left, bottom-view.sideListHeight+1, view.sideWidth, bottom)
	} else if view.showSchema {
		view.schemaList.Resize(view.left, view.top, view.sideWidth, view.sideListHeight)
		view.tableTree.Resize(view.left, view.sideListHeight+1, view.sideWidth, bottom-view.sideListHeight)
		view.indexTree.Resize(view.left, bottom-view.sideListHeight+1, view.sideWidth, bottom)
	} else {
		view.tableTree.Resize(view.left, view.top, view.sideWidth, bottom-view.sideListHeight)
		view.indexTree.Resize(view.left, bottom-view.sideListHeight+1, view.sideWidth, bottom)
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

func (view *MainView) SetIndexTree(tables []db.Table) {
	if len(tables) == 0 {
		view.indexTree.SetItems([]*comp.TreeItem{})
	}

	var items []*comp.TreeItem
	for _, v := range tables {
		table := &comp.TreeItem{
			Label: []rune(v.Name),
			Value: v.Name,
			Level: 0,
		}

		for _, ix := range v.Indexes {
			index := &comp.TreeItem{
				Label: []rune(ix.Name),
				Value: ix.Name,
				Level: 1,
				Child: true,
			}

			for _, col := range ix.Cols {
				col := &comp.TreeItem{
					Label: []rune(fmt.Sprintf("%s - seq %d", col.ColumnName, col.SeqNo)),
					Value: col.ColumnName,
					Level: 2,
					Child: true,
				}

				index.Children = append(index.Children, col)
			}

			table.Children = append(table.Children, index)
		}

		items = append(items, table)
	}

	view.indexTree.SetItems(items)
}

func (view *MainView) SetTableTree(tables []db.Table) {
	if len(tables) == 0 {
		view.tableTree.SetItems([]*comp.TreeItem{})
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

	view.tableTree.SetItems(items)
}

func (view *MainView) Render(screen tcell.Screen) {
	view.editor.Render(screen)
	view.tableTree.Render(screen)
	view.indexTree.Render(screen)

	if view.showDB {
		view.dbList.Render(screen)
	}
	if view.showSchema {
		view.schemaList.Render(screen)
	}

	view.dataTable.Render(screen)
	view.status.Render(screen)
}

func (view *MainView) EditorInNormalMode() bool {
	return view.editor.InNormalMode()
}

func (view *MainView) HandleInput(ev *tcell.EventKey) {
	switch view.State {
	case NoFocus:
		break
	case Editor:
		if ev.Rune() == 'i' {
			view.SetState(EditorInsert)
		}
		if ev.Rune() == 'v' || ev.Rune() == 'V' {
			view.SetState(EditorVisual)
		}
		fallthrough
	case EditorVisual, EditorInsert:
		if ev.Key() == tcell.KeyEsc {
			view.SetState(Editor)
		}
		view.editor.HandleInput(ev)
	case DbList:
		view.dbList.HandleInput(ev)
	case SchemaList:
		view.schemaList.HandleInput(ev)
	case TblList:
		view.tableTree.HandleInput(ev)
	case DataTableExpanded:
		if ev.Key() == tcell.KeyEsc {
			view.State = DataTable
		}
		fallthrough
	case DataTable:
		if ev.Key() == tcell.KeyEnter {
			view.State = DataTableExpanded
		}
		view.dataTable.HandleInput(ev)
	case Indexes:
		view.indexTree.HandleInput(ev)
	}
}

func (view *MainView) SetState(state MainViewState) {
	view.State = state

	switch state {
	case Editor:
		view.status.SetStatus([]rune("Editor: Normal"), EditorStatusStyle)
	case EditorVisual:
		view.status.SetStatus([]rune("Editor: Visual"), EditorVisualStatusStyle)
	case EditorInsert:
		view.status.SetStatus([]rune("Editor: Insert"), EditorInsertStatusStyle)
	case DbList:
		view.status.SetStatus([]rune("Databases"), DatabaseStatusStyle)
	case SchemaList:
		view.status.SetStatus([]rune("Schemas"), SchemaStatusStyle)
	case TblList:
		view.status.SetStatus([]rune("Tables"), TableStatusStyle)
	case DataTable:
		view.status.SetStatus([]rune("DataTable"), DataTableStatusStyle)
	case Indexes:
		view.status.SetStatus([]rune("Indexes"), IndexesStatusStyle)
	}
}

func (view *MainView) SetStatus(status string) {
	switch status {
	case OpenConnStatus:
		view.status.SetStatus([]rune(OpenConnStatus), OpenConnStatusStyle)
	case NewConnStatus:
		view.status.SetStatus([]rune(NewConnStatus), NewConnStatusStyle)
	default:
		view.status.SetStatus([]rune(status), NoModeStatusStyle)
	}
}

func (view *MainView) SetError(msg []rune) {
	view.status.SetErr(msg)
}

func (view *MainView) SetInfo(msg []rune) {
	view.status.SetInfo(msg)
}

func (view *MainView) TableFunc() comp.TableDataFunc {
	return view.dataTable.TableFunc()
}

func (view *MainView) SetSQLFunc(fn comp.ExecSQLFunc) {
	view.editor.SetSQLFunc(fn)
}
