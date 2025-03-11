package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"

	// STEP 5-1: uncomment this line
	_ "github.com/mattn/go-sqlite3"
)

const (
	// it's flag to switch between JSON and DB implementation.
	// please set it to false to use JSON implementation.
	// trainee don't need to implement this flag.
	useDB = true
)

var errImageNotFound = errors.New("image not found")

// Items is a struct to store a list of items to json.
type Items struct {
	Items []*Item `json:"items"`
}

type Item struct {
	ID        int    `db:"id" json:"-"`
	Name      string `db:"name" json:"name"`
	Category  string `db:"category" json:"category"`
	ImageName string `db:"image_name" json:"image_name"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	SelectAll(ctx context.Context) ([]*Item, error)
	GetItem(ctx context.Context, id int) (*Item, error)
	SearchFromName(ctx context.Context, name string) ([]*Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	db *sql.DB
	// fileName is the path to the JSON file storing items.
	fileName string
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	db, err := sql.Open("sqlite3", "./db/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	// TODO: How should I close db ...
	return &itemRepository{
		db:       db,
		fileName: "items.json",
	}
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	if !useDB {
		return i.insertToFile(ctx, item)
	}

	_, err := i.db.Exec(
		"INSERT INTO item (name, category, image_name) VALUES (?, ?, ?)",
		item.Name, item.Category, item.ImageName,
	)
	return err
}

func (i *itemRepository) GetItem(ctx context.Context, id int) (*Item, error) {
	if !useDB {
		items, err := i.SelectAll(ctx)
		if err != nil {
			return nil, err
		}
		if items == nil || len(items) <= id {
			return nil, err
		}
		return items[id], nil
	}

	var item Item
	err := i.db.QueryRow("SELECT id, name, category, image_name FROM item WHERE id = ?", id).Scan(&item.ID, &item.Name, &item.Category, &item.ImageName)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Insert inserts an item into the repository.
func (i *itemRepository) SelectAll(ctx context.Context) ([]*Item, error) {
	// Added this to leave the code for the JSON implementation.

	if !useDB {
		items, err := i.getItemsFromFile(ctx)
		return items, err
	}

	rows, err := i.db.Query("SELECT id, name, category, image_name FROM item")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	items := []*Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName); err != nil {
			log.Fatal(err)
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (i *itemRepository) SearchFromName(ctx context.Context, name string) ([]*Item, error) {
	if !useDB {
		return nil, errors.New("not implemented")
	}

	rows, err := i.db.Query("SELECT id, name, category, image_name FROM item where name like ?", "%"+name+"%")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	items := []*Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName); err != nil {
			log.Fatal(err)
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (i *itemRepository) getItemsFromFile(ctx context.Context) ([]*Item, error) {
	var items Items
	if _, err := os.Stat(i.fileName); err == nil {
		// File exists, open it for reading
		f, err := os.Open(i.fileName)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		// Decode existing items from the file
		if err := json.NewDecoder(f).Decode(&items); err != nil {
			return nil, err
		}
	} else if os.IsNotExist(err) {
		// File does not exist, initialize items list
		items.Items = []*Item{}
	} else {
		// Some other error occurred
		return nil, err
	}
	return items.Items, nil
}
func (i *itemRepository) insertToFile(ctx context.Context, item *Item) error {
	items, err := i.getItemsFromFile(ctx)

	if err != nil {
		return err
	}
	slog.Info("items before insert", "items", items)

	// Append the new item
	items = append(items, item)
	newItems := Items{Items: items}

	// Marshal items to JSON
	b, err := json.Marshal(newItems)
	if err != nil {
		return err
	}

	// Open or create the file for writing
	f, err := os.Create(i.fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the JSON data to the file
	_, err = f.Write(b)
	if err != nil {
		return err
	}

	slog.Info("items after insert", "items", items)
	return nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(filePath string, image []byte) error {
	return os.WriteFile(filePath, image, 0644)
}
