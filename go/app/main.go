package main

import (
	"context"
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
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	imageDirPath = "images"
	port         = "9001"
)

type Handlers struct {
	// imgDirPath is the path to the directory storing images.
	imgDirPath string
	itemRepo   ItemRepository
}

type HelloResponse struct {
	Message string `json:"message"`
}

// Hello is an endpoint to return a Hello, world! message.
// GET /
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
	Image    []byte `form:"image"`
}

type AddItemResponse struct {
	Message string `json:"message"`
}

// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	f, _, err := r.FormFile("image")
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	req := &AddItemRequest{
		Name:     r.FormValue("name"),
		Category: r.FormValue("category"),
		Image:    buf,
	}

	// validate the request
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

// GetItems returns a list of items.
// GET /items
func (s *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	items, err := s.itemRepo.GetItems(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get items: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "items returned", "items", items)
}

// AddItem is an endpoint to add a new item.
// POST /items
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseAddItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileName, err := storeImage(req.Image)
	if err != nil {
		slog.ErrorContext(ctx, "failed to store image: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item := &Item{Name: req.Name, Category: req.Category, ImageName: fileName}
	message := fmt.Sprintf("item received: %#v", item)
	slog.InfoContext(ctx, message)

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

// storeImage stores the image and returns the file name.
// Although this function should be implemented with repository,
// it's implemented with a simple function for simplicity.
func storeImage(image []byte) (string, error) {
	hashSum := sha256.Sum256(image)
	hashSumStr := hex.EncodeToString(hashSum[:])
	fileName := hashSumStr + ".jpg"
	imgPath := filepath.Join(imageDirPath, fileName)

	err := os.WriteFile(imgPath, image, 0644)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

type GetItemRequest struct {
	ID int // path value
}

func parseGetItemRequest(r *http.Request) (*GetItemRequest, error) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse id: %w", err)
	}
	if id < 0 {
		return nil, errors.New("id must be greater than or equal to 0")
	}

	return &GetItemRequest{ID: id}, nil
}

type GetItemResponse struct {
	*Item
}

func (s *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseGetItemRequest(r)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse get item request: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, err := s.itemRepo.GetItem(ctx, req.ID)
	if err != nil {
		slog.WarnContext(ctx, "failed to get item: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := GetItemResponse{Item: item}
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
		FileName: r.PathValue("filename"),
	}

	// validate the request
	if req.FileName == "" {
		return nil, errors.New("filename is required")
	}

	return req, nil
}

// GetImage is an endpoint to return an image.
// GET /images/{filename}
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

type Item struct {
	ID        int    `db:"id"`
	Name      string `db:"name"`
	Category  string `db:"category"`
	ImageName string `db:"image_name"`
}

// //go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE}_mock -destination=./mock/$GOFILE
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetItems(ctx context.Context) ([]*Item, error)
	GetItem(ctx context.Context, id int) (*Item, error)
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

func (i *itemRepository) GetItems(ctx context.Context) ([]*Item, error) {
	const sql = "SELECT id, name, category, image_name FROM items"

	rows, err := i.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		err = rows.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		items = append(items, &item)
	}

	return items, nil
}

func (i *itemRepository) GetItem(ctx context.Context, id int) (*Item, error) {
	const sql = "SELECT id, name, category, image_name FROM items WHERE id = ?"

	rows, err := i.db.QueryContext(ctx, sql, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var item Item
	for rows.Next() {
		err = rows.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
	}

	return &item, nil
}

type itemRepositoryJson struct {
	fileName string
}

func NewItemRepositoryJson(fileName string) ItemRepository {
	return &itemRepositoryJson{fileName: fileName}
}

func (i *itemRepositoryJson) Insert(ctx context.Context, item *Item) error {
	buf, err := os.ReadFile(i.fileName)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var items []*Item
	err = json.Unmarshal(buf, &items)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	item.ID = len(items)
	items = append(items, item)

	buf, err = json.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	err = os.WriteFile(i.fileName, buf, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (i *itemRepositoryJson) GetItems(ctx context.Context) ([]*Item, error) {
	buf, err := os.ReadFile(i.fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var items []*Item
	err = json.Unmarshal(buf, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return items, nil
}

func (i *itemRepositoryJson) GetItem(ctx context.Context, id int) (*Item, error) {
	items, err := i.GetItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}

	return nil, fmt.Errorf("item not found: id=%d", id)
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

func simpleLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request received", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr, "user_agent", r.UserAgent())
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

	// db, err := sql.Open("sqlite3", dbPath)
	// if err != nil {
	// 	slog.Error("failed to open db: ", "error", err)
	// 	os.Exit(1)
	// }
	// defer db.Close()

	// // TODO: replace it with real SQL file.
	// cmd := `CREATE TABLE IF NOT EXISTS items (
	// 	id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 	name VARCHAR(255),
	// 	category VARCHAR(255)
	// )`
	// _, err = db.ExecContext(context.Background(), cmd)
	// if err != nil {
	// 	slog.Error("failed to create table: ", "error", err)
	// 	os.Exit(1)
	// }
	// defer db.Close()

	itemRepo := NewItemRepositoryJson("items.json")
	h := &Handlers{imgDirPath: imageDirPath, itemRepo: itemRepo}
	mux.HandleFunc("GET /", h.Hello)
	mux.HandleFunc("GET /items", h.GetItems)
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /items/{id}", h.GetItem)

	mux.HandleFunc("GET /images/{filename}", h.GetImage)

	slog.Info("http server started on", "port", port)
	err = http.ListenAndServe(":"+port, simpleCORSMiddleware(simpleLoggerMiddleware(mux), frontURL, []string{"GET", "HEAD", "POST", "OPTIONS"}))
	if err != nil {
		slog.Error("failed to start server: ", "error", err)
	}
}
