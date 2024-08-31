package app

import (
	"errors"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sleepy-day/sqline/components"
	"github.com/sleepy-day/sqline/db"
	"github.com/sleepy-day/sqline/util"
	"github.com/sleepy-day/sqline/views"
)

type MainAppState byte

const (
	NoViewFocused MainAppState = iota
	NormalMode
	MainView
	NewConnView
	OpenConnView
	Editor

	NormalInfo    = "e - Editor | d - DataTable | D - Databases | s - Schemas | t - Tables | i - Indexes | A - Add | C - Connect | Q - Quit"
	EditorInfo    = "i - Insert Mode | v - Visual Mode | V - Visual Mode (Whole Line) | Esc - Normal Mode/Exit Editor Mode"
	DataTableInfo = "Arrow Keys - Select Row/Col | Enter - Expand Cell | Esc - Normal Mode/Exit Expanded Cell"
	ListInfo      = "Up/Down - Select Item | Esc - Normal Mode"
	TreeInfo      = "Up/Down - Select Item | Enter - Expand/Collapse Selection | Esc - NormalMode"
	OpenConnInfo  = "Up/Down - Select Connection | Enter - Connect | Esc - Cancel"
	NewConnInfo   = "Tab - Change Selection | 1-4 - Change Driver Selection on Radio | Enter - Select Buttons (If highlighted) | Esc - Cancel"
)

var (
	defStyle   tcell.Style = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	hlStyle    tcell.Style = tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite)
	maxX, maxY int

	editor *components.Editor
	screen tcell.Screen

	// Pop up window coords
	pLeft, pTop, pRight, pBottom int
	pWidth, pHeight              int = 85, 30
)

type Sqline struct {
	state                        MainAppState
	database                     db.Database
	screen                       tcell.Screen
	config                       *util.SqlineConf
	mainView                     *views.MainView
	newConnView                  *views.NewConnView
	openConnView                 *views.OpenConnView
	maxX, maxY                   int
	pLeft, pTop, pRight, pBottom int
	pWidth, pHeight              int
}

func createSqline(maxX, maxY int, screen tcell.Screen) *Sqline {
	var buf []byte
	if len(os.Args) > 1 {
		filePath := os.Args[1]
		f, err := os.ReadFile(filePath)
		if err == nil {
			buf = f
		}
	}

	sqline := Sqline{
		state:   NormalMode,
		maxX:    maxX,
		maxY:    maxY,
		pWidth:  85,
		pHeight: 30,
		screen:  screen,
	}

	sqline.CalcPopupSize()

	conf, err := util.LoadConf()
	if err != nil {
		defer sqline.handleError(err)
	}

	sqline.config = conf
	sqline.mainView = views.CreateMainView(0, 0, maxX, maxY, sqline.pLeft, sqline.pTop, sqline.pRight, sqline.pBottom, true, true, buf, &defStyle, &hlStyle)
	sqline.newConnView = views.CreateNewConnView(sqline.pLeft, sqline.pTop, sqline.pRight, sqline.pBottom, &defStyle, &hlStyle, sqline.createTestFunc(), sqline.createSaveFunc())
	sqline.openConnView = views.CreateOpenConnView(sqline.pLeft, sqline.pTop, sqline.pRight, sqline.pBottom, &defStyle, &hlStyle, sqline.config.SavedConns, sqline.createSelectFunc())

	sqline.setInfo()
	return &sqline
}

func (sqline *Sqline) handleError(err error) {
	sqline.mainView.SetError([]rune(err.Error()))
}

func (sqline *Sqline) createTestFunc() views.TestFunc {
	return func(connStr, driver string) error {
		db, err := sqlx.Connect(driver, connStr)
		if err != nil {
			sqline.handleError(err)
			return err
		}
		defer db.Close()

		return nil
	}
}

func (sqline *Sqline) createSelectFunc() views.SelectFunc {
	return func(dbEntry util.DBEntry) {
		sqline.setDB(dbEntry)
		screen.Fill(' ', defStyle)
		sqline.state = NormalMode
		sqline.mainView.SetStatus("Normal")
	}
}

