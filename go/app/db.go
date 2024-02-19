package main

import (
	"database/sql"
	"os"

	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

func loadDb(path string) (*sql.DB, error) {
	// Open database
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := createTableIfNotExists(db); err != nil {
		return nil, err
	}
	return db, nil
}

func createTableIfNotExists(db *sql.DB) error {
	// Create table if not exists
	file, err := os.Open(DbSchemaPath)
		if err != nil {
			return err
		}
		defer file.Close()
		var schema string
		if _, err := file.Read([]byte(schema)); err != nil {
			return err
		}
	_, err = db.Exec(schema) // CREATE TABLE IF NOT EXIST ...;
		if err != nil {
			return err
	}
	return nil
}

func loadItems() (Items, error) {
	// Load items from ItemsTablePath
	db, err := sql.Open("sqlite3", ItemsTablePath)
	if err != nil {
		return Items{}, err
	}
	defer db.Close()

	if createTableIfNotExists(db) != nil {
		return Items{}, err
	}
	cmd_sel := "SELECT * FROM items"
	rows, err := db.Query(cmd_sel)
	if err != nil {
		return Items{}, err
	}
	defer rows.Close()

	var items Items
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Id, &item.Name, &item.Category, &item.ImageName); err != nil {
			return Items{}, err
		}
		items.Items = append(items.Items, item)
	}
	return items, nil
}

func insertItem(item Item) error {
	// Save new items to database
	db, err := sql.Open("sqlite3", ItemsTablePath)
	if err != nil {
		return err
	}
	defer db.Close()
	if createTableIfNotExists(db) != nil {
		return err
	}
	cmd_ins := "INSERT INTO items(id, name, category, image_name) VALUES(?, ?, ?, ?)"
	_, err = db.Exec(cmd_ins, item.Id, item.Name, item.Category, item.ImageName)
	if err != nil {
		return err
	}
	return nil
}
}

func joinItemsAndCategories(db *sql.DB) (*JoinedItems, error) {
	// Join category name to items
	joined_items := JoinedItems{}
	rows, err := db.Query("SELECT items.id, items.name, categories.name, items.image_name FROM items INNER JOIN categories ON items.category_id = categories.id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var joined_item JoinedItem
		if err := rows.Scan(&joined_item.Id, &joined_item.Name, &joined_item.ImageName, &joined_item.CategoryName); err != nil {
			return nil, err
		}
		joined_items.Items = append(joined_items.Items, joined_item)
	}
	return &joined_items, nil
}
