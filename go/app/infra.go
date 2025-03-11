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
	Select(ctx context.Context, id int) (*Item, error)  // リポジトリのinterfaceにselectを追加
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


// 5-1 Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	query := `INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)`
	result, err := i.db.ExecContext(ctx, query, item.Name, item.Category, item.ImageName)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}
	item.ID = int(id) // ここで ID をセット
	return nil
}

// インターフェイス（関数名、引数、戻り値組み合わせ）と同じ関数名と引数と戻り値を指定する
func (i *itemRepository) List(ctx context.Context) ([]*Item, error) {
	rows, err := i.db.QueryContext(ctx, "SELECT id, name, category, image_name FROM items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := &Item{}
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// 5-1selectの実装
func (i *itemRepository) Select(ctx context.Context, id int) (*Item, error) {
	// idが1以上の値以外になる場合はidをNotFoundにする
	var item Item
	err := i.db.QueryRowContext(ctx, "SELECT id, name, category, image_name FROM items WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Category, &item.ImageName)
	if err == sql.ErrNoRows {
		return nil, errItemNotFound
	}
	if err != nil {
		slog.Error("failed to select item", "error", err)
		return nil, fmt.Errorf("failed to select item: %w", err)
	}
	return &item, err
}
