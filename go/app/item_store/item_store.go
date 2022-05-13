package item_store

import (
	"database/sql"
	"fmt"
)

func InsertItem(name string, category string) error {
	db, _ := sql.Open("sqlite3", "../db/mercari.sqlite3")
	defer db.Close()

	command := "INSERT INTO items (name, category) VALUES (?, ?)"
	_, err := db.Exec(command, name, category)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func GetItems() (*sql.Rows, error) {
	db, _ := sql.Open("sqlite3", "../db/mercari.sqlite3")
	defer db.Close()

	command := "SELECT name, category FROM items"
	rows, err := db.Query(command)
	if err != nil {
		fmt.Println(err.Error())
	}
	return rows, err
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
