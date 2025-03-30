package app
import (
	"context"
	"errors"
	//"encoding/json"
	"database/sql"
	"fmt"
	"os"
	//STEP 5-1: uncomment this line
	 _ "github.com/mattn/go-sqlite3"
)
var errImageNotFound = errors.New("image not found")
type Item struct {
	ID        int    `db:"id" json:"-"`
	Name      string `db:"name" json:"name"`
	CategoryID   int    `db:"category_id" json:"category_id"`
	CategoryName string `db:"category_name" json:"category_name"`
	ImageName string `db:"image_name" json:"image_name"`
}
// Category represents a category in the database
type Category struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
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
	GetCategories(ctx context.Context) ([]Category, error)
	GetCategoryByName(ctx context.Context, name string) (*Category, error)
	InsertCategory(ctx context.Context, name string) (*Category, error)
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
	const query = `INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, item.Name, item.CategoryID, item.ImageName)
	if err != nil {
		return fmt.Errorf("insert item failed: %w", err)
	}

	if id, err := result.LastInsertId(); err != nil {
		return fmt.Errorf("retrieving last insert ID failed: %w", err)
	} else {
		item.ID = int(id)
	}

	return nil
}



// InsertCategory inserts a new category into the repository.
func (r *itemRepository) InsertCategory(ctx context.Context, name string) (*Category, error) {
	// カテゴリを追加するSQLクエリ
	query := `INSERT INTO categories (name) VALUES (?) RETURNING id`
	result, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return nil, fmt.Errorf("insert category failed: %w", err)
	}

	// 挿入したカテゴリのIDを取得
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("retrieve last insert ID failed: %w", err)
	}

	return &Category{ID: int(id), Name: name}, nil
}
//商品一覧を取得する
// FindAll retrieves all items from the repository.
func (r *itemRepository) FindAll(ctx context.Context) ([]Item, error) {
	query := `
		SELECT i.id, i.name, i.category_id, c.name as category_name, i.image_name 
		FROM items i
		LEFT JOIN categories c ON i.category_id = c.id
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("retrieve items failed: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.CategoryName, &item.ImageName); err != nil {
			return nil, fmt.Errorf("scan item failed: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}
// 商品の詳細を返す
func (r *itemRepository) FindByID(ctx context.Context, id int) (*Item, error) {
	query := `
		SELECT i.id, i.name, i.category_id, c.name as category_name, i.image_name 
		FROM items i
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE i.id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var item Item
	if err := row.Scan(&item.ID, &item.Name, &item.CategoryID, &item.CategoryName, &item.ImageName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("item with id %d not found", id)
		}
		return nil, fmt.Errorf("retrieve item failed: %w", err)
	}
	return &item, nil

}

// 商品を検索する
func (r *itemRepository) Search(ctx context.Context, keyword string) ([]Item, error) {
	query := `
		SELECT i.id, i.name, i.category_id, c.name as category_name, i.image_name 
		FROM items i
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE i.name LIKE ? OR c.name LIKE ?
	`
	likeKeyword := "%" + keyword + "%"
	rows, err := r.db.QueryContext(ctx, query, likeKeyword, likeKeyword)
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.CategoryName, &item.ImageName); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// GetCategories retrieves all categories
func (r *itemRepository) GetCategories(ctx context.Context) ([]Category, error) {
	query := `SELECT id, name FROM categories`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}
	return categories, nil
}

// GetCategoryByName retrieves a category by name
func (r *itemRepository) GetCategoryByName(ctx context.Context, name string) (*Category, error) {
	query := `SELECT id, name FROM categories WHERE name = ?`
	row := r.db.QueryRowContext(ctx, query, name)

	var category Category
	if err := row.Scan(&category.ID, &category.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("category with name %s not found", name)
		}
		return nil, fmt.Errorf("failed to retrieve category: %w", err)
	}
	return &category, nil
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