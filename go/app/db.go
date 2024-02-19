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

func loadItemsByQuery(db *sql.DB, field string, table string, condition string) (*Items, error) {
	// compose query
	query := "SELECT " + field + " FROM " + table
	if condition != "" {
		query += " WHERE " + condition
	}
	log.Infof("query: %s", query)

	// Load items from db by query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Load items from rows
	items, err := loadItemRows(rows)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func loadItemRows(rows *sql.Rows) (*Items, error) {
	// Load items from rows
	items := &Items{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Id, &item.Name, &item.CategoryId, &item.ImageName); err != nil {
			return nil, err
		}
		(*items).Items = append((*items).Items, item)
	}
	return items, nil
}

func insertItem(db *sql.DB, item Item) error {
	// Save new items to database
	cmd_ins := "INSERT INTO items(name, category_id, image_name) VALUES(?, ?, ?)"
	_, err := db.Exec(cmd_ins, item.Name, item.CategoryId, item.ImageName)
	if err != nil {
		return err
	}
	return nil
}

func joinItemAndCategory(db *sql.DB, item Item) (*JoinedItem, error) {
	// Join category name to item
	joined_item := JoinedItem{}
	rows, err := db.Query("SELECT categories.name FROM categories WHERE categories.id = ?", item.CategoryId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&joined_item.CategoryName); err != nil {
			return nil, err
		}
	}
	joined_item.Id = item.Id
	joined_item.Name = item.Name
	joined_item.ImageName = item.ImageName
	return &joined_item, nil

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
