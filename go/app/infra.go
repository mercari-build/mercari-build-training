package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	//"golang.org/x/tools/go/analysis/passes/nilfunc"
	//"path/filepath"
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
	// 既存のデータを読み込む
	var items []Item
	filename := "items.json"
	//file, err := os.Open("items.json")
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		slog.Error("Failed to open file", "error", err)
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&items); err != nil && err != io.EOF {
		slog.Error("Failed to decode JSON", "error", err)
		return err
	}
	// 追加するデータの内容をログ出力
	slog.Info("Adding new item", "name", item.Name, "category", item.Category)

	// 新しい item を追加
	items = append(items, *item) // 新しい item をitemsに追加*itemでポインタから値をとりだしている

	// JSONファイルに保存（上書き）
	//file, err = os.Create("items.json")
	//if err != nil {
	//	return err
	//}
	//defer file.Close()

	// ファイルに書き込む
	encoder := json.NewEncoder(file)
	slog.Error("Failed to encode JSON", "error", err)
	return encoder.Encode(items)

}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}