func (sqline *Sqline) createSaveFunc() views.SaveFunc {
	return func(name, connStr, driver string) {
		sqline.config.SavedConns = append(sqline.config.SavedConns, util.DBEntry{
			Name:    name,
			ConnStr: connStr,
			Driver:  driver,
		})

		err := util.SaveConf(sqline.config)
		if err != nil {
			sqline.handleError(err)
		}

		conf, err := util.LoadConf()
		if err != nil {
			sqline.handleError(err)
		} else {
			sqline.config.SavedConns = conf.SavedConns
			sqline.openConnView.SetConns(conf.SavedConns)
		}

		sqline.state = NormalMode
		sqline.mainView.SetStatus("Normal")
		screen.Fill(' ', defStyle)
		sqline.mainView.SetInfo([]rune("Connection Saved!"))
	}
}

func (sqline *Sqline) setInfo() {
	switch {
	case sqline.state == NormalMode:
		sqline.mainView.SetInfo([]rune(NormalInfo))
	case sqline.state == Editor && sqline.mainView.State == views.EditorInsert:
		fallthrough
	case sqline.state == Editor && sqline.mainView.State == views.EditorVisual:
		fallthrough
	case sqline.state == Editor && sqline.mainView.State == views.Editor:
		sqline.mainView.SetInfo([]rune(EditorInfo))
	case sqline.state == MainView && sqline.mainView.State == views.TblList:
		fallthrough
	case sqline.state == MainView && sqline.mainView.State == views.Indexes:
		sqline.mainView.SetInfo([]rune(TreeInfo))
	case sqline.state == MainView && sqline.mainView.State == views.DataTable:
		sqline.mainView.SetInfo([]rune(DataTableInfo))
	case sqline.state == MainView && sqline.mainView.State == views.SchemaList:
		fallthrough
	case sqline.state == MainView && sqline.mainView.State == views.DbList:
		sqline.mainView.SetInfo([]rune(ListInfo))
	case sqline.state == OpenConnView:
		sqline.mainView.SetInfo([]rune(OpenConnInfo))
	case sqline.state == NewConnView:
		sqline.mainView.SetInfo([]rune(NewConnInfo))
	}

}

func (sqline *Sqline) setDB(dbEntry util.DBEntry) {
	var err error
	var database db.Database
	switch dbEntry.Driver {
	case "sqlite3":
		database, err = db.CreateSqlite(dbEntry.ConnStr, sqline.mainView.TableFunc())
	case "postgres":
		database = db.CreatePg()
	}

	if err != nil {
		sqline.handleError(err)
		return
	}

	sqline.database = database
	showDB, showSchema := true, true
	databases, err := sqline.database.GetDatabases()
	if err != nil && errors.Is(db.ErrNotSupported, err) {
		showDB = false
	} else if err != nil {
		sqline.handleError(err)
		return
	} else {
		sqline.setDBInfo(databases)
	}

	schemas, err := sqline.database.GetSchemas()
	if errors.Is(db.ErrNotSupported, err) {
		showSchema = false
	} else if err != nil {
		sqline.handleError(err)
		return
	} else {
		sqline.mainView.SetSchemaList(schemas)
	}

	tables, err := sqline.database.GetTables()
	if err != nil {
		sqline.handleError(err)
		return
	}

	sqline.mainView.SetSQLFunc(sqline.database.GetExecSQLFunc())
	sqline.mainView.SetTableTree(tables)
	sqline.mainView.SetIndexTree(tables)
	sqline.mainView.SetVisibleComponents(showDB, showSchema, sqline.screen)
}

func (sqline *Sqline) setDBInfo(dbInfo []db.DbInfo) {
	sqline.mainView.SetDatabaseList(dbInfo)
}

