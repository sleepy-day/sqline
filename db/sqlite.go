package db

import (
	"fmt"
	"regexp"

	"github.com/jmoiron/sqlx"
	"github.com/sleepy-day/sqline/components"
)

type Sqlite struct {
	db            *sqlx.DB
	connStr       string
	driver        string
	tableDataFunc func([][][]rune, []rune)
	selectRegex   *regexp.Regexp
}

func CreateSqlite(connStr string, tableFunc func([][][]rune, []rune)) (*Sqlite, error) {
	sqlite := &Sqlite{
		driver:        "sqlite3",
		connStr:       connStr,
		tableDataFunc: tableFunc,
		selectRegex:   selectRegex(),
	}

	var err error
	sqlite.db, err = sqlx.Connect(sqlite.driver, sqlite.connStr)

	return sqlite, err
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

func (lite *Sqlite) GetDatabases() ([]DbInfo, error) {
	return nil, ErrNotSupported
}

func (lite *Sqlite) GetSchemas() ([]SchemaInfo, error) {
	return nil, ErrNotSupported
}

func (lite *Sqlite) GetTables() ([]Table, error) {
	var tableData []TableData
	err := lite.db.Select(&tableData, `
		SELECT
			ss.name AS TableName,
			pti.name AS ColumnName,
			pti.type AS Type,
			"notnull" AS "NotNull",
			dflt_value AS DefaultValue,
			pk AS PrimaryKey,
			pfkl."to" AS FKTo,
			on_update AS OnUpdate,
			on_delete AS OnDelete
		FROM
			sqlite_schema ss
		INNER JOIN
			pragma_table_info(ss.name) pti
		LEFT JOIN
			pragma_foreign_key_list(ss.name) pfkl;
	`)
	if err != nil {
		return nil, err
	}

	var indexData []IndexData
	err = lite.db.Select(&indexData, `
		SELECT 
			ss.name AS TableName, 
			pil.name AS IndexName, 
			pil."unique" AS "Unique", 
			pil."partial" AS "Partial", 
			pii.seqno AS SeqNo, 
			pii.name AS ColumnName 
		FROM 
			sqlite_schema ss 
		INNER JOIN 
			pragma_index_list(ss.name) pil 
		INNER JOIN 
			pragma_index_info(pil.name) pii
		ORDER BY 
			pil.name DESC,
			pii.seqno DESC;
	`)
	if err != nil {
		return nil, err
	}

	indexes := map[string]Index{}
	for _, v := range indexData {
		indexVal, ok := indexes[v.IndexName]
		if !ok {
			newIndex := Index{
				Name:      v.IndexName,
				TableName: v.TableName,
				Cols: []IndexCol{{
					ColumnName: v.ColumnName,
					Unique:     v.Unique,
					Partial:    v.Partial,
					SeqNo:      v.SeqNo,
				}},
			}

			indexes[v.IndexName] = newIndex
			continue
		}

		indexVal.Cols = append(indexVal.Cols, IndexCol{
			ColumnName: v.ColumnName,
			Unique:     v.Unique,
			Partial:    v.Partial,
			SeqNo:      v.SeqNo,
		})

		indexes[v.IndexName] = indexVal
	}

	tableMap := make(map[string]Table)
	for _, v := range tableData {
		tableVal, ok := tableMap[v.TableName]

		if !ok {
			newTable := Table{
				Name: v.TableName,
				Columns: []Column{{
					Name:         v.ColumnName,
					Type:         v.Type,
					NotNull:      v.NotNull,
					DefaultValue: v.DefaultValue,
					PrimaryKey:   v.PrimaryKey,
					FKTo:         v.FKTo,
					OnUpdate:     v.OnUpdate,
					OnDelete:     v.OnDelete,
				},
				},
			}

			tableMap[v.TableName] = newTable
			continue
		}

		col := Column{
			Name:         v.ColumnName,
			Type:         v.Type,
			NotNull:      v.NotNull,
			DefaultValue: v.DefaultValue,
			PrimaryKey:   v.PrimaryKey,
			FKTo:         v.FKTo,
			OnUpdate:     v.OnUpdate,
			OnDelete:     v.OnDelete,
		}

		tableVal.Columns = append(tableVal.Columns, col)
		tableMap[v.TableName] = tableVal
	}

	for _, v := range indexes {
		tableVal, ok := tableMap[v.TableName]
		if ok {
			tableVal.Indexes = append(tableVal.Indexes, v)
			tableMap[v.TableName] = tableVal
		}
	}

	var finalTables []Table
	for _, v := range tableMap {
		finalTables = append(finalTables, v)
	}

	return finalTables, nil
}

func (lite *Sqlite) GetRoles() ([]RoleInfo, error) {
	return nil, ErrNotSupported
}

func (lite *Sqlite) Select(cmd string) ([][][]rune, error) {
	rows, err := lite.db.Query(cmd)
	if err != nil {
		return nil, err
	}

	table, err := convertRowsToRuneArr(rows)
	if err != nil {
		return nil, err
	}

	return table, nil
}

func (lite *Sqlite) Exec(cmd string) ([]rune, error) {
	result, err := lite.db.Exec(cmd)
	if err != nil {
		return nil, err
	}

	return []rune(fmt.Sprintf("%d rows affected", result.RowsAffected)), nil
}

func (lite *Sqlite) GetExecSQLFunc() components.ExecSQLFunc {
	return func(cmd []rune) error {
		if len(cmd) == 0 {
			return nil
		}

		cmdStr := string(cmd)
		var table [][][]rune = nil
		var result []rune = nil

		var err error

		table, err = lite.Select(cmdStr)
		if err != nil {
			return err
		}

		lite.tableDataFunc(table, result)
		return nil
	}
}
