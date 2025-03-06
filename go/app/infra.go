package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image    string `json:"image"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetItems(ctx context.Context) ([]Item, error)
	GetItemID(ctx context.Context, itemID int) (Item, error)
}

func (i *itemRepository) GetItemID(ctx context.Context, itemID int) (Item, error) {
	file, err := os.Open("items.json")

	if err != nil {
		return Item{}, fmt.Errorf("failed to open items file: %w", err)
	}
	defer file.Close()

	var data struct {
		Items []Item `json:"items"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return Item{}, fmt.Errorf("failed to parse items JSON: %w", err)
	}

	return data.Items[itemID], nil
}

func (i *itemRepository) GetItems(ctx context.Context) ([]Item, error) {
	var data struct {
		Items []Item `json:"items"`
	}

	if _, err := os.Stat(i.fileName); os.IsNotExist(err) {
		return []Item{}, nil
	}

	file, err := os.ReadFile(i.fileName)
	if err != nil {
		log.Println("error:", err)
		return nil, err
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Println("error:", err)
		return nil, err
	}

	return data.Items, nil
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
	// STEP 4-1: add an implementation to store an item
	var data struct {
		Items []Item `json:"items"`
	}

	if _, err := os.Stat(i.fileName); err == nil {
		file, err := os.ReadFile(i.fileName)
		if err != nil {
			log.Println("error:", err)
			return err
		}
		json.Unmarshal(file, &data)
	}

	data.Items = append(data.Items, *item)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println("error:", err)
		return err
	}

	err = os.WriteFile(i.fileName, jsonData, 0666)
	if err != nil {
		log.Println("error:", err)
		return err
	}

	return nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	imageDir := "images"

	if _, err := os.Stat(imageDir); os.IsNotExist(err) {
		err := os.MkdirAll(imageDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	hash := sha256.Sum256(image)
	hashString := hex.EncodeToString(hash[:])

	filePath := filepath.Join(imageDir, hashString+".jpg")

	err := os.WriteFile(filePath, image, 0666)
	if err != nil {
		return err
	}
	return nil
}
