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
	FindAll(ctx context.Context) ([]Item, error)
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

// STEP 4-2: 新しい商品を登録する
// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// STEP 4-2: add an implementation to store an item

	// 1. 既存のアイテムを読み込む
	existingItems := []Item{}
	data, err := os.ReadFile(i.fileName)
	if err == nil {
		if err := json.Unmarshal(data, &existingItems); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	// 2. アイテムをリストに追加
	existingItems = append(existingItems, *item)

	// 3. JSON に変換
	data, err = json.MarshalIndent(existingItems, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// 4. `items.json` に書き込む
	err = os.WriteFile(i.fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// STEP 4-3: 商品一覧を取得する
// FindAll retrieves all items from the repository.
func (i *itemRepository) FindAll(ctx context.Context) ([]Item, error) {
	// `items.json` を読み込む
	data, err := os.ReadFile(i.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			// ファイルが存在しない場合は空のリストを返す
			return []Item{}, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// JSON を `[]Item` に変換
	var items []Item
	if len(data) > 0 {
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	return items, nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}
