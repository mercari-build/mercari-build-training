package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

/*
getDatabasePath()

return absolute path of mercari.sqlite3 database file
which always will be located one up level of go directory.
*/
func openDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath) // Adjust the relative path as needed
	if err != nil {
		return nil, err
	}
	if err := createTableIfNotExists(db); err != nil {
		return nil, err
	}
	return db, nil
}

func createTableIfNotExists(db *sql.DB) error {
	sch, err := os.ReadFile(dbSchemaPath)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(sch))
	return err
}
