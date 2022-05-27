package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"mercari-build-training-2022/app/config"
)

var DbConnection *sql.DB

func init() {
	var err error
	if env := os.Getenv("ENV"); env == "test" {
		DbConnection, err = sql.Open(config.Config.SQLDriver, config.Config.TestDbName)
	} else {
		DbConnection, err = sql.Open(config.Config.SQLDriver, config.Config.DbName)
	}
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf(os.Getenv("Exec db cmd"))
	// CREATE DB TABLES
	cmd := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS [users] (
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            name STRING UNIQUE NOT NULL,
            password STRING
        );
		CREATE TABLE IF NOT EXISTS [items] (
			id INTEGER PRIMARY KEY NOT NULL,
			name STRING,
			category STRING,
			image STRING,
            price INTEGER,
            price_lower_limit INTEGER, 
            user_id INTEGER,
            foreign key (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
        CREATE TABLE IF NOT EXISTS transaction_statuses(
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            name STRING
        );
        CREATE TABLE IF NOT EXISTS transactions(
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            determined_price INTEGER,
            item_id INTEGER NOT NULL,
            buyer_id INTNEGER NOT NULL,
            transaction_status_id INTEGER NOT NULL,
            foreign key (item_id) REFERENCES items(id) ON DELETE CASCADE,
            foreign key (buyer_id) REFERENCES users(id) ON DELETE CASCADE,
            foreign key (transaction_status_id) REFERENCES transaction_statuses(id) ON DELETE CASCADE,
            UNIQUE (item_id, buyer_id)
        );
        CREATE TABLE IF NOT EXISTS qa_types(
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            label STRING
        );
        CREATE TABLE IF NOT EXISTS qas(
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            item_id INTEGER NOT NULL,
            question STRING,
            answer STRING,
            qa_type_id INTEGER NOT NULL,
            foreign key (item_id) REFERENCES items(id) ON DELETE CASCADE,
            foreign key (qa_type_id) REFERENCES qa_types(id) ON DELETE CASCADE,
            UNIQUE (item_id)
        )
        `)
	_, err = DbConnection.Exec(cmd)
	if err != nil {
		log.Fatalln(err)
	}
}
