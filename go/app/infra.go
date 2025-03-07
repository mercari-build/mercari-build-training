package app

import (
	"context"      // コンテキストの管理を行うためのパッケージ（APIのリクエストやデータベース操作などで使用）
	"errors"       // エラーハンドリング用のパッケージ
	"fmt"          // 文字列フォーマットや標準出力を扱うためのパッケージ
	"os"           // ファイルやプロセスの操作を行うためのパッケージ
	"io"
	"encoding/json" // JSON データのエンコード/デコードを行うためのパッケージ
	// STEP 5-1: この行のコメントを解除してください
	// _ "github.com/mattn/go-sqlite3" // SQLite3 ドライバーのインポート（現在コメントアウト）
)

// 画像が見つからない場合のエラーを定義___Define an error when an image is not found
var errImageNotFound = errors.New("image not found")

// Item はアイテムを表す構造体___Item represents an item
type Item struct {
	ID       int    `db:"id" json:"-"`       // データベースの ID（JSON に含めない）
	Name     string `db:"name" json:"name"`  // データベースの name カラム（JSON に含める）
	Category string `db:"category" json:"category"` // データベースの category カラム（JSON に含める）
}

// `go generate ./...` を実行してモック実装を生成してください___Please run `go generate ./...` to generate the mock implementation
// ItemRepository はアイテムを管理するためのインターフェース___ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error  // アイテムをリポジトリに挿入するメソッド
	GetAllItems(ctx context.Context) ([]Item, error) // --------------------すべてのアイテムを取得するメソッド
}

// itemRepository は ItemRepository の実装___itemRepository is an implementation of ItemRepository
type itemRepository struct {
	fileName string // アイテムを保存する JSON ファイルのパス
}

// NewItemRepository は新しい itemRepository を作成___NewItemRepository creates a new itemRepository
func NewItemRepository() ItemRepository {
	return &itemRepository{fileName: "items.json"} // データを "items.json" に保存
}

// Insert はアイテムをリポジトリに挿入する___Insert inserts an item into the repository
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// ステップ 4-2: アイテムを保存するための実装を追加___STEP 4-2: add an implementation to store an item

	// 既存のアイテムを取得
	items, err := i.GetAllItems(ctx)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // ファイルが存在しない場合のエラーハンドリング
			return fmt.Errorf("データファイルが見つかりません: %w", err)
		}
		return err
	}

	// 新しいアイテムをリストに追加
	items = append(items, *item)

	// JSON ファイルを作成（既存ファイルを上書き）
	file, err := os.Create(i.fileName)
	if err != nil {
		return fmt.Errorf("データファイルの作成に失敗しました: %w", err)
	}
	defer file.Close() // ファイルをクローズ

	// JSON データをエンコードして保存
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ") // インデントを設定（可読性向上）
	if err := encoder.Encode(items); err != nil {
		return fmt.Errorf("データの保存に失敗しました: %w", err)
	}

	return nil
}

// GetAllItems はすべてのアイテムを取得する___GetAllItems retrieves all items from the repository
func (i *itemRepository) GetAllItems(ctx context.Context) ([]Item, error) {
	// ファイルを開く
	file, err := os.Open(i.fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // ファイルが存在しない場合
			return nil, fmt.Errorf("データファイルが見つかりません: %w", err)
		}
		return nil, err
	}
	defer file.Close() // ファイルをクローズ

	// JSON データをデコード
	var items []Item
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&items); err != nil {
		if errors.Is(err, io.EOF) { // ファイルが空の場合
			return []Item{}, nil
		}
		return nil, fmt.Errorf("データの読み込みに失敗しました: %w", err)
	}

	return items, nil
}

// StoreImage は画像を保存し、エラーがあれば返す___StoreImage stores an image and returns an error if any
// 簡潔にするため、このパッケージには関連するインターフェースがない___This package doesn't have a related interface for simplicity
func StoreImage(fileName string, image []byte) error {
	// ステップ 4-4: 画像を保存するための実装を追加___STEP 4-4: add an implementation to store an image

	return nil
}
