package main

import (
	"os"
	"fmt"
	"testing"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"mercari-build-training-2022/app/config"
)

func TestMain(m *testing.M) {
    // os.Exit skips defer calls
    // so we need to call another function
    code, err := run(m)
    if err != nil {
        fmt.Errorf("Create Test DB Error: %s", err)
    }
    os.Exit(code)
}

func run(m *testing.M) (code int, err error) {

	// create test db
    db, err := sql.Open(config.Config.SQLDriver, config.Config.TestDbName)
    if err != nil {
        return -1, fmt.Errorf("could not connect to database: %w", err)
    }

    // begin transaction
    tx, err := db.Begin()
    if err != nil {
        return -1, err
    }

    // rollback trannsaction after the test are run
    defer func() {
        tx.Rollback()
        db.Close()
    }()

    return m.Run(), nil
}