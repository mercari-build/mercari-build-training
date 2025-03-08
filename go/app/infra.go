package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"

	// STEP 5-1: uncomment this line
	_ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID   int    `db:"id" json:"-"`
	Name string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	ImageName string `db:"image_name" json:"image_name"`
}

type JsonFormat struct {
	Items []Item `json:"items"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	GetItems(ctx context.Context) ([]Item, error)
	Insert(ctx context.Context, item *Item) error
	GetFileName() string
	GetItemByKeyword(keyword string) ([]Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	// fileName is the path to the JSON file storing items.
	fileName string
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	return &itemRepository{fileName: "items.json"}
}

// GetItems returns all items in the repository.
func (i *itemRepository) GetItems(ctx context.Context) ([]Item, error) {
	// open db
	db, err := sql.Open("sqlite3", "./db/mercari.sqlite3")
	if err != nil {
		slog.Error("failed to open database: ", "error", err)
		return nil, err
	}
	defer db.Close()

	// read items from db
	rows, err := db.Query("SELECT * FROM items")
	if err != nil {
		slog.Error("failed to prepare statement: ", "error", err)
		return nil, err
	}
	defer rows.Close()
	
	var items []Item
	for rows.Next() {
		var id int
		var name string
		var category string
		var imageName string
		err = rows.Scan(&id, &name, &category, &imageName)
		if err != nil {
			slog.Error("failed to scan rows: ", "error", err)
			return nil, err
		}
		items = append(items, Item{ID: id, Name: name, Category: category, ImageName: imageName})
	}
	err = rows.Err()
	if err != nil {
		slog.Error("failed to scan rows: ", "error", err)
		return nil, err
	}

	return items, nil
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {

	// open db
	db, err := sql.Open("sqlite3", "./db/mercari.sqlite3")
	if err != nil {
		slog.Error("failed to open database: ", "error", err)
		return err
	}
	defer db.Close()

	// insert item
	_, err = db.Exec("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)", item.Name, item.Category, item.ImageName)
	if err != nil {
		slog.Error("failed to insert item: ", "error", err)
		return err
	}

	return nil
}

func (i *itemRepository) GetFileName() string {
	return i.fileName
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	// store image
	file, err := os.Create(fileName)
	if err != nil {
		slog.Error("failed to create image file: ", "error", err)
		// return
	}
	defer file.Close()
	
	_, err = file.Write(image)
	if err != nil {
		slog.Error("failed to write image: ", "error", err)
		// return
	}

	return nil
}

func (i *itemRepository) GetItemByKeyword(keyword string) ([]Item, error) {
	// open db
	db, err := sql.Open("sqlite3", "./db/mercari.sqlite3")
	if err != nil {
		slog.Error("failed to open database: ", "error", err)
		return nil, err
	}
	defer db.Close()

	// read items from db
	rows, err := db.Query("SELECT * FROM items WHERE name = ?", keyword)
	if err != nil {
		slog.Error("failed to prepare statement: ", "error", err)
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var id int
		var name string
		var category string
		var imageName string
		err = rows.Scan(&id, &name, &category, &imageName)
		if err != nil {
			slog.Error("failed to scan rows: ", "error", err)
			return nil, err
		}
		items = append(items, Item{ID: id, Name: name, Category: category, ImageName: imageName})
	}
	err = rows.Err()
	if err != nil {
		slog.Error("failed to scan rows: ", "error", err)
		return nil, err
	}
	return items, nil
}