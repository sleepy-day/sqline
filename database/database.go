package database

import "github.com/jmoiron/sqlx"

type Database interface {
	Info() (string, string)
	Initialize(connStr string) error
	GetDatabases() ([]DbInfo, error)
	GetSchemas() ([]SchemaInfo, error)
	GetTables() ([]TableInfo, error)
	GetRoles() ([]RoleInfo, error)
}

type DbInfo struct {
	Name  string
	Owner string
}

type SchemaInfo struct {
	Name     string
	Owner    string
	Database string
}

type TableInfo struct {
	Name        string
	Owner       string
	Schema      string
	Database    string
	Index       bool
	Rules       bool
	Triggers    bool
	RowSecurity bool
}

type RoleInfo struct {
	Name string
}

func TestConnection(driver, connStr string) error {
	_, err := sqlx.Connect(driver, connStr)
	return err
}
