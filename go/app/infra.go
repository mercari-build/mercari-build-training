package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	// STEP 5-1: uncomment this line
	//"github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image    string `db:"image" json:"image"`
}

type jsonData struct {
	Items []Item `json:"items"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) ([]Item, error)
	GetByID(ctx context.Context, id int) (*Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	// fileName is the path to the JSON file storing items.
	fileName string
	mu       sync.Mutex
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	relPath := "../../items.json"

	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get working directory: %w", err))
	}

	filePath := filepath.Join(dir, relPath)

	return &itemRepository{fileName: filePath}

}
func (i *itemRepository) GetByID(ctx context.Context, id int) (*Item, error) {
	items, err := i.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if id < 0 || id >= len(items) {
		return nil, fmt.Errorf("item not found")
	}
	return &items[id], nil
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// STEP 4-2: add an implementation to store an item
	i.mu.Lock()
	defer i.mu.Unlock()

	items, err := i.GetAll(ctx)
	if err != nil {
		return err
	}
	item.ID = len(items)

	data, err := os.ReadFile(i.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			data = []byte(`{"items": []}`)
		} else {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}

	var jsonData jsonData
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	//add ID
	jsonData.Items = append(jsonData.Items, *item)

	//parse JSON to struct
	newData, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	//save to JSON files
	if err := os.WriteFile(i.fileName, newData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil

}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image
	relPath := "images"

	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get working directory: %w", err))
	}

	imageDirPath := filepath.Join(dir, relPath)

	filePath := filepath.Join(imageDirPath, fileName)

	//save file to the filepath
	err = os.WriteFile(filePath, image, 0644)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}
	return nil
}

func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
	//read content from file as data(string)
	data, err := os.ReadFile(i.fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)

	}
	var jsonData jsonData
	//parse JSON to struct(change data to jsonData(struct))
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	//parsed Item list is in jsonData.Items
	return jsonData.Items, nil
}
