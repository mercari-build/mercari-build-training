package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var (
	errImageNotFound = errors.New("image not found")
	errItemNotFound  = errors.New("item not found")
)

type Item struct {
	ID        int    `db:"id" json:"-"`
	Name      string `db:"name" json:"name"`
	Category  string `db:"category" json:"category"`
	ImageName string `db:"image_name" json:"image_name"` // STEP 4-4: add an image field
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	List(ctx context.Context) ([]*Item, error)
	Select(ctx context.Context, id int) (*Item, error)
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
	// Prepare an empty structure with Items
	data := &struct {
		Items []*Item `json:"items"`
	}{}

	oldData, err := os.ReadFile(i.fileName)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// If JSON exists, parse and read existing Items
	if len(oldData) > 0 {
		if err := json.Unmarshal(oldData, data); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	}

	// Add new item
	data.Items = append(data.Items, item)

	// Convert to JSON with indentation
	newData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file (0644 is rw-r--r-- permissions)
	if err := os.WriteFile(i.fileName, newData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// List list items from the repository
func (i *itemRepository) List(ctx context.Context) ([]*Item, error) {
	// Temporary structure for JSON file
	data := &struct {
		Items []*Item `json:"items"`
	}{}

	// Read file
	bytes, err := os.ReadFile(i.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal JSON if file is not empty
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	}

	return data.Items, nil
}

// Select gets item from repository by id
func (i *itemRepository) Select(ctx context.Context, id int) (*Item, error) {
	// Get entire list
	items, err := i.List(ctx)
	if err != nil {
		return nil, err
	}

	// Check if id is valid
	if id <= 0 || id > len(items) {
		return nil, errItemNotFound
	}

	// Return item
	return items[id-1], nil

}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image
	if err := os.WriteFile(fileName, image, 0644); err != nil {
		return fmt.Errorf("failed to write image file: %w", err)
	}

	return nil
}
