package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	Get(ctx context.Context) (json.RawMessage, error)
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
	type ItemsWrapper struct {
		Items []Item `json:"items"`
	}

	var wrapper ItemsWrapper
	data, err := os.ReadFile(i.fileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &wrapper); err != nil {
			return err
		}
	}

	wrapper.Items = append(wrapper.Items, Item{Name: item.Name, Category: item.Category})

	output, err := json.Marshal(wrapper)
	if err != nil {
		return err
	}

	err = os.WriteFile(i.fileName, output, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (i *itemRepository) Get(ctx context.Context) (json.RawMessage, error) {
	type ItemsWrapper struct {
		Items []Item `json:"items"`
	}

	var wrapper ItemsWrapper
	data, err := os.ReadFile(i.fileName)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &wrapper); err != nil {
			return nil, err
		}
	}

	output, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}
