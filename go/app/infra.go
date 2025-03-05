package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
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

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// Added this to leave the code for the JSON implementation.
	if i.fileName != "" {
		return i.insertToFile(ctx, item)
	}

	return fmt.Errorf("SelectAll is not implemented")
}

// Insert inserts an item into the repository.
func (i *itemRepository) SelectAll(ctx context.Context) ([]*Item, error) {
	// Added this to leave the code for the JSON implementation.
	if i.fileName != "" {
		items, err := i.getItemsFromFile(ctx)
		return items, err
	}

	return nil, fmt.Errorf("SelectAll is not implemented")
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
