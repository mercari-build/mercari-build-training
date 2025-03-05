package app

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	// "io/ioutil"
	// "io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	// set up logger
	// STEP 4-6: set the log level to DEBUG
	// levelInfo から levelDebugに変更??? できない
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &opts))
	slog.SetDefault(logger)

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
	mux.HandleFunc("GET /", h.Hello)
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /items", h.GetItems)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)
	mux.HandleFunc("GET /items/{item_id}", h.GetItemById)

	// start the server
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

// GetItems ハンドラーを実装 for GET /items
func (s *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	// GetAllメソッドを呼び出す
	items, err := s.itemRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// itemsをラップ
	response := struct {
		Items []Item `json:"items"`
	}{
		Items: items,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type AddItemRequest struct {
	Name     string `form:"name"`
	Category string `form:"category"` // STEP 4-2: add a category field -- done
	Image    []byte `form:"image"`    // STEP 4-4: add an image field -- done
}

type AddItemResponse struct {
	Message string `json:"message"`
}

// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	req := &AddItemRequest{
		Name: r.FormValue("name"),
		// STEP 4-2: add a category field -- done
		Category: r.FormValue("category"),
		// STEP 4-4: add an image field
		Image: []byte(r.FormValue("image")),
	}

	// validate the request
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	// STEP 4-2: validate the category field -- done
	if req.Category == "" {
		return nil, errors.New("category is required")
	}
	// STEP 4-4: validate the image field -- done
	if string(req.Image) == "" {
		return nil, errors.New("image is required")
	}

	// 拡張子を取得して、jpgのみ受け付ける
	fName := r.FormValue("image")
	ex := filepath.Ext(fName)
	if ex != ".jpg" {
		return nil, errors.New("invalid image extension. Only accept: jpg")
	}

	return req, nil
}

// AddItem is a handler to add a new item for POST /items .
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseAddItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// STEP 4-4: uncomment on adding an implementation to store an image -- done
	fileName, err := s.storeImage(req.Image)
	if err != nil {
		slog.Error("failed to store image: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item := &Item{
		Name: req.Name,
		// STEP 4-2: add a category field -- done
		Category: req.Category,
		// STEP 4-4: add an image field -- done
		// 先頭のimages/を削除 もっといい方法ありそう
		Image: strings.TrimPrefix(string(fileName), "images/"),
	}
	message := fmt.Sprintf("item received: %s", item.Name)
	slog.Info(message)

	// STEP 4-2: add an implementation to store an item -- done?
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

// storeImage stores an image and returns the file path and an error if any.
// this method calculates the hash sum of the image as a file name to avoid the duplication of a same file
// and stores it in the image directory.
func (s *Handlers) storeImage(image []byte) (filePath string, err error) {
	// STEP 4-4: add an implementation to store an image
	// TODO:
	// - calc hash sum
	hash := sha256.Sum256(image)
	// - build image file path
	// バックスラッシュをスラッシュに
	fileName := fmt.Sprintf("%x.jpg", hash)
	filePath = filepath.Join(s.imgDirPath, fileName)
	filePath = filepath.ToSlash(filePath)
	// - check if the image already exists
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}
	// - store image
	if err := os.WriteFile(filePath, image, 0644); err != nil {
		return "", fmt.Errorf("failed to write image file: %w", err)
	}
	// - return the image file path
	return filePath, nil
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

// GetItemById
type GetItemByIdRequest struct {
	Id string
}

func parseGetItemByIdRequest(r *http.Request) (*GetItemByIdRequest, error) {
	req := &GetItemByIdRequest{
		Id: r.PathValue("item_id"),
	}

	// validate the request
	if req.Id == "" {
		return nil, errors.New("id is required")
	}

	return req, nil
}

func (s *Handlers) GetItemById(w http.ResponseWriter, r *http.Request) {
	req, err := parseGetItemByIdRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	item, err := s.itemRepo.GetItemById(r.Context(), req.Id)
	// if &item == nil {
	// 	http.Error(w, err.Error(), http.StatusNotFound)
	// 	return
	// }
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// jsonに変換
	jsonData, err := json.Marshal(item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

/*
	そもそもハンドラーとは
	-> HTTPリクエストを受け取り、適切なレスポンスを返す関数

	step4-4
	# ローカルから.jpgをポストする
	$ curl \
	-X POST \
	--url 'http://localhost:9000/items' \
	-F 'name=jacket' \
	-F 'category=fashion' \
	-F 'image=@images/local_image.jpg' <-ローカルのuploadしたいiamgeのパス "image=go/images/default.jpg"とか

*/
