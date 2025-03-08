package app

import (
	"context"
	"errors"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID   int    `db:"id" json:"-"`
	Name string `db:"name" json:"name"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
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
	// STEP 4-2: add an implementation to store an item
	_, err := os.Stat(i.fileName)

	if os.IsNotExist(err) {
		f, creationErr := os.Create(i.fileName)
		if creationErr != nil {
			return errors.New("Unable to create file")
		}
		defer f.Close()
		newItems := []*item{item}
		newItemsJSON, _ := json.Marshal(newItems)
		_, err := f.Write(newItemsJSON)
		if err != nil {
			return errors.New("Unable to write")
		}
	} else {
		var items []*Item
		f, openErr := os.OpenFile(i.fileName, os.O_RDWR, 0644)
		if openErr != nil {
			return errors.New("Unable TO open file")
		}
		defer f.Close()
		items, getErr := i.GettAllItems(ctx)
		if getErr != nil {
			return errors.New("Unable to get items")
		}
		itemsJSON, _ := json.Marshal(append(items, item))
		_, err = f.Write(itemsJSON)
		if err != nil {
			return errors.New("Unable to write")
		}
	}	

	return nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}
