package app

import (
	"encoding/json" // JSON のエンコード/デコードを行うためのパッケージ
	"errors"        // エラーハンドリング用のパッケージ
	"fmt"           // フォーマット付き文字列を扱うためのパッケージ
	"log/slog"      // ログ出力用のパッケージ
	"net/http"      // HTTP サーバーを扱うためのパッケージ
	"os"            // 環境変数やファイル操作用のパッケージ
	"path/filepath" // ファイルパスを扱うためのパッケージ
	"strings"       // 文字列操作を行うためのパッケージ
)

// Server はサーバーの設定を管理する構造体___Server is a struct to manage server settings
type Server struct {
	Port        string // リッスンするポート番号___Port number to listen on
	ImageDirPath string // 画像を保存するディレクトリのパス___Path to the directory storing images
}

// Run はサーバーを起動するメソッド___Run is a method to start the server
// サーバーが正常に起動した場合は 0 を返し、エラーが発生した場合は 1 を返す___Returns 0 if the server started successfully, otherwise returns 1
func (s Server) Run() int {
	// ロガーの設定___Set up the logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	// ログレベルを DEBUG に設定___STEP 4-6: set the log level to DEBUG
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// CORS 設定のセットアップ___Set up CORS settings
	frontURL, found := os.LookupEnv("FRONT_URL")
	if !found {
		frontURL = "http://localhost:3000"
	}

	// データベース接続のセットアップ (未実装)___STEP 5-1: set up the database connection

	// ハンドラーのセットアップ___Set up handlers
	itemRepo := NewItemRepository()
	h := &Handlers{imgDirPath: s.ImageDirPath, itemRepo: itemRepo}

	// ルーティングの設定___Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.Hello)                 // ルートパスで Hello メッセージを返す
	mux.HandleFunc("POST /items", h.AddItem)        // 新しいアイテムを追加
	mux.HandleFunc("GET /images/{filename}", h.GetImage) // 画像を取得

	// サーバーの起動___Start the server
	slog.Info("http server started on", "port", s.Port)
	err := http.ListenAndServe(":"+s.Port, simpleCORSMiddleware(simpleLoggerMiddleware(mux), frontURL, []string{"GET", "HEAD", "POST", "OPTIONS"}))
	if err != nil {
		slog.Error("failed to start server: ", "error", err)
		return 1
	}

	return 0
}

// Handlers は各 HTTP ハンドラーを管理する構造体___Handlers struct to manage HTTP handlers
type Handlers struct {
	imgDirPath string       // 画像を保存するディレクトリのパス___Path to the directory storing images
	itemRepo   ItemRepository // アイテムを管理するリポジトリ___Repository to manage items
}

// HelloResponse は Hello メッセージのレスポンス___HelloResponse is a response struct for Hello message
type HelloResponse struct {
	Message string `json:"message"` // メッセージ本文___Response message
}

// Hello は "Hello, world!" を返すハンドラー___Hello is a handler to return a "Hello, world!" message
func (s *Handlers) Hello(w http.ResponseWriter, r *http.Request) {
	resp := HelloResponse{Message: "Hello, world!"}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// AddItemRequest はアイテム追加のリクエストデータを表す構造体___AddItemRequest represents the request data for adding an item
type AddItemRequest struct {
	Name     string `form:"name"`      // アイテム名___Item name
	Category string `form:"category"`  // カテゴリ___Category (STEP 4-2)
	Image    []byte `form:"image"`     // 画像データ___Image data (STEP 4-4)
}

// AddItemResponse はアイテム追加のレスポンスデータを表す構造体___AddItemResponse represents the response data for adding an item
type AddItemResponse struct {
	Message string `json:"message"` // メッセージ本文___Response message
}

// parseAddItemRequest はリクエストデータを解析し、検証する___parseAddItemRequest parses and validates the request to add an item
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	req := &AddItemRequest{
		Name:     r.FormValue("name"),
		Category: r.FormValue("category"),
	}

	// 必須フィールドの検証___Validate required fields
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Category == "" {
		return nil, errors.New("category is required")
	}

	return req, nil
}

// AddItem は新しいアイテムを追加するハンドラー___AddItem is a handler to add a new item
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseAddItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 画像を保存する (未実装)___STEP 4-4: store image
	fileName, err := s.storeImage(req.Image)
	if err != nil {
		slog.Error("failed to store image: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item := &Item{
		Name:     req.Name,
		Category: req.Category,
	}

	message := fmt.Sprintf("item received: %s", item.Name)
	slog.Info(message)

	// アイテムをデータベースに保存___Store item in database
	err = s.itemRepo.Insert(ctx, item)
	if err != nil {
		slog.Error("failed to store item: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := AddItemResponse{Message: message}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// storeImage は画像を保存し、ファイルのパスを返す___storeImage stores an image and returns the file path
func (s *Handlers) storeImage(image []byte) (filePath string, err error) {
	// TODO:
	// - 画像のハッシュを計算し、ファイル名を決定する___Calculate hash sum of image and determine file name
	// - 画像ファイルのパスを構築する___Build image file path
	// - 既に同じ画像があるかチェックする___Check if the image already exists
	// - 画像を保存する___Store image
	// - 画像のファイルパスを返す___Return the image file path

	return
}

// GetImage は画像を返すハンドラー___GetImage is a handler to return an image
func (s *Handlers) GetImage(w http.ResponseWriter, r *http.Request) {
	req, err := parseGetImageRequest(r)
	if err != nil {
		slog.Warn("failed to parse get image request: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	imgPath, err := s.buildImagePath(req.FileName)
	if err != nil {
		if !errors.Is(err, errImageNotFound) {
			slog.Warn("failed to build image path: ", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		slog.Debug("image not found", "filename", imgPath)
		imgPath = filepath.Join(s.imgDirPath, "default.jpg")
	}

	slog.Info("returned image", "path", imgPath)
	http.ServeFile(w, r, imgPath)
}
