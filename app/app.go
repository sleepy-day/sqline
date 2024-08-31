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
		//TODO: pass error into error message display
	}

	sqline.config = conf
	sqline.mainView = views.CreateMainView(0, 0, maxX, maxY, true, true, buf, &defStyle, &hlStyle)
	sqline.newConnView = views.CreateNewConnView(sqline.pLeft, sqline.pTop, sqline.pRight, sqline.pBottom, &defStyle, &hlStyle, sqline.createTestFunc(), sqline.createSaveFunc())
	sqline.openConnView = views.CreateOpenConnView(sqline.pLeft, sqline.pTop, sqline.pRight, sqline.pBottom, &defStyle, &hlStyle, sqline.config.SavedConns, sqline.createSelectFunc())

	return &sqline
}

func (sqline *Sqline) handleError(err error) {
	panic(err)
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
	case "mssql":

	case "mysql":
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
			case ev.Key() == tcell.KeyCtrlC:
				screen.Fini()
				return
			case ev.Key() == tcell.KeyEsc && sqline.mainView.EditorInNormalMode():
				sqline.state = NormalMode
				screen.Fill(' ', defStyle)
				screen.Sync()
				sqline.mainView.SetStatus([]rune("Normal"))
			case ev.Key() == tcell.KeyEsc && !sqline.mainView.EditorInNormalMode():
				sqline.mainView.HandleInput(ev)
			case ev.Rune() == 'e' && sqline.state == NormalMode:
				sqline.state = Editor
				sqline.mainView.SetState(views.Editor)
			case ev.Rune() == 't' && sqline.state == NormalMode:
				sqline.state = MainView
				sqline.mainView.SetState(views.TblList)
			case ev.Rune() == 'd' && sqline.state == NormalMode:
				sqline.state = MainView
				sqline.mainView.SetState(views.DataTable)
			case ev.Rune() == 's' && sqline.state == NormalMode:
				sqline.state = MainView
				sqline.mainView.SetState(views.SchemaList)
			case ev.Rune() == 'D' && sqline.state == NormalMode:
				sqline.state = MainView
				sqline.mainView.SetState(views.DbList)
			case ev.Rune() == 'A' && sqline.state == NormalMode:
				sqline.state = NewConnView
				sqline.mainView.SetStatus([]rune("NewConn"))
			case ev.Rune() == 'C' && sqline.state == NormalMode:
				sqline.state = OpenConnView
				sqline.mainView.SetStatus([]rune("OpenConn"))
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
