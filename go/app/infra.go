package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	// STEP 5-1: uncomment this line
	_ "github.com/mattn/go-sqlite3"
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
	GetAll(ctx context.Context) ([]byte, error)
	GetByID(ctx context.Context, id string) (*Item, error)
	Insert(ctx context.Context, item *Item) error
	Search(ctx context.Context, keyword string) ([]byte, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	db *sql.DB
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	db, err := sql.Open("sqlite3", "mercari.sqlite3")
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return nil
	}

	// Create tables if they don't exist
	schema, err := os.ReadFile("db/items.sql")
	if err != nil {
		slog.Error("failed to read schema file", "error", err)
		return nil
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		slog.Error("failed to create tables", "error", err)
		return nil
	}

	return &itemRepository{db: db}
}

func (i *itemRepository) GetAll(ctx context.Context) ([]byte, error) {
	rows, err := i.db.QueryContext(ctx, "SELECT id, name, category, image_name FROM items")
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate items: %w", err)
	}

	return json.Marshal(GetItemsResponse{Items: items})
}

func (i *itemRepository) GetByID(ctx context.Context, id string) (*Item, error) {
	var item Item
	err := i.db.QueryRowContext(ctx, "SELECT id, name, category, image_name FROM items WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Category, &item.Image)
	if err == sql.ErrNoRows {
		return nil, errItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query item: %w", err)
	}
	return &item, nil
}

func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	result, err := i.db.ExecContext(ctx,
		"INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
		item.Name, item.Category, item.Image)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	item.ID = int(id)
	return nil
}

func (i *itemRepository) Search(ctx context.Context, keyword string) ([]byte, error) {
	rows, err := i.db.QueryContext(ctx, "SELECT id, name, category, image_name FROM items WHERE name LIKE ?", "%"+keyword+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate items: %w", err)
	}

	return json.Marshal(GetItemsResponse{Items: items})
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
