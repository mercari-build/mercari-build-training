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
	errItemNotFound = errors.New("item not found")
)
type Item struct {
	ID   	  int    `db:"id" json:"-"`
	Name 	  string `db:"name" json:"name"`
	Category  string `db:"category" json:"category"`
	ImageName string `db:"image_name" json:"image_name"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	List(ctx context.Context) ([]*Item, error)
	Select(ctx context.Context, id int) (*Item,error)  // リポジトリのinterfaceにselectを追加
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
	var data struct {
		Items []*Item `json:"items"` //jsonファイルの形式を読み、呼び出す
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

	data.Items = append(data.Items, item)  // 最後の行に追加したいデータを追加

	newData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if err := os.WriteFile(i.fileName, newData, 0644); err != nil { //書き込み
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// インターフェイス（関数名、引数、戻り値組み合わせ）と同じ関数名と引数と戻り値を指定する
// jsonファイルを読んで任意の構造体にして返す
func (i *itemRepository) List(ctx context.Context) ([]*Item, error) {
	var data struct {
		Items []*Item `json:"items"`
	}

	dataBytes, err := os.ReadFile(i.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return data.Items, nil
}

// selectの実装
func (i *itemRepository) Select(ctx context.Context, id int) (*Item, error) {
	// idが1以上の値以外になる場合はidをNotFoundにする
	if id <= 0 {
		return nil, errItemNotFound
	}

	items, err := i.List(ctx)
	if err != nil {
		return nil, err
	}

	// idがデータの量よりも大きかったらエラーを返す
	if len(items) < id {
		return nil, errItemNotFound
	}

	// 全部データを取った後、id-1番目を返す
	return items[id-1], nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
// 画像を保存する処理
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image
	if err := os.WriteFile(fileName, image, 0644); err != nil {
		return fmt.Errorf("failed to write image file: %w", err)
	}
	return nil
}