func quit(screen tcell.Screen) {
	maybePanic := recover()
	screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

func Run() {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	err = screen.Init()
	if err != nil {
		panic(err)
	}

	maxX, maxY = screen.Size()
	sqline := createSqline(maxX, maxY, screen)

	screen.SetStyle(defStyle)
	screen.EnablePaste()
	screen.Clear()
	defer quit(screen)

	sync := false
	var ev tcell.Event
	for {

		prevState := sqline.state
		ev = screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()
			maxX, maxY = screen.Size()
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'Q':
				screen.Fini()
				return
			case ev.Key() == tcell.KeyEsc && sqline.mainView.EditorInNormalMode():
				sqline.state = NormalMode
				sqline.setInfo()
				screen.Fill(' ', defStyle)
				screen.Sync()
				sqline.mainView.SetStatus("Normal")
			case ev.Key() == tcell.KeyEsc && !sqline.mainView.EditorInNormalMode():
				sqline.mainView.HandleInput(ev)
			case ev.Rune() == 'e' && sqline.state == NormalMode:
				sqline.state = Editor
				sqline.mainView.SetState(views.Editor)
				sqline.setInfo()
			case ev.Rune() == 't' && sqline.state == NormalMode:
				sqline.state = MainView
				sqline.mainView.SetState(views.TblList)
				sqline.setInfo()
			case ev.Rune() == 'd' && sqline.state == NormalMode:
				sqline.state = MainView
				sqline.mainView.SetState(views.DataTable)
				sqline.setInfo()
			case ev.Rune() == 's' && sqline.state == NormalMode:
				sqline.state = MainView
				sqline.mainView.SetState(views.SchemaList)
				sqline.setInfo()
			case ev.Rune() == 'D' && sqline.state == NormalMode:
				sqline.state = MainView
				sqline.mainView.SetState(views.DbList)
				sqline.setInfo()
			case ev.Rune() == 'i' && sqline.state == NormalMode:
				sqline.state = MainView
				sqline.mainView.SetState(views.Indexes)
				sqline.setInfo()
			case ev.Rune() == 'A' && sqline.state == NormalMode:
				sqline.state = NewConnView
				sqline.mainView.SetStatus("NewConn")
				sqline.setInfo()
			case ev.Rune() == 'C' && sqline.state == NormalMode:
				sqline.state = OpenConnView
				sqline.mainView.SetStatus("OpenConn")
				sqline.setInfo()
			default:
				switch sqline.state {
				case NewConnView:
					sqline.newConnView.HandleInput(ev)
				case OpenConnView:
					sqline.openConnView.HandleInput(ev)
				case Editor:
					if ev.Rune() == '\t' {
						sync = true
					}
					fallthrough
				case MainView:
					sqline.mainView.HandleInput(ev)
				}
			}

			if prevState != sqline.state {
				sqline.ResetViews()
			}
		}

		sqline.mainView.Render(screen)
		switch sqline.state {
		case NewConnView:
			sqline.newConnView.Render(screen)
		case OpenConnView:
			sqline.openConnView.Render(screen)
		}

		if sync {
			screen.Sync()
		}

		screen.Show()
	}
}

func (sqline *Sqline) ResetViews() {
	sqline.newConnView.Reset()
}

func (sqline *Sqline) CalcPopupSize() {
	if sqline.maxX < sqline.pWidth {
		sqline.pWidth = sqline.maxX
	}
	if sqline.maxY < sqline.pHeight {
		sqline.pHeight = sqline.maxY
	}

	if sqline.pWidth == sqline.maxX {
		sqline.pLeft = 0
		sqline.pRight = sqline.maxX
	} else {
		pad := (sqline.maxX - sqline.pWidth) / 2
		sqline.pLeft = pad
		sqline.pRight = pad + pWidth
	}

	if sqline.pHeight == sqline.maxY {
		sqline.pTop = 0
		sqline.pBottom = sqline.maxY
	} else {
		pad := (sqline.maxY - sqline.pHeight) / 2
		sqline.pTop = pad
		sqline.pBottom = pad + sqline.pHeight
	}
}
