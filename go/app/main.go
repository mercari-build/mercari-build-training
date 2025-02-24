package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	imageDirPath     = "images"
	itemJSONFilePath = "items.json"
	port             = "9000"
)

var errImageNotFound = errors.New("image not found")

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
	Name string `form:"name"`
}

type AddItemResponse struct {
	Message string `json:"message"`
}

// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	req := &AddItemRequest{
		Name: r.FormValue("name"),
	}

	// validate the request
	if req.Name == "" {
		return nil, errors.New("name is required")
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

	// STEP 3-4: add an implementation to store an image

	item := &Item{Name: req.Name}
	message := fmt.Sprintf("item received: %#v", item)
	slog.InfoContext(ctx, message)

	// STEP 3-2: add an implementation to store an image
	err = s.itemRepo.Insert(ctx, item)
	if err != nil {
		slog.ErrorContext(ctx, "failed to store item: ", "error", err)
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
func (s *Handlers) GetImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseGetImageRequest(r)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse get image request: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	imgPath, err := s.buildImagePath(req.FileName)
	if err != nil {
		// when the image is not found, it returns the default image.
		if errors.Is(err, errImageNotFound) {
			slog.DebugContext(ctx, "image not found", "filename", imgPath)
			imgPath = filepath.Join(s.imgDirPath, "default.jpg")
		} else {
			slog.WarnContext(ctx, "failed to build image path: ", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	slog.InfoContext(ctx, "returned image", "path", imgPath)
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

	// validate the image path
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

type Item struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE

// ItemRepository is an interface to manage items.
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
}

// itemRepositoryJSON is an implementation of ItemRepository using JSON files.
type itemRepositoryJSON struct {
	// fileName is the path to the JSON file storing items.
	fileName string
}

// NewItemRepositoryJSON creates a new itemRepositoryJSON.
func NewItemRepositoryJSON(fileName string) ItemRepository {
	return &itemRepositoryJSON{fileName: fileName}
}

// Insert inserts an item into the JSON file.
func (i *itemRepositoryJSON) Insert(ctx context.Context, item *Item) error {
	// STEP 3-2: add an implementation to store an image

	return nil
}

func main() {
	// set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	// STEP 3-6: set the log level to info
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// set up CORS settings
	frontURL, found := os.LookupEnv("FRONT_URL")
	if !found {
		frontURL = "http://localhost:3000"
	}

	// set up handlers
	itemRepo := NewItemRepositoryJSON(itemJSONFilePath)
	h := &Handlers{imgDirPath: imageDirPath, itemRepo: itemRepo}

	// set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.Hello)
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)

	// start the server
	slog.Info("http server started on", "port", port)
	err := http.ListenAndServe(":"+port, simpleCORSMiddleware(simpleLoggerMiddleware(mux), frontURL, []string{"GET", "HEAD", "POST", "OPTIONS"}))
	if err != nil {
		slog.Error("failed to start server: ", "error", err)
	}
}
