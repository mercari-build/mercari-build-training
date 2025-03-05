package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID   int    `db:"id" json:"-"`
	Name string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image_Name string `db:"image_name" json:"image_name"`
}

type JsonFormat struct {
	Items []Item `json:"items"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetFileName() string
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
	
	// open json file
	jsonFile, err := os.Open(i.fileName)

	if err != nil {
		slog.Error("failed to open jsonFile: ", "error", err)
		return err
	}
	defer jsonFile.Close()
	
	// read json file as bytes
	bytes, _ := io.ReadAll(jsonFile)

	// decode bytes to map
	var decodeData JsonFormat
	json.Unmarshal(bytes, &decodeData)

	// append new item
	decodeData.Items = append(decodeData.Items, *item)

	// create json file to write
	newJsonFile, err := os.Create(i.fileName)
	if err != nil {
		slog.Error("failed to create jsonFile: ", "error", err)
		return err
	}
	defer newJsonFile.Close()

	// encode and wrute to json file
	encoder := json.NewEncoder(newJsonFile)
	encoder.SetIndent("", "  ")

	return encoder.Encode(decodeData)
}

func (i *itemRepository) GetFileName() string {
	return i.fileName
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}
