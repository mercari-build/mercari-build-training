package app

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex" //Hexadecimal
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
	"github.com/rs/cors"
)

type Server struct {
	Port         string // Port is the port number to listen on.
	ImageDirPath string // ImageDirPath is the path to the directory storing images.
	DB           *sql.DB
}

// Run is a method to start the server.
// This method returns 0 if the server started successfully, and 1 otherwise.
func (s Server) Run() int {
	// set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// set up CORS settings
	frontURL, found := os.LookupEnv("FRONT_URL")
	if !found {
		frontURL = "http://localhost:3000"
	}

	// STEP 5-1: set up the database connection
	db, err := sql.Open("sqlite3", "../../db/items.db")
	if err != nil {
		slog.Error("failed to open database: ", "error", err)
		return 1
	}
	s.DB = db

	// set up handlers
	h := &Handlers{imgDirPath: s.ImageDirPath, db: db}

	// set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)
	mux.HandleFunc("GET /items/{id}", h.GetItem)
	mux.HandleFunc("GET /items", h.GetAllItems)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{frontURL},
		AllowedMethods:   []string{"GET", "HEAD", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	slog.Info("http server started on", "port", s.Port)
	err = http.ListenAndServe(":"+s.Port, c.Handler(simpleLoggerMiddleware(mux)))
	if err != nil {
		slog.Error("failed to start server: ", "error", err)
		return 1
	}

	return 0
}

type Handlers struct {
	imgDirPath string  // imgDirPath is the path to the directory storing images.
	db         *sql.DB //define struct of db
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
	req := &AddItemRequest{
		Name:     r.FormValue("name"),
		Category: r.FormValue("category"),
	}

	// get the image file
	imageFile, _, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			req.Image = nil
		} else {
			return nil, errors.New("failed to get image file")
		}
	} else {
		defer imageFile.Close()
		imageData, err := io.ReadAll(imageFile)
		if err != nil {
			return nil, errors.New("failed to read image file")
		}
		req.Image = imageData
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

// AddItem is a handler to add a new item for POST /items .
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	req, err := parseAddItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var fileName string
	if req.Image != nil {
		storedFileName, err := s.storeImage(req.Image)
		if err != nil {
			slog.Error("failed to store image: ", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fileName = storedFileName
	} else {
		fileName = "default.jpg"
	}

	item := &Item{
		Name:     req.Name,
		Category: req.Category,
		Image:    fileName,
	}

	result, err := s.db.Exec("INSERT INTO items (name,category, image) VALUES (?,?,?)", item.Name, item.Category, item.Image)
	if err != nil {
		slog.Error("failed to insert item: ", "error", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("failed to get last inserted ID: ", "error", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	item.ID = int(id)

	// return response
	response := map[string]string{
		"message": fmt.Sprintf("item received: %s (category: %s, image: %s)", item.Name, item.Category, item.Image),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	// get id from URL
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/items/"))
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var item Item
	err = s.db.QueryRow("SELECT id, name, category, image FROM items WHERE id = ?", id).Scan(&item.ID, &item.Name, &item.Category, &item.Image)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)

}

func (s *Handlers) GetAllItems(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id, name, category, image FROM items")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image); err != nil {
			http.Error(w, "Failed to scan item", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(items)
	if err != nil {
		http.Error(w, "Failed to encode items", http.StatusInternalServerError)
		return
	}
}

// storeImage stores an image and returns the file path and an error if any.
// this method calculates the hash sum of the image as a file name to avoid the duplication of a same file
// and stores it in the image directory.
func (s *Handlers) storeImage(image []byte) (filePath string, err error) {
	hash := sha256.Sum256(image)
	hashStr := hex.EncodeToString(hash[:])
	fileName := hashStr + ".jpg"                     // - calc hash sum
	filePath = filepath.Join(s.imgDirPath, fileName) // - build image file path

	if _, err := os.Stat(filePath); err == nil {
		return fileName, nil //if the file already exists, just return file
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to check image file existence: %w", err)
	}

	if err := os.MkdirAll(s.imgDirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create image directory: %w", err)
	}

	err = os.WriteFile(filePath, image, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return fileName, nil // - return the image file path
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
		if errors.Is(err, errImageNotFound) {
			slog.Warn("image not found", "filename", req.FileName)
			http.Error(w, fmt.Sprintf("Image not found: %s", req.FileName), http.StatusNotFound)
			return
		}

		slog.Warn("failed to build image path: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
