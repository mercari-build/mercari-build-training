package main

import (
	"database/sql"
	"os"

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

func loadItemById(db *sql.DB, id int) (*Item, error) {
	// Load item from db by id
	rows, err := db.Query("SELECT * FROM items WHERE items.id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var item Item
	if err := rows.Scan(&item.Id, &item.Name, &item.CategoryId, &item.ImageName); err != nil {
		return nil, err
	}
	return &item, nil
}

func loadItemsByQuery(db *sql.DB, query string) (*Items, error) {
	// Load items from db by query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Load items from rows
	var items *Items
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Id, &item.Name, &item.CategoryId, &item.ImageName); err != nil {
			return nil, err
		}
		items.Items = append(items.Items, item)
	}
	if err != nil {
		return nil, err
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

func loadCategoryById(db *sql.DB, id int) (*Category, error) {
	// Load category from db by id
	rows, err := db.Query("SELECT * FROM categories WHERE categories.id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var category Category
	if err := rows.Scan(&category.Id, &category.Name); err != nil {
		return nil, err
	}
	return &category, nil
}

func joinItemAndCategory(db *sql.DB, item Item) (*JoinedItem, error) {
	category, err := loadCategoryById(db, item.CategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, nil
	}

	joined_item := JoinedItem{Id: item.Id, Name: item.Name, ImageName: item.ImageName, CategoryName: category.Name}
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
