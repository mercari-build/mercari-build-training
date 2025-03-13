package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// STEP 5-1: uncomment this line
	_ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image    string `db:"image_name" json:"image"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) ([]Item, error)
	GetItemByID(ctx context.Context, id int) (*Item, error)
	GetKeyword(ctx context.Context, keyword string) ([]Item, error)
	AddItem(ctx context.Context, item *Item) (*Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	db *sql.DB
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() (*itemRepository, error) {
	dbPath := "db/items.db"
	logger := log.New(os.Stdout, "ItemRepository: ", log.LstdFlags)
	logger.Println("Opening database at:", dbPath)

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist at path: %s", dbPath)
	}

	//conntect to database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	//confirm if database runs correctly (test connection)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &itemRepository{db: db}, nil

}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	var categoryID int
	err := i.db.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ?", item.Category).Scan(&categoryID)
	if err == sql.ErrNoRows {
		// if the category does not exist, add the category as new category
		result, err := i.db.ExecContext(ctx, "INSERT INTO categories (name) VALUES (?)", item.Category)
		if err != nil {
			return fmt.Errorf("failed to insert category: %w", err)
		}
		lastID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last inserted category ID: %w", err)
		}
		categoryID = int(lastID)
	} else if err != nil {
		return fmt.Errorf("failed to fetch category ID: %w", err)
	}
	if item.Name == "" || item.Category == "" {
		return ErrInvalidInput
	}

	// insert the item to items.table
	query := "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)"
	result, err := i.db.ExecContext(ctx, query, item.Name, categoryID, item.Image)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}

	// get the inserted item's id
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last inserted ID: %w", err)
	}
	item.ID = int(id)

	return nil
}

// AddItem adds an item and returns it.
func (i *itemRepository) AddItem(ctx context.Context, item *Item) (*Item, error) {
	// use the Insert method to insert data
	if err := i.Insert(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {

	relPath := "images"
	filePath := filepath.Join(relPath, fileName)

	//save file to the filepath
	err := os.WriteFile(filePath, image, 0644) //0644...permit writing
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}
	return nil
}

func (i *itemRepository) GetItemByID(ctx context.Context, id int) (*Item, error) {
	query := `
	SELECT items.id, items.name, COALESCE(categories.name, '') AS category,items.image_name
	FROM items
	LEFT JOIN categories ON items.category_id = categories.id
	WHERE items.id = ?
	`
	row := i.db.QueryRowContext(ctx, query, id)

	var item Item
	err := row.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //when item not found, return nil
		}
		return nil, fmt.Errorf("failed to fetch item: %w", err)
	}
	return &item, nil

}

func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
	query := `
        SELECT items.id, items.name, COALESCE(categories.name, '') AS category, items.image_name
        FROM items
        LEFT JOIN categories ON items.category_id = categories.id
    `
	//receive data from database
	rows, err := i.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch items: %w", err)
	}
	defer rows.Close()

	//get the each row's data
	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (i *itemRepository) GetKeyword(ctx context.Context, keyword string) ([]Item, error) {
	query := `
        SELECT items.id, items.name, COALESCE(categories.name, '') AS category, items.image_name
        FROM items
        LEFT JOIN categories ON items.category_id = categories.id
        WHERE items.name LIKE ? OR categories.name LIKE ?
    `

	rows, err := i.db.QueryContext(ctx, query, "%"+keyword+"%", "%"+keyword+"%")
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image); err != nil {
			return nil, fmt.Errorf("failed to parse database result: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return items, nil
}
