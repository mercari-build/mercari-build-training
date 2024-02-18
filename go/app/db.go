package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func createTableIfNotExists(db *sql.DB) error {
	// Create table if not exists
	cmd_check := "SELECT * FROM items;"
	_, err := db.Exec(cmd_check)
	if err != nil {
		// Table not exists, create table
		file, err := os.Open(ItemsSchemaPath)
		if err != nil {
			return err
		}
		defer file.Close()
		var schema string
		if _, err := file.Read([]byte(schema)); err != nil {
			return err
		}
		_, err = db.Exec(schema) // CREATE TABLE items(...);
		if err != nil {
			return err
		}
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
