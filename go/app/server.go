package app

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Server 構造体: サーバー設定の管理
type Server struct {
	// Port is the port number to listen on.
	Port string
	// ImageDirPath is the path to the directory storing images.
	ImageDirPath string
	DB           *sql.DB
}

// Run メソッド: サーバーの起動処理
// Run is a method to start the server.
// This method returns 0 if the server started successfully, and 1 otherwise.
func (s Server) Run() int {
	// set up logger
	// DEBUG 以上のログを出力するよう設定
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// set up CORS settings
	frontURL, found := os.LookupEnv("FRONT_URL")
	if !found {
		frontURL = "http://localhost:3000"
	}

	// set up the database connection
	db, err := InitDB()
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		return 1
	}

	// set up handlers
	itemRepo := NewItemRepository(db)
	h := &Handlers{imgDirPath: s.ImageDirPath, itemRepo: itemRepo}

	// set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /items", h.GetItems)
	mux.HandleFunc("GET /items/{id}", h.GetItemByID)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)
	mux.HandleFunc("GET /search", h.SearchItems)
	mux.HandleFunc("GET /", h.Hello)

	// start the server
	slog.Info("http server started on", "port", s.Port)
	err = http.ListenAndServe(":"+s.Port, simpleCORSMiddleware(simpleLoggerMiddleware(mux), frontURL, []string{"GET", "HEAD", "POST", "OPTIONS"}))
	if err != nil {
		slog.Error("failed to start server", "error", err)
		return 1
	}

	return 0
}

// Handlers 構造体: リクエストの処理
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
	resp := HelloResponse{Message: "Hello, world!"}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type AddItemRequest struct {
	Name     string `form:"name"`
	Category string `form:"category"` // STEP 4-2: add a category field
	Image    []byte `form:"image"`    // STEP 4-4: add an image field
}

type AddItemResponse struct {
	Message string `json:"message"`
}

// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	// Parse multipart form data with 10MB max memory
	err := r.ParseMultipartForm(10 << 20) // 10MB max memory
	if err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	req := &AddItemRequest{
		Name:     r.FormValue("name"),
		Category: r.FormValue("category"),
	}

	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if len(req.Name) > 255 {
		return nil, errors.New("name is too long (max 255 chars)")
	}

	if req.Category == "" {
		return nil, errors.New("category is required")
	}
	if len(req.Category) > 255 {
		return nil, errors.New("category is too long (max 255 chars)")
	}

	// 画像ファイルの取得
	file, _, err := r.FormFile("image")
	if err != nil {
		return nil, errors.New("image is required")
	}
	defer file.Close()

	// 画像のバイナリデータを読み込む
	imageData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	// **空の画像データのチェック**
	if len(imageData) == 0 {
		return nil, errors.New("image file is empty")
	}

	// **MIMEタイプを `http.DetectContentType` で取得**
	contentType := http.DetectContentType(imageData[:512]) // 最初の 512 バイトから MIME を判定
	validMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	if !validMimeTypes[contentType] {
		return nil, fmt.Errorf("invalid image format (must be JPEG or PNG, got %s)", contentType)
	}

	req.Image = imageData
	return req, nil
}

