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

func GetItems() ([]Item, error) {
	var err error
	cmd := "SELECT * FROM items"
	rows, _ := db.Query(cmd)
	defer rows.Close()
	var item_list []Item
	for rows.Next() {
		var item Item
		var id uuid.UUID
		err = rows.Scan(&id, &item.Name, &item.Category)
		if err != nil {
			return nil, err
		}
		item_list = append(item_list, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return item_list, nil
}

func AddItem(item Item) error {
	id, err :=  uuid.NewUUID()
	if err != nil {
		return err
	}
	if db == nil {
		return err
	}
	stmt, err := db.Prepare("INSERT INTO items (id, name, category) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, item.Name, item.Category)
	if err != nil {
		return err
	}	
	return nil
}

func SearchItem(name string) ([]Item, error) {
	rows, err := db.Query("SELECT * FROM items WHERE name = $1", name)
	defer rows.Close()
	var item_list []Item
	//Add items to a list in a for roop
	for rows.Next() {
		var item Item
		var id uuid.UUID
		// Add data
		err = rows.Scan(&id, &item.Name, &item.Category)
		if err != nil {
			return nil, err
		}
		item_list = append(item_list, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return item_list, nil
}

