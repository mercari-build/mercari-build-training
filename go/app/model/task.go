package model

import (
	"fmt"
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/google/uuid"
)

type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
}

type Items struct {
	Items []Item `json:"items"` 
}

var db *sql.DB

const dbSchema = "../db/items.db"
const dbSource = "../db/mercari.sqlite3"

func DBConnection() (*sql.DB, error) {
	// open database
	db_opened, err := sql.Open("sqlite3", dbSource)
	if err != nil {
		return nil, err
	}
	db = db_opened

	file, err := os.OpenFile(dbSchema, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	schema, err := os.ReadFile(dbSchema)
	if err != nil{
		return nil, err
	}

	_, err = db.Exec(string(schema))
	if err != nil{
		return nil, err
	}
	return db, nil
}

func GetItems([]Item, error) {

}

func AddItem(item Item) error {
	id, err :=  uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("Error: %s\n", "canot make a new uuid")
	}
	if db == nil {
		return fmt.Errorf("Error: %s\n", "db is nil")
	}
	stmt, err := db.Prepare("INSERT INTO items (id, name, category) VALUES (?,?,?)")
	if err != nil {
		return fmt.Errorf("Error: %s\n", "cannot use prepare function")
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, item.Name, item.Category)
	if err != nil {
		return fmt.Errorf("Error: %s\n", "cannot add a new item to db")
	}	
	return nil
}

