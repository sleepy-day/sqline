package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db      *sqlx.DB
	connStr string
	driver  string
}

func CreatePg() *Postgres {
	return &Postgres{}
}

func (psql *Postgres) Info() (string, string) {
	return psql.driver, psql.connStr
}

func (psql *Postgres) Initialize(connStr string) error {
	psql.driver = "postgres"
	psql.connStr = connStr

	var err error
	psql.db, err = sqlx.Connect(psql.driver, psql.connStr)
	return err
}

func (psql *Postgres) GetDatabases() ([]DbInfo, error) {
	var dbs []DbInfo
	err := psql.db.Select(&dbs, `
		SELECT
			datname AS Name,
			rolename AS Owner
		FROM
			pg_database
		INNER JOIN
			pg_roles
		ON
			pg_database.datdba = pg_roles.oid;
	`)

	return dbs, err
}

func (psql *Postgres) GetSchemas() ([]SchemaInfo, error) {
	var s []SchemaInfo
	err := psql.db.Select(&s, `
		SELECT
			nspname AS Name,
			rolname AS Owner
		FROM
			pg_catalog.pg_namespace s
		INNER JOIN
			pg_roles r
		ON
			s.nspowner = r.oid;
	`)

	return s, err
}

func (psql *Postgres) GetTables() ([]TableInfo, error) {
	return nil, nil
}

func (psql *Postgres) GetRoles() ([]RoleInfo, error) {
	return nil, nil
}
