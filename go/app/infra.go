package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	Image    string `db:"image" json:"image"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) ([]Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	db *sql.DB
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() (ItemRepository, error) {
	dbPath := "../../db/items.db"

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute database path: %w", err)
	}

	fmt.Println("Opening database at:", absPath)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist at path: %s", absPath)
	}
	//conntect to database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	err = db.Ping() //confirm if database runs correctly
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to the database")

	return &itemRepository{db: db}, nil

}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	//insert item struct to database
	result, err := i.db.ExecContext(ctx, "INSERT INTO items (name, category, image) VALUES (?, ?, ?)", item.Name, item.Category, item.Image)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}
	//get the automated ID(main key)
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	item.ID = int(id)

	return nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	relPath := "../../" //main.go → go/cmd/api/main.go、　images→go/images

	//convert to absolut path
	imageDirPath, err := filepath.Abs(relPath)
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	//make a directory to avoid error with no directory
	if err := os.MkdirAll(imageDirPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	filePath := filepath.Join(imageDirPath, fileName)

	//save file to the filepath
	err = os.WriteFile(filePath, image, 0644) //permit writing
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}
	return nil
}

func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
	//receive data from database
	rows, err := i.db.QueryContext(ctx, "SELECT id, name, category, image FROM items")
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
