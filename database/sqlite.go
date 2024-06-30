package database

import "github.com/jmoiron/sqlx"

type Sqlite struct {
	db      *sqlx.DB
	connStr string
	driver  string
}

func CreateSqlite() *Sqlite {
	return &Sqlite{}
}

func (lite *Sqlite) Info() (string, string) {
	return lite.driver, lite.connStr
}

func (lite *Sqlite) Initialize(connStr string) error {
	lite.driver = "sqlite3"
	lite.connStr = connStr

	var err error
	lite.db, err = sqlx.Connect(lite.driver, lite.connStr)
	return err
}
