package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Server struct {
	// Port is the port number to listen on.
	Port string
	// ImageDirPath is the path to the directory storing images.
	ImageDirPath string
}

// Run is a method to start the server.
// This method returns 0 if the server started successfully, and 1 otherwise.
func (s Server) Run() int {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelInfo)

	dbPath := "db/mercari.sqlite3"
	itemRepo, err := NewItemRepository(dbPath)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return 1 // エラーが発生したら終了
	}

	h := &Handlers{imgDirPath: s.ImageDirPath, itemRepo: itemRepo}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /items", h.GetItems)
	mux.HandleFunc("GET /items/{item_id}", h.GetItemByID)
	mux.HandleFunc("GET /search", h.SearchItems)

	slog.Info("http server started on", "port", s.Port)
	err = http.ListenAndServe(":"+s.Port, mux)
	if err != nil {
		slog.Error("failed to start server", "error", err)
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

func (s *Handlers) GetItemByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// URL パスから item_id を取得
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}
	itemIDStr := parts[2] // `/items/{item_id}` の `{item_id}` に相当

	// 数値に変換
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil || itemID < 1 {
		http.Error(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	// アイテムを取得
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	// JSON でレスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Handlers) SearchItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// クエリパラメータから `keyword` を取得
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "keyword is required", http.StatusBadRequest)
		return
	}

	// `keyword` を含む商品を検索
	items, err := s.itemRepo.SearchByName(ctx, keyword)
	if err != nil {
		http.Error(w, "failed to search items", http.StatusInternalServerError)
		return
	}

	// JSON でレスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetItemsResponse{Items: items})
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
	Category string `form:"category"`
	Image    []byte `form:"image"` // STEP 4-4: add an image field
}

type AddItemResponse struct {
	Message string `json:"message"`
}

type GetItemsResponse struct {
	Items []Item `json:"items"`
}

// parseAddItemRequest parses and validates the request to add an item.
// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	err := r.ParseMultipartForm(10 << 20) // 最大 10MB のファイルを許可
	if err != nil {
		return nil, errors.New("failed to parse multipart form")
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		return nil, errors.New("image file is required")
	}
	defer file.Close()

	imageData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New("failed to read image file")
	}

	req := &AddItemRequest{
		Name:     r.FormValue("name"),
		Category: r.FormValue("category"),
		Image:    imageData,
	}

	// バリデーション
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Category == "" {
		return nil, errors.New("category is required")
	}
	if len(req.Image) == 0 {
		return nil, errors.New("image is required")
	}

	return req, nil
}

func (s *Handlers) hashAndStoreImage(filePath string) (string, error) {
	// 画像データを読み込む
	imageData, err := ioutil.ReadFile(filePath)
	if err != nil {
		slog.Error("failed to read image", "error", err)
		return "", err
	}

	// SHA-256 ハッシュを計算
	hash := sha256.Sum256(imageData)
	hashedFileName := hex.EncodeToString(hash[:]) + ".jpg"
	newFilePath := filepath.Join(s.imgDirPath, hashedFileName)

	// 既に存在する場合は保存せず、そのままファイル名を返す
	if _, err := os.Stat(newFilePath); err == nil {
		slog.Info("image already exists", "path", newFilePath)
		return hashedFileName, nil
	}

	// 画像を保存
	err = ioutil.WriteFile(newFilePath, imageData, 0644)
	if err != nil {
		slog.Error("failed to store image", "error", err)
		return "", err
	}

	slog.Info("image stored", "path", newFilePath, "hashedFileName", hashedFileName)
	return hashedFileName, nil
}

// AddItem is a handler to add a new item for POST /items .
// AddItem is a handler to add a new item for POST /items .
// func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()

// 	// `default.jpg` のパス
// 	defaultImagePath := filepath.Join(s.imgDirPath, "default.jpg")

// 	// 画像のハッシュ化 & 保存
// 	fileName, err := s.hashAndStoreImage(defaultImagePath)
// 	if err != nil {
// 		slog.Error("failed to store image: ", "error", err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	slog.Info("Hashed image file name", "fileName", fileName)