// AddItem ハンドラー: アイテムを追加
// POST /items でアイテムを追加するハンドラー
// AddItem is a handler to add a new item for POST /items .
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// (1) リクエストデータの取得
	req, err := parseAddItemRequest(r)
	if err != nil {
		slog.Error("Request parsing error", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// (2) カテゴリの取得または作成
	category, err := s.itemRepo.GetCategoryByName(ctx, req.Category)
	if err != nil {
		// カテゴリが存在しない場合、新しく追加
		slog.Warn("Category not found, creating new category", "category", req.Category)
		category, err = s.itemRepo.InsertCategory(ctx, req.Category)
		if err != nil {
			slog.Error("Failed to create category", "category", req.Category, "error", err)
			http.Error(w, "Failed to create category", http.StatusInternalServerError)
			return
		}
	}

	// (3) 画像の保存
	fileName, err := s.storeImage(req.Image)
	if err != nil {
		slog.Error("failed to store image: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// (4) アイテムの保存
	item := &Item{
		Name:       req.Name,
		CategoryID: category.ID, // 取得または作成したカテゴリIDを使用
		ImageName:  fileName,
	}
	message := fmt.Sprintf("item received: %s", item.Name)
	slog.Info(message)

	err = s.itemRepo.Insert(ctx, item)
	if err != nil {
		slog.Error("failed to store item: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// (5) レスポンスを返す
	resp := AddItemResponse{Message: message}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetItems ハンドラー: 登録された商品の一覧を取得
func (s *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	items, err := s.itemRepo.FindAll(ctx)
	if err != nil {
		slog.Error("failed to retrieve items: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string][]Item{"items": items}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetItemByID ハンドラー: 商品詳細を取得する
func (s *Handlers) GetItemByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. リクエストデータの取得（IDの取得とパース）
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	// 文字列をintに変換
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		slog.Error("invalid id format", "error", err)
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	// 2. アイテムの取得
	item, err := s.itemRepo.FindByID(ctx, id)
	if err != nil {
		slog.Error("failed to retrieve item", "id", id, "error", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 3. レスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		slog.Error("failed to encode response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// storeImage stores an image and returns the file path and an error if any.
// this method calculates the hash sum of the image as a file name to avoid the duplication of a same file
// and stores it in the image directory.
func (s *Handlers) storeImage(image []byte) (filePath string, err error) {
	// Calculate SHA-256 hash of the image (ハッシュ化)
	hasher := sha256.New()
	hasher.Write(image)
	hashSum := hasher.Sum(nil)
	fileName := fmt.Sprintf("%x.jpg", hashSum)

	// Build the full file path
	filePath = filepath.Join(s.imgDirPath, fileName)

	// Check if the image already exists
	_, err = os.Stat(filePath)
	if err == nil {
		// Image already exists, return the filename
		return fileName, nil
	}

	// Store the image
	err = StoreImage(filePath, image)
	if err != nil {
		return "", fmt.Errorf("failed to store image: %w", err)
	}

	return fileName, nil
}

type GetImageRequest struct {
	FileName string // path value
}

// parseGetImageRequest parses and validates the request to get an image.
func parseGetImageRequest(r *http.Request) (*GetImageRequest, error) {
	req := &GetImageRequest{
		FileName: r.PathValue("filename"), // from path parameter
	}

	// validate the request
	if req.FileName == "" {
		return nil, errors.New("filename is required")
	}

	return req, nil
}

// GetImage ハンドラー: 画像を取得
// GetImage is a handler to return an image for GET /images/{filename} .
// If the specified image is not found, it returns the default image.
func (s *Handlers) GetImage(w http.ResponseWriter, r *http.Request) {
	req, err := parseGetImageRequest(r)
	if err != nil {
		slog.Warn("failed to parse get image request: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// (1) 画像のパスを取得
	imgPath, err := s.buildImagePath(req.FileName)
	if err != nil {
		if !errors.Is(err, errImageNotFound) {
			slog.Warn("failed to build image path: ", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// when the image is not found, it returns the default image without an error.
		slog.Debug("image not found", "filename", imgPath)
		imgPath = filepath.Join(s.imgDirPath, "default.jpg")
	}

	// (2) 画像をクライアントに返す
	slog.Info("returned image", "path", imgPath)
	http.ServeFile(w, r, imgPath)
}

// buildImagePath メソッド: 画像のパスを取得
// buildImagePath builds the image path and validates it.
func (s *Handlers) buildImagePath(imageFileName string) (string, error) {
	imgPath := filepath.Join(s.imgDirPath, filepath.Clean(imageFileName))

	// (1) パスの安全性をチェック
	// to prevent directory traversal attacks
	rel, err := filepath.Rel(s.imgDirPath, imgPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid image path: %s", imgPath)
	}

	// validate the image suffix
	if !strings.HasSuffix(imgPath, ".jpg") && !strings.HasSuffix(imgPath, ".jpeg") {
		return "", fmt.Errorf("image path does not end with .jpg or .jpeg: %s", imgPath)
	}

	// (2) 画像が存在するか確認
	// check if the image exists
	_, err = os.Stat(imgPath)
	if err != nil {
		return imgPath, errImageNotFound
	}

	return imgPath, nil
}

// SearchItems is a handler to search items by keyword for GET /search .
func (h *Handlers) SearchItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// クエリパラメータからキーワードを取得
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "keyword is required", http.StatusBadRequest)
		return
	}

	// アイテムを検索
	items, err := h.itemRepo.Search(ctx, keyword)
	if err != nil {
		slog.Error("failed to search items: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンスを返す
	response := map[string][]Item{"items": items}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetCategories ハンドラー: カテゴリー一覧を取得
func (s *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	categories, err := s.itemRepo.GetCategories(ctx)
	if err != nil {
		slog.Error("failed to retrieve categories: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string][]Category{"categories": categories}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
