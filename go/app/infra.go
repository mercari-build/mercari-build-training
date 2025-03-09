package app

import (
	"context"
	"errors"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"crypto/sha256"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID   int    `db:"id" json:"-"`
	Name string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image string `db:"image" json:"image"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	FindAll(ctx context.Context)([]Item, error)
	FindID(ctx context.Context, id int)(*Item, error)
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
	itemSlice := []Item{}
	data, err := os.ReadFile(i.fileName)
	if err == nil {
		if err := json.Unmarshal(data, &itemSlice); err != nil {
			return fmt.Errorf("Could not parse JSON, %w", err)
		}
	}
	
	itemSlice = append(itemSlice, *item);
	data, err = json.MarshalIndent(itemSlice, "", " ")
	if err != nil {
		return fmt.Errorf("Could not marshal JSON, %w", err)
	}
	
	err = os.WriteFile(i.fileName, data, 0644)
	fmt.Println("filename %s", i.fileName)
	path, err := filepath.Abs(i.fileName)
	fmt.Println("File path to be used: %s", path)

	if err != nil {
		return fmt.Errorf("Could not write to file, %w", err)
	}

	return nil
}

func (i *itemRepository) FindAll(ctx context.Context) ([]Item, error) {
	data, err := os.ReadFile(i.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty slice
			return []Item{}, nil
		} 
		return nil, fmt.Errorf("Could not read file, %w", err)
	}

	var items []Item
	if len(data) > 0 {
		err := json.Unmarshal(data, &items)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling JSON, %w", err)
		}
	}
	return items, nil
}


// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image
	imgDir := "images"
	err := os.MkdirAll(imgDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error create image directory: %w", err)
	}
	
	hash := sha256.Sum256(image)
	hashFileName := fmt.Sprint("%x.jpg", hash)

	filepath := filepath.Join(imgDir, hashFileName)

	// Check if file already exists
	_, err = os.Stat(filepath)
	if err == nil {
		return nil 
	}

	err = os.WriteFile(filepath, image, 0644)
	if err != nil {
		return fmt.Errorf("Error writing image to file: %w", err)
	}
	
	return nil
}

// step 4-5 to findid
func (i *itemRepository) FindID(ctx context.Context, id int)(*Item, error) {
	items, err := i.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error retriving items: %w", err)
	}
	
	// Check valid ID
	if id > len(items) || id < 1 {
		return nil, fmt.Errorf("Item ID does not exist: %w", err)
	}

	//id--
	return &items[id], nil
}
