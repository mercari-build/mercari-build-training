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

// Item represents an item in the database
type Item struct {
	ID         int    `db:"id" json:"-"`
	Name       string `db:"name" json:"name"`
	CategoryID int    `db:"category_id" json:"-"`
	Category   string `db:"category_name" json:"category"`
	Image      string `db:"image_name" json:"image_name"`
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
	// STEP 5-1: add WAL mode and cache options
	db, err := sql.Open("sqlite3", "mercari.sqlite3?_journal_mode=WAL&cache=shared&mode=rwc")
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
	rows, err := i.db.QueryContext(ctx, `
		SELECT i.id, i.name, i.category_id, c.name as category_name, i.image_name 
		FROM items i 
		JOIN categories c ON i.category_id = c.id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.Category, &item.Image)
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
	err := i.db.QueryRowContext(ctx, `
		SELECT i.id, i.name, i.category_id, c.name as category_name, i.image_name 
		FROM items i 
		JOIN categories c ON i.category_id = c.id 
		WHERE i.id = ?
	`, id).Scan(&item.ID, &item.Name, &item.CategoryID, &item.Category, &item.Image)
	if err == sql.ErrNoRows {
		return nil, errItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query item: %w", err)
	}
	return &item, nil
}

func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// First, ensure the category exists and get its ID
	var categoryID int
	err := i.db.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ?", item.Category).Scan(&categoryID)
	if err == sql.ErrNoRows {
		// Category doesn't exist, create it
		result, err := i.db.ExecContext(ctx, "INSERT INTO categories (name) VALUES (?)", item.Category)
		if err != nil {
			return fmt.Errorf("failed to insert category: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id for category: %w", err)
		}
		categoryID = int(id)
	} else if err != nil {
		return fmt.Errorf("failed to query category: %w", err)
	}

	// Now insert the item with the category ID
	result, err := i.db.ExecContext(ctx,
		"INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
		item.Name, categoryID, item.Image)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	item.ID = int(id)
	item.CategoryID = categoryID
	return nil
}

func (i *itemRepository) Search(ctx context.Context, keyword string) ([]byte, error) {
	rows, err := i.db.QueryContext(ctx, `
		SELECT i.id, i.name, i.category_id, c.name as category_name, i.image_name 
		FROM items i 
		JOIN categories c ON i.category_id = c.id 
		WHERE i.name LIKE ?
	`, "%"+keyword+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.Category, &item.Image)
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
