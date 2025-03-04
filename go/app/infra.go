package app

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
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
	SearchByName(ctx context.Context, keyword string) ([]Item, error)
}

// func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
// 	file, err := os.Open(i.fileName)
// 	if err != nil {
// 		if errors.Is(err, os.ErrNotExist) {
// 			// ファイルが存在しない場合、空のリストを返す
// 			return []Item{}, nil
// 		}
// 		return nil, err
// 	}
// 	defer file.Close()

// 	var items []Item
// 	err = json.NewDecoder(file).Decode(&items)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return items, nil
// }

func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
	rows, err := i.db.QueryContext(ctx, `
		SELECT items.id, items.name, categories.name AS category, items.image_name
		FROM items
		JOIN categories ON items.category_id = categories.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	db *sql.DB
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository(dbPath string) (ItemRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return &itemRepository{db: db}, nil
}

// Insert inserts an item into the repository.
// func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
// 	// 既存のアイテムを取得
// 	items, err := i.GetAll(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	// デバッグログ
// 	slog.Info("Inserting item", "name", item.Name, "category", item.Category, "image", item.Image)

// 	// 新しいアイテムを追加
// 	items = append(items, *item)

// 	// JSONファイルに保存
// 	file, err := os.Create(i.fileName)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	return json.NewEncoder(file).Encode(items)
// }

func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// カテゴリ名から category_id を取得
	var categoryID int
	err := i.db.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ?", item.Category).Scan(&categoryID)
	if err != nil {
		return err
	}

	_, err = i.db.ExecContext(ctx, "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
		item.Name, categoryID, item.Image)
	return err
}

// GetByID retrieves an item by its index (1-based).
func (i *itemRepository) GetByID(ctx context.Context, itemID int) (*Item, error) {
	row := i.db.QueryRowContext(ctx, `
		SELECT items.id, items.name, categories.name AS category, items.image_name
		FROM items
		JOIN categories ON items.category_id = categories.id
		WHERE items.id = ?
	`, itemID)

	var item Item
	if err := row.Scan(&item.ID, &item.Name, &item.Category, &item.Image); err != nil {
		return nil, err
	}
	return &item, nil
}

func (i *itemRepository) SearchByName(ctx context.Context, keyword string) ([]Item, error) {
	rows, err := i.db.QueryContext(ctx, `
		SELECT items.id, items.name, categories.name AS category, items.image_name
		FROM items
		JOIN categories ON items.category_id = categories.id
		WHERE items.name LIKE ?
	`, "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}
