package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errItemNotFound = errors.New("item not found")
var errImageNotFound = errors.New("image not found")

type Item struct {
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image    string `db:"image" json:"image_name"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, id string) (*Item, error)
	GetAll(ctx context.Context) (json.RawMessage, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	// fileName is the path to the JSON file storing items.
	fileName string
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("failed to get current working directory", "error", err)
		return &itemRepository{fileName: "items.json"}
	}

	fileName := filepath.Join(cwd, "items.json")
	slog.Debug("creating item repository", "path", fileName)
	return &itemRepository{fileName: fileName}
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// STEP 4-1: add an implementation to store an item

	var data struct {
		Items []*Item `json:"items"`
	}

	oldData, err := os.ReadFile(i.fileName)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(oldData) > 0 {
		if err := json.Unmarshal(oldData, &data); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	}

	data.Items = append(data.Items, item)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(i.fileName, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (i *itemRepository) GetAll(ctx context.Context) (json.RawMessage, error) {
	type ItemsWrapper struct {
		Items []*Item `json:"items"`
	}

	var wrapper ItemsWrapper
	data, err := os.ReadFile(i.fileName)
	if err != nil {
		slog.Debug("failed to read items.json", "error", err, "path", i.fileName)
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	if err == nil && len(data) > 0 {
		slog.Debug("read items.json", "content", string(data))
		if err := json.Unmarshal(data, &wrapper); err != nil {
			slog.Error("failed to unmarshal items.json", "error", err)
			return nil, err
		}
	}

	output, err := json.Marshal(wrapper)
	if err != nil {
		slog.Error("failed to marshal items", "error", err)
		return nil, err
	}

	slog.Debug("returning items", "output", string(output))
	return output, nil
}

func (i *itemRepository) GetByID(ctx context.Context, id string) (*Item, error) {
	type ItemsWrapper struct {
		Items []*Item `json:"items"`
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

	itemID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	if itemID <= 0 || itemID > len(wrapper.Items) {
		return nil, errItemNotFound
	}

	return wrapper.Items[itemID-1], nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image
	err := os.WriteFile(fileName, image, 0644)
	if err != nil {
		return err
	}
	return nil
}
