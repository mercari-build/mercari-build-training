package app

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image    string `db:"image" json:"image"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) ([]Item, error)
	GetByID(ctx context.Context, itemID int) (*Item, error)
}

func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
	file, err := os.Open(i.fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// ファイルが存在しない場合、空のリストを返す
			return []Item{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var items []Item
	err = json.NewDecoder(file).Decode(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
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
	// 既存のアイテムを取得
	items, err := i.GetAll(ctx)
	if err != nil {
		return err
	}

	// デバッグログ
	slog.Info("Inserting item", "name", item.Name, "category", item.Category, "image", item.Image)

	// 新しいアイテムを追加
	items = append(items, *item)

	// JSONファイルに保存
	file, err := os.Create(i.fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(items)
}

// GetByID retrieves an item by its index (1-based).
func (i *itemRepository) GetByID(ctx context.Context, itemID int) (*Item, error) {
	items, err := i.GetAll(ctx)
	if err != nil {
		slog.Error("Failed to get all items", "error", err)
		return nil, err
	}

	// itemIDは1から始まる想定
	if itemID < 1 || itemID > len(items) {
		slog.Warn("Item not found", "itemID", itemID, "totalItems", len(items))
		return nil, errors.New("item not found")
	}

	slog.Info("Item found", "itemID", itemID, "item", items[itemID-1])
	return &items[itemID-1], nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}
