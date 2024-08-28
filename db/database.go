package db

import (
	"database/sql"
	"errors"
	"regexp"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sleepy-day/sqline/components"
)

var (
	ErrNotSupported = errors.New("not supported on target database")
)

type Database interface {
	Info() (string, string)
	GetDatabases() ([]DbInfo, error)
	GetSchemas() ([]SchemaInfo, error)
	GetTables() ([]Table, error)
	GetRoles() ([]RoleInfo, error)
	GetExecSQLFunc() components.ExecSQLFunc
	Select(cmd string) ([][][]rune, error)
	Exec(cmd string) ([]rune, error)
}

type DbInfo struct {
	Name  string `db:"Name"`
	Owner string `db:"Owner"`
}

type SchemaInfo struct {
	Name     string `db:"Name"`
	Owner    string `db:"Owner"`
	Database string `db:"Database"`
}

type TableInfo struct {
	Name        string `db:"Name"`
	Owner       string `db:"Owner"`
	Schema      string `db:"Schema"`
	Database    string `db:"Database"`
	Index       bool   `db:"Index"`
	Rules       bool   `db:"Rules"`
	Triggers    bool   `db:"Triggers"`
	RowSecurity bool   `db:"RowSecurity"`
}

type Column struct {
	Name         string
	Type         string
	NotNull      bool
	DefaultValue *string
	PrimaryKey   bool
	FKTo         *string
	OnUpdate     *string
	OnDelete     *string
}

type Table struct {
	Name    string
	Columns []Column
	Indexes []Index
}

type TableData struct {
	TableName    string  `db:"TableName"`
	ColumnName   string  `db:"ColumnName"`
	Type         string  `db:"Type"`
	NotNull      bool    `db:"NotNull"`
	DefaultValue *string `db:"DefaultValue"`
	PrimaryKey   bool    `db:"PrimaryKey"`
	FKTo         *string `db:"FKTo"`
	OnUpdate     *string `db:"OnUpdate"`
	OnDelete     *string `db:"OnDelete"`
}

type IndexData struct {
	TableName  string `db:"TableName"`
	IndexName  string `db:"IndexName"`
	Unique     bool   `db:"Unique"`
	Partial    bool   `db:"Partial"`
	SeqNo      int    `db:"SeqNo"`
	ColumnName string `db:"ColumnName"`
}

type Index struct {
	TableName string
	Name      string
	Cols      []IndexCol
}

type IndexCol struct {
	ColumnName string
	Unique     bool
	Partial    bool
	SeqNo      int
}

type RoleInfo struct {
	Name string
}

func TestConnection(driver, connStr string) error {
	_, err := sqlx.Connect(driver, connStr)
	return err
}

func convertRowsToRuneArr(rows *sql.Rows) ([][][]rune, error) {
	headers, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	table := [][][]rune{}
	headerArr := [][]rune{}

	for _, v := range headers {
		headerArr = append(headerArr, []rune(v))
	}
	table = append(table, headerArr)

	for rows.Next() {
		row := make([]interface{}, len(headers))
		for i := range headers {
			row[i] = new(sql.RawBytes)
		}

		err := rows.Scan(row...)
		if err != nil {
			panic(err)
		}

		rowRunes := [][]rune{}
		for _, cell := range row {
			rowRunes = append(rowRunes, []rune(string(*cell.(*sql.RawBytes))))
		}

		table = append(table, rowRunes)
	}

	return table, nil
}

func selectRegex() *regexp.Regexp {
	regex, _ := regexp.Compile(`\sselect\s`)
	return regex
}
