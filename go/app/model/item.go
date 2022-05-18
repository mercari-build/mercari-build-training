package model

import (
	"database/sql"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Items struct {
	Items []Item `json:"items"`
}

func GetItems(db *sql.DB) ([]Item, error) {
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

func AddItem(item Item, db *sql.DB) error {
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

func SearchItem(name string, db *sql.DB) ([]Item, error) {
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
