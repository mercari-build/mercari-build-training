package app

import (
	"context"
	"database/sql"
	"log/slog"
	"errors"
	"fmt"
	"os"
	// STEP 5-1: uncomment this line
	_ "github.com/mattn/go-sqlite3"
)

var (
	errImageNotFound = errors.New("image not found")
	errItemNotFound = errors.New("item not found")
	errCategoryNotFound = errors.New("category not found")
)

/*
type Item struct {
	ID   	  int    `db:"id" json:"-"`
	Name 	  string `db:"name" json:"name"`
	Category  string `db:"category" json:"category"`
	ImageName string `db:"image_name" json:"image_name"`
}
*/

// カテゴリー構造体
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// アイテム構造体
type Item struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	List(ctx context.Context) ([]*Item, error)
	Select(ctx context.Context, id int) (*Item, error)  // リポジトリのinterfaceにselectを追加
	Search(ctx context.Context, keyword string) ([]*Item, error)
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

// items.sql を実行し、テーブルを作成する
func SetupDatabase(db *sql.DB) error {
	schema, err := os.ReadFile("db/items.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}
	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	slog.Info("Database setup complete")
	return nil
}

// `categories` から ID を取得（なければ新規作成）
func (i *itemRepository) getOrCreateCategoryID(ctx context.Context, categoryName string) (int, error) {
	var categoryID int
	err := i.db.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ?", categoryName).Scan(&categoryID)
	if err == sql.ErrNoRows {
		// カテゴリが存在しない場合、新しく追加
		result, err := i.db.ExecContext(ctx, "INSERT INTO categories (name) VALUES (?)", categoryName)
		if err != nil {
			return 0, fmt.Errorf("failed to insert category: %w", err)
		}
		lastInsertID, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to retrieve last insert ID for category: %w", err)
		}
		categoryID = int(lastInsertID)
	} else if err != nil {
		return 0, fmt.Errorf("failed to retrieve category ID: %w", err)
	}
	return categoryID, nil
}

// 5-1 Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// カテゴリ ID を取得（新規なら追加）
	categoryID, err := i.getOrCreateCategoryID(ctx, item.Category)
    if err != nil {
		slog.Error("failed to get category ID", "category", item.Category, "error", err)
        return err
    }
	query := `INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)`
	slog.Info("Executing insert query", "query", query, "name", item.Name, "category_id", categoryID, "image_name", item.ImageName)
	
	result, err := i.db.ExecContext(ctx, query, item.Name, categoryID, item.ImageName)
	if err != nil {
		slog.Error("failed to execute insert query", "error", err)
		return fmt.Errorf("failed to insert item: %w", err)
	}

	var id int64
	id, err = result.LastInsertId()
	if err != nil {
        slog.Error("failed to retrieve last insert ID", "error", err)
		return fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}
	item.ID = int(id) // ここで ID をセット
    slog.Info("Item inserted successfully", "id", item.ID)
	return nil
}

// インターフェイス（関数名、引数、戻り値組み合わせ）と同じ関数名と引数と戻り値を指定する
func (i *itemRepository) List(ctx context.Context) ([]*Item, error) {
	query := `
        SELECT i.id, i.name, c.name AS category, i.image_name 
        FROM items i
        JOIN categories c ON i.category_id = c.id
    `
	rows, err := i.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	return items, nil
}

// 5-1selectの実装
func (i *itemRepository) Select(ctx context.Context, id int) (*Item, error) {
	query := `
        SELECT i.id, i.name, c.name AS category, i.image_name 
        FROM items i
        JOIN categories c ON i.category_id = c.id
        WHERE i.id = ?
    `
	row := i.db.QueryRowContext(ctx, query, id)

	// idが1以上の値以外になる場合はidをNotFoundにする
	var item Item
	err := row.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errItemNotFound
	}
	if err != nil {
		slog.Error("failed to select item", "error", err)
		return nil, fmt.Errorf("failed to select item: %w", err)
	}
	return &item, err
}

// itemRepository の Search メソッド実装
func (i *itemRepository) Search(ctx context.Context, keyword string) ([]*Item, error) {
	// LIKE検索で部分一致するものを探す
	query := `
        SELECT i.id, i.name, c.name AS category, i.image_name 
        FROM items i
        JOIN categories c ON i.category_id = c.id
        WHERE i.name LIKE ?
    `
	rows, err := i.db.QueryContext(ctx, query, "%"+keyword+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	return items, nil
}