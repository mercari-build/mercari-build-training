package app

import (
	"crypto/sha256"
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
// サーバーを立ち上げる：Run関数で指定
func (s Server) Run() int {
	// set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	// STEP 4-6: set the log level to DEBUG
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// set up CORS settings
	frontURL, found := os.LookupEnv("FRONT_URL")
	if !found {
		frontURL = "http://localhost:3000"
	}

	// STEP 5-1: set up the database connection

	// set up handlers
	itemRepo := NewItemRepository()
	h := &Handlers{imgDirPath: s.ImageDirPath, itemRepo: itemRepo}

	// set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.Hello)  // GET /が呼ばれたらHelloを呼び出す
	mux.HandleFunc("GET /items", h.GetItems)  // 一覧を返すエンドポイント
	mux.HandleFunc("POST /items", h.AddItem)  // POST /itemsが呼ばれたらAddItemを呼び出す
	mux.HandleFunc("GET /items/{id}", h.GetItem)  // 商品を取得する(パスに含まれるデータを取得するにはこの形がいい)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)

	// start the server
	// サーバーを立てる
	slog.Info("http server started on", "port", s.Port)
	err := http.ListenAndServe(":"+s.Port, simpleCORSMiddleware(simpleLoggerMiddleware(mux), frontURL, []string{"GET", "HEAD", "POST", "OPTIONS"}))
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
	resp := HelloResponse{Message: "Hello, world!"}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type GetItemsResponse struct{
	Items []*Item `json:"items"`
}

func (s *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	items, err := s.itemRepo.List(ctx)  //リポジトリを見て商品をすべて取得する
	if err != nil {
		slog.Error("failed to get items: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := GetItemsResponse{Items:items}  //リポジトリからかえってきたレスポンスを形式の沿って帰す
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// 
func (s *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// パスパラメーターからIDを引っ張ってくる
	sid := r.PathValue("id")
	if sid == "" {
		http.Error(w,"id is required", http.StatusBadRequest)
		return
	}
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
		slog.Error("failed to get item: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// selectした商品を返す
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// AddItemRequestは以下の情報を受け取れる
type AddItemRequest struct {
	Name 	 string `form:"name"`
	Category string `form:"category"` // STEP 4-2: add a category field
	Image 	 []byte `form:"image"` // STEP 4-4: add an image field  受け取った画像ファイルを構造体にそのまま載せる
}

type AddItemResponse struct {
	Message string `json:"message"`
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

	// validate the request
	if req.Name == "" {
		return nil, errors.New("name is required") //Nameの中身が空だったらエラーを返す
	}

	// STEP 4-2: validate the category field
	if req.Category == "" {
		return nil, errors.New("category is required") //categoryの中身が空だったらエラーを返す
	}
	// STEP 4-4: validate the image field
	if len(req.Image) == 0 {
		return nil, errors.New("uploaded image file is empty")
	} 
	return req, nil
}

// AddItem is a handler to add a new item for POST /items .
// 直接乗せた画像ファイルを変更
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
		slog.Error("failed to store image: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item := &Item{
		Name: req.Name,
		// STEP 4-2: add a category field
		Category: req.Category,
		// STEP 4-4: add an image field
		ImageName: fileName,
	}
	message := fmt.Sprintf("item received: %s", item.Name)
	slog.Info(message)

	// STEP 4-2: add an implementation to store an image
	// 受け取ったリクエストをサーバーのリポジトリ(何かを保管する場所)に保存する
	err = s.itemRepo.Insert(ctx, item)
	if err != nil {
		slog.Error("failed to store item: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// AddItemResponseにメッセージを入れて返す
	resp := AddItemResponse{Message: message}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// storeImage stores an image and returns the file path and an error if any.
// this method calculates the hash sum of the image as a file name to avoid the duplication of a same file
// and stores it in the image directory.
func (s *Handlers) storeImage(image []byte) (filePath string, err error) {
	// STEP 4-4: add an implementation to store an image

	// TODO:
	// - calc hash sum
	// sha256でハッシュの文字列にする
	hash := sha256.Sum256(image)
	hashStr := hex.EncodeToString(hash[:])

	// - build image file path
	// ハッシュの文字列からファイルパスを作る
	fileName := fmt.Sprintf("%s.jpg", hashStr)
	filePath = filepath.Join(s.imgDirPath, fileName)

	// - check if the image already exists
	// 画像がすでにある場合のハンドリング
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("error checking image existence: %w", err)
	}
	// - store image
	// 画像の保存
	if err := StoreImage(filePath, image); err != nil {
		return "", fmt.Errorf("failed to store image: %w", err)
	}
	// - return the image file path
	// ファイル名を返す
	return fileName, nil

	return
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
