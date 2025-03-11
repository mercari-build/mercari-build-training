package app

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	// Port is the port number to listen on.
	Port 		 string
	// ImageDirPath is the path to the directory storing images.
	ImageDirPath string
	DB           *sql.DB
}

// Run is a method to start the server.
// This method returns 0 if the server started successfully, and 1 otherwise.
// サーバーを立ち上げる：Run関数で指定
func (s Server) Run() int {
	// set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	// STEP 5-1: set up the database connection
	db, err := sql.Open("sqlite3", "db/mercari.sqlite3")
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return 1
	}
	defer db.Close()

	if err := SetupDatabase(db); err != nil {
		slog.Error("failed to setup database", "error", err)
		return 1
	}

	// set up handlers
	itemRepo := NewItemRepository(db)
	h := &Handlers{imgDirPath: s.ImageDirPath, itemRepo: itemRepo}

	// set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.Hello)  // GET /が呼ばれたらHelloを呼び出す
	mux.HandleFunc("GET /items", h.GetItems)  // 一覧を返すエンドポイント
	mux.HandleFunc("POST /items", h.AddItem)  // POST /itemsが呼ばれたらAddItemを呼び出す
	mux.HandleFunc("GET /items/{id}", h.GetItem)  // 商品を取得する(パスに含まれるデータを取得するにはこの形がいい)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)
	mux.HandleFunc("GET /search", h.SearchItems) // 検索エンドポイント


	// start the server
	// サーバーを立てる
	slog.Info("http server started on", "port", s.Port)
	err = http.ListenAndServe(":"+s.Port, mux)
	if err != nil {
		slog.Error("failed to start server: ", "error", err)
		return 1
	}

	return 0
}

type Handlers struct {
	// imgDirPath is the path to the directory storing images.
	imgDirPath string
	itemRepo   ItemRepository
}

type HelloResponse struct {
	Message string `json:"message"`
}

// Hello is a handler to return a Hello, world! message for GET / .
func (s *Handlers) Hello(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello, world!"}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// `items` テーブルと `categories` テーブルを `JOIN` してデータを取得
	items, err := h.itemRepo.List(ctx)
	if err != nil {
		http.Error(w, "failed to get items", http.StatusInternalServerError)
		return
	}

	// JSON レスポンスを返す
	resp := map[string]interface{}{"items": items}
	json.NewEncoder(w).Encode(resp)
}

func (s *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// パスパラメーターからIDを引っ張ってくる
	sid := r.PathValue("id")

	// 文字列を数値に変換する
	id, err := strconv.Atoi(sid)
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}

	// リポジトリからIdを使って商品をselectする
	// Listに対してselectを作る
	item, err := s.itemRepo.Select(ctx, id)
	if err != nil {
		if errors.Is(err, errItemNotFound) {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get item", http.StatusInternalServerError)
		return
	}

	// selectした商品を返す
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(item)
}

// AddItemRequestは以下の情報を受け取れる
type AddItemRequest struct {
	Name 	 string `form:"name"`
	Category string `form:"category"` // STEP 4-2: add a category field
	Image 	 []byte `form:"image"` // STEP 4-4: add an image field  受け取った画像ファイルを構造体にそのまま載せる
}

// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	req := &AddItemRequest{
		Name: r.FormValue("name"),
		// STEP 4-2: add a category field
		Category: r.FormValue("category"),
	}

	// STEP 4-4: add an image field
	// リクエストで受け取った画像がFormFile("image")に入る
	uploadedFile, _, err := r.FormFile("image")
	if err != nil {
		return nil, errors.New("image is required")
	}
	defer uploadedFile.Close()

	imageData, err := io.ReadAll(uploadedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	} 
	req.Image = imageData
	return req, nil
}

// AddItem is a handler to add a new item for POST /items .
// 直接乗せた画像ファイルを変更
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.Info("Received request to add item")

	req, err := parseAddItemRequest(r)  // リクエストが来た時にAddItemRequestにリクエストの中身を入れて返す
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) 
		return
	}

	// STEP 4-4: uncomment on adding an implementation to store an image
	// storeImageを呼び出すと画像ファイルを保存してファイル名を返す
	// Insertでまとめて画像も保存できるようにする
	fileName, err := s.storeImage(req.Image) //画像を保存する処理
	if err != nil {
		http.Error(w, "failed to store image: ", http.StatusInternalServerError)
		return
	}

	item := &Item{
		Name: req.Name,
		// STEP 4-2: add a category field
		Category: req.Category,
		// STEP 4-4: add an image field
		ImageName: fileName,
	}

	// STEP 4-2: add an implementation to store an image
	// 受け取ったリクエストをサーバーのリポジトリ(何かを保管する場所)に保存する
	// DBにデータを追加
	err = s.itemRepo.Insert(ctx, item)
	if err != nil {
		slog.Error("failed to store item", "error", err)
		http.Error(w, fmt.Sprintf("failed to store item: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	slog.Info("Item successfully stored", "id", item.ID)

	// JSONレスポンスを返す
	resp := map[string]interface{}{
		"id":      item.ID,
		"message": "item received: " + item.Name,
	}
	json.NewEncoder(w).Encode(resp)
}

// storeImage stores an image and returns the file path and an error if any.
// this method calculates the hash sum of the image as a file name to avoid the duplication of a same file
// and stores it in the image directory.
func (s *Handlers) storeImage(image []byte) (string, error) {
	// STEP 4-4: add an implementation to store an image

	// TODO:
	// - calc hash sum
	// sha256でハッシュの文字列にする
	hash := sha256.Sum256(image)
	hashStr := hex.EncodeToString(hash[:])

	// - build image file path
	// ハッシュの文字列からファイルパスを作る
	fileName := fmt.Sprintf("%s.jpg", hashStr)
	imgPath := filepath.Join(s.imgDirPath, fileName)

	// - check if the image already exists
	// 画像がすでにある場合のハンドリング
	if _, err := os.Stat(imgPath); err == nil {
		return fileName, nil
	}
	// - store image
	// 画像の保存
	if err := os.WriteFile(imgPath, image, 0644); err != nil {
		return "", fmt.Errorf("failed to store image: %w", err)
	}
	// - return the image file path
	// ファイル名を返す
	return fileName, nil
}

// GetImage is a handler to return an image for GET /images/{filename} .
// If the specified image is not found, it returns the default image.
func (s *Handlers) GetImage(w http.ResponseWriter, r *http.Request) {
	fileName := r.PathValue("filename")
	imgPath := filepath.Join(s.imgDirPath, fileName)

		// when the image is not found, it returns the default image without an error.
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		imgPath = filepath.Join(s.imgDirPath, "default.jpg")
	}

	http.ServeFile(w, r, imgPath)
}

// 検索用エンドポイントを追加
func (h *Handlers) SearchItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// クエリパラメータ "keyword" を取得
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "keyword is required", http.StatusBadRequest)
		return
	}

	// リポジトリで検索
	items, err := h.itemRepo.Search(ctx, keyword)
	if err != nil {
		http.Error(w, "failed to search items", http.StatusInternalServerError)
		return
	}

	// 結果を JSON で返す
	resp := map[string]interface{}{"items": items}
	json.NewEncoder(w).Encode(resp)
}