// 	// アイテムの追加
// 	item := &Item{
// 		Name:     "Default Item",
// 		Category: "Default Category",
// 		Image:    fileName, // ハッシュ化された画像のファイル名
// 	}

// 	message := fmt.Sprintf("item received: %s", item.Name)
// 	slog.Info("Saving item", "item", item)
// 	err = s.itemRepo.Insert(ctx, item)
// 	if err != nil {
// 		slog.Error("failed to store item: ", "error", err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	resp := AddItemResponse{Message: message}
// 	err = json.NewEncoder(w).Encode(resp)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseAddItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 画像の保存処理
	fileName, err := s.storeImage(req.Image)
	if err != nil {
		http.Error(w, "failed to store image", http.StatusInternalServerError)
		return
	}

	// データベースにアイテムを追加
	item := &Item{
		Name:     req.Name,
		Category: req.Category, // ここでカテゴリIDに変換
		Image:    fileName,
	}

	err = s.itemRepo.Insert(ctx, item)
	if err != nil {
		http.Error(w, "failed to save item", http.StatusInternalServerError)
		return
	}

	// 成功レスポンス
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "item added"})
}

// GetItems メソッド（アイテム一覧を取得する処理）
// func (s *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()

// 	// アイテムを取得
// 	items, err := s.itemRepo.GetAll(ctx)
// 	if err != nil {
// 		slog.Error("failed to retrieve items: ", "error", err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// JSON で返す
// 	resp := GetItemsResponse{Items: items}
// 	err = json.NewEncoder(w).Encode(resp)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

func (s *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	items, err := s.itemRepo.GetAll(ctx)
	if err != nil {
		http.Error(w, "failed to retrieve items", http.StatusInternalServerError)
		return
	}

	// JSON形式でデータを返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetItemsResponse{Items: items})
}

// storeImage stores an image and returns the file path and an error if any.
// this method calculates the hash sum of the image as a file name to avoid the duplication of a same file
// and stores it in the image directory.
func (s *Handlers) storeImage(image []byte) (string, error) {
	// SHA-256 ハッシュを計算
	hash := sha256.Sum256(image)
	hashedFileName := hex.EncodeToString(hash[:]) + ".jpg"
	filePath := filepath.Join(s.imgDirPath, hashedFileName)

	// 画像が既に存在するか確認
	if _, err := os.Stat(filePath); err == nil {
		slog.Info("image already exists", "path", filePath)
		return hashedFileName, nil
	}

	// 画像を保存
	err := ioutil.WriteFile(filePath, image, 0644)
	if err != nil {
		slog.Error("failed to store image", "error", err)
		return "", err
	}

	slog.Info("image stored", "path", filePath)
	return hashedFileName, nil
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

// GetImage is a handler to return an image for GET /images/{filename} .
// If the specified image is not found, it returns the default image.
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

		// when the image is not found, it returns the default image without an error.
		slog.Debug("image not found", "filename", imgPath)
		imgPath = filepath.Join(s.imgDirPath, "default.jpg")
	}

	slog.Info("returned image", "path", imgPath)
	http.ServeFile(w, r, imgPath)
}

// buildImagePath builds the image path and validates it.
func (s *Handlers) buildImagePath(imageFileName string) (string, error) {
	imgPath := filepath.Join(s.imgDirPath, filepath.Clean(imageFileName))

	// to prevent directory traversal attacks
	rel, err := filepath.Rel(s.imgDirPath, imgPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid image path: %s", imgPath)
	}

	// validate the image suffix
	if !strings.HasSuffix(imgPath, ".jpg") && !strings.HasSuffix(imgPath, ".jpeg") {
		return "", fmt.Errorf("image path does not end with .jpg or .jpeg: %s", imgPath)
	}

	// check if the image exists
	_, err = os.Stat(imgPath)
	if err != nil {
		return imgPath, errImageNotFound
	}

	return imgPath, nil
}
