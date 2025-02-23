package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	imageDirPath = "images"
	port         = "9000"
)

type Handlers struct {
	// imgDirPath is the path to the directory storing images.
	imgDirPath string
	itemRepo   ItemRepository
}

type HelloResponse struct {
	Message string `json:"message"`
}

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
}

type AddItemResponse struct {
	Message string `json:"message"`
}

// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	req := &AddItemRequest{
		Name:     r.FormValue("name"),
		Category: r.FormValue("category"),
	}

	// validate the request
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Category == "" {
		return nil, errors.New("category is required")
	}

	return req, nil
}

type Item struct {
	Name     string `db:"name"`
	Category string `db:"category"`
}

// //go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE}_mock -destination=./mock/$GOFILE
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
}

type itemRepository struct {
	db *sql.DB
}

func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	const sql = "INSERT INTO items (name, category) VALUES (?, ?)"

	stmt, err := i.db.PrepareContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(item.Name, item.Category)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseAddItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message := fmt.Sprintf("item received: %#v", *req)
	slog.InfoContext(ctx, message)

	// TODO: add implementation to store the item
	err = s.itemRepo.Insert(ctx, &Item{Name: req.Name, Category: req.Category})
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
	FileName string `form:"filename"`
}

// parseGetImageRequest parses and validates the request to get an image.
func parseGetImageRequest(r *http.Request) (*GetImageRequest, error) {
	req := &GetImageRequest{
		FileName: r.PathValue("filename"),
	}

	// validate the request
	if req.FileName == "" {
		return nil, errors.New("filename is required")
	}

	return req, nil
}

func (s *Handlers) GetImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseGetImageRequest(r)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse get image request: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// when the image is not found,
	imgPath, err := s.buildImagePath(req.FileName)
	if err != nil {
		slog.WarnContext(ctx, "failed to build image path: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
		return "", fmt.Errorf("image not found: %s", imgPath)
	}

	return imgPath, nil
}

func simpleCORSMiddleware(next http.Handler, origin string, methods []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelInfo)

	mux := http.NewServeMux()

	// set up CORS settings
	frontURL, found := os.LookupEnv("FRONT_URL")
	if !found {
		frontURL = "http://localhost:3000"
	}

	dbPath, found := os.LookupEnv("DB_PATH")
	if !found {
		f, err := os.Create("mercari.sqlite3")
		if err != nil {
			slog.Error("failed to create db file: ", "error", err)
			os.Exit(1)
		}
		defer f.Close()
		dbPath = "mercari.sqlite3"
	}
	// confirm existence
	_, err := os.Stat(dbPath)
	if err != nil {
		slog.Error("failed to get db file info: ", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		slog.Error("failed to open db: ", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// TODO: replace it with real SQL file.
	cmd := `CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
    	name VARCHAR(255),
    	category VARCHAR(255)
	)`
	_, err = db.ExecContext(context.Background(), cmd)
	if err != nil {
		slog.Error("failed to create table: ", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	h := &Handlers{imgDirPath: imageDirPath, itemRepo: &itemRepository{db: db}}
	mux.HandleFunc("GET /hello", h.Hello)
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)

	slog.Info("starting server", "port", port)
	err = http.ListenAndServe(":"+port, simpleCORSMiddleware(mux, frontURL, []string{"GET", "HEAD", "POST", "OPTIONS"}))
	if err != nil {
		slog.Error("failed to start server: ", "error", err)
	}
}
