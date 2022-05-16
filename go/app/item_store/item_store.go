package item_store

import (
	"database/sql"
	"fmt"
)

func InsertItem(name string, category string, image string) error {
	db, _ := sql.Open("sqlite3", "../db/mercari.sqlite3")
	defer db.Close()

	command := "INSERT INTO items (name, category, image) VALUES (?, ?, ?)"
	_, err := db.Exec(command, name, category, image)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func GetItems() (*sql.Rows, error) {
	db, _ := sql.Open("sqlite3", "../db/mercari.sqlite3")
	defer db.Close()

	command := "SELECT name, category, image FROM items"
	rows, err := db.Query(command)
	if err != nil {
		fmt.Println(err.Error())
	}
	return rows, err
}

func GetItemById(id int) *sql.Row {
	db, _ := sql.Open("sqlite3", "../db/mercari.sqlite3")
	defer db.Close()

	command := "SELECT name, category, image FROM items WHERE id = ?"
	row := db.QueryRow(command, id)
	return row
}

func SerchItems(keyword string) (*sql.Rows, error) {
	db, _ := sql.Open("sqlite3", "../db/mercari.sqlite3")
	defer db.Close()

	command := "SELECT name, category FROM items WHERE name LIKE ?"
	rows, err := db.Query(command, "%"+keyword+"%")
	if err != nil {
		fmt.Println(err.Error())
	}
	return rows, err
}
