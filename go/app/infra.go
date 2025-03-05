package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID        int    `db:"id" json:"-"`
	Name      string `db:"name" json:"name"`
	Category  string `db:"category" json:"category"`
	ImageName string `db:"image_name" json:"image_name"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	FindAll(ctx context.Context) ([]Item, error)
	FindByID(ctx context.Context, id int) (*Item, error)
	Search(ctx context.Context, keyword string) ([]Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	// db is a database connection
	db *sql.DB
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository(db *sql.DB) ItemRepository {
	return &itemRepository{db: db}
}

// 新しい商品を登録する
// Insert inserts an item into the repository.
func (r *itemRepository) Insert(ctx context.Context, item *Item) error {
	query := `INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, item.Name, item.Category, item.ImageName)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}
	item.ID = int(id)
	return nil
}

// 商品一覧を取得する
// FindAll retrieves all items from the repository.
func (r *itemRepository) FindAll(ctx context.Context) ([]Item, error) {
	query := `SELECT id, name, category, image_name FROM items`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// 商品の詳細を返す
func (r *itemRepository) FindByID(ctx context.Context, id int) (*Item, error) {
	query := `SELECT id, name, category, image_name FROM items WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var item Item
	if err := row.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("item with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve item: %w", err)
	}
	return &item, nil

}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image_name []byte) error {
	err := os.WriteFile(fileName, image_name, 0644)
	if err != nil {
		return fmt.Errorf("failed to write image file: %w", err)
	}
	return nil
}
