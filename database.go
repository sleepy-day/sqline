package main

import (
	"log"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func ConnectToDb(provider, connStr string) error {
	var err error
	db, err = sqlx.Connect(provider, connStr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	return nil
}

func GetDatabases() ([]string, error) {
	var dbs []string
	err := db.Select(&dbs, "SELECT datname FROM pg_databases;")
	if err != nil {
		return nil, err
	}

	return dbs, nil
}
