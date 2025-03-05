package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")
var errItemNotFound = errors.New("item not found")

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
	GetItemById(ctx context.Context, item_id string) (Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	// fileName is the path to the JSON file storing items.
	fileName string
}

// NewItemRepository creates a new itemRepository.
// main.goを実行するディレクトリによってfileNameを変更する
func NewItemRepository() ItemRepository {
	return &itemRepository{fileName: "cmd/api/items.json"}
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// STEP 4-2: add an implementation to store an item
	// /api内で go run main.go(パスを"items.json"でやるなら)

	// jsonファイルを読み込み
	jsonFile, err := os.ReadFile(i.fileName)
	// error処理
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// ファイルがなかったらnilを返す
			println("file does not exist.")
			return nil
		} else {
			// その他のエラーだったら中断、エラーを返す
			println("Error:", err)
			return err
		}
	}
	if len(jsonFile) == 0 {
		jsonFile = []byte(`{"items":[]}`)
	}

	// jsonから構造体に変換
	// jsonのitems配列の各要素がItem構造体として格納される
	var data struct {
		Items []Item `json:"items"`
	}
	if err = json.Unmarshal(jsonFile, &data); err != nil {
		return err
	}

	// itemsにitemを追加
	data.Items = append(data.Items, *item)

	// 構造体からjsonに変換
	itemsJson, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// jsonファイルに書き込み
	// 0644 ファイルのパーミッション
	err = os.WriteFile(i.fileName, itemsJson, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetAll()
func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
	// jsonファイルを読み込み
	jsonFile, err := os.ReadFile(i.fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// ファイルが存在しなかったら空のスライスとnilを返す
			println("file does not exist.")
			return []Item{}, nil
		} else {
			// その他のエラーだったら中断、空のスライスとエラーを返す
			println("Error:", err)
			return []Item{}, err
		}
	}

	// jsonから構造体に変換
	var data struct {
		Items []Item `json:"items"`
	}
	if err = json.Unmarshal(jsonFile, &data); err != nil {
		return nil, err
	}

	return data.Items, nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	// 保存先
	savePath := filepath.Join("images", fileName)

	// バックスラッシュをスラッシュに
	savePath = filepath.ToSlash(savePath)
	// ファイルを保存
	err := os.WriteFile(savePath, image, 0644)
	if err != nil {
		return err
	}

	return nil

}

// GetItemById()
func (i *itemRepository) GetItemById(ctx context.Context, item_id string) (Item, error) {
	// jsonファイルを読み込み
	jsonFile, err := os.ReadFile(i.fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			println("file does not exist.")
			return Item{}, nil
		} else {
			println("Error:", err)
			return Item{}, err
		}

	}
	// jsonから構造体に変換
	var data struct {
		Items []Item `json:"items"`
	}
	if err = json.Unmarshal(jsonFile, &data); err != nil {
		return Item{}, err
	}

	itemIdInt, err := strconv.Atoi(item_id)
	if err != nil {
		return Item{}, errors.New("invalid item id")
	}

	// 範囲外参照を確認
	idx := itemIdInt - 1
	if idx < 0 || idx >= len(data.Items) {
		return Item{}, errors.New("item not found: index out of range")
	}

	item := data.Items[idx]
	if item == (Item{}) {
		return Item{}, errItemNotFound
	}
	return item, nil
}

/*

*** STEP 4 ***
GETとPOSTのリクエストの違いについて調べてみましょう
	->GET:  サーバーにリクエストを送信、リソースを取得
	->POST: サーバーにデータを送信、リソースの更新など

ブラウザで http://127.0.0.1:9000/items にアクセスしても {"message": "item received: <name>"} が返ってこないのはなぜでしょうか？
	-> server.go の route に GET /items がないから？

アクセスしたときに返ってくるHTTPステータスコードはいくつですか？
	-> 200 OK

それはどんな意味をもつステータスコードですか？
	-> リクエストが正常に処理された

ハッシュ化とはなにか？
	-> 特定のルール(ハッシュ関数)に基づいて値を変換すること

SHA-256 以外にどんなハッシュ関数があるか調べてみましょう
	-> SHA-3, MD5など >アルゴリズムの設計、セキュリティ強度、速度、用途が違う らしい

Log levelとは？
	-> ソフトウェアが記録するログ(どんな動作が行われたかの記録)の詳細度と重要度を調整するための仕組み

webサーバーでは、本番はどのログレベルまで表示する？
	-> INFO以上が一般的 開発環境だとDEBUG

port (ポート番号)
	-> コンピュータが通信に使用するプログラムを識別するための番号 HTTP:80 etc.

localhost, 127.0.0.1
	-> localhost: コンピューター自身を指し示すためのホスト名
	-> 127.0.0.1: IPv4における特別なIPアドレス

HTTPリクエストメソッド
	-> Webサーバーにどのような処理をするかを伝える役割
	-> GET/POST/PUT(更新)/PATCH(一部更新)/DELETE(削除)


*/

// curlじゃなくて curl.exe で実行
// cd go してから go run cmd/api/main.go でサーバーを起動するなら
// main.go の実行ディレクトリは go/
