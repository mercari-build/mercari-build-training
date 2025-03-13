package app

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex" //Hexadecimal
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	ImageDirPath string // imageDirPath is the path to the directory storing images.
	DB           *sql.DB
}

func NewHandlers(imageDirPath string) *Handlers {
	return &Handlers{
		imageDirPath: imageDirPath,
	}
}

var ErrInvalidInput = errors.New("invalid input")

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
	db, err := sql.Open("sqlite3", "db/items.db")
	if err != nil {
		slog.Error("failed to open database: ", "error", err)
		return 1
	}

	repo, err := NewItemRepository()
	if err != nil {
		slog.Error("failed to create item repository: ", "error", err)
		return 1
	}

	imageDir, found := os.LookupEnv("IMAGE_DIR_PATH")
	if !found {
		s.ImageDirPath = "images"
		log.Println("Using default image directory: ", s.ImageDirPath)
	}

	s.ImageDirPath = imageDir

	// set up handlers
	h := &Handlers{imageDirPath: s.ImageDirPath, db: db, repo: repo}

	if db == nil {
		slog.Error("Database connection is nil")
	}
	if repo == nil {
		slog.Error("ItemRepository is nil")
	}
	// set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)
	mux.HandleFunc("GET /items/{id}", h.GetItem)
	mux.HandleFunc("GET /items", h.GetAllItems)
	mux.HandleFunc("GET /search", h.Getkeyword)

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
	imageDirPath string  // imiageDirPath is the path to the directory storing images.
	db           *sql.DB //define struct of db
	repo         ItemRepository
}
type HelloResponse struct {
	Message string `json:"message"`
}

type AddItemRequest struct {
	Name     string `form:"name"`
	Category string `form:"category"`
	Image    []byte `form:"image"`
}

type AddItemResponse struct {
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

// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	name := r.FormValue("name")
	category := r.FormValue("category")
	file, _, err := r.FormFile("image")
	if err != nil {
		log.Println("No image file found, using default image")
		file = nil
	} else {
		defer file.Close()
	}

	var image []byte
	if file != nil {
		image, err = io.ReadAll(file) //read image file
		if err != nil {
			log.Printf("Error reading image file: %v", err)
			return nil, fmt.Errorf("failed to read image file: %w", err)
		}
	}

	if name == "" {
		log.Println("Name is empty")
		return nil, errors.New("name is required")
	}
	if category == "" {
		log.Println("Category is empty")
		return nil, errors.New("category is required")
	}

	return &AddItemRequest{
		Name:     name,
		Category: category,
		Image:    image,
	}, nil
}

// AddItem is a handler to add a new item for POST /items .
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	req, err := parseAddItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//save image
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
		fileName = "default.jpg" // if req.Image is nil, set fileName as default.jpg
	}

	item := &Item{
		Name:     req.Name,
		Category: req.Category,
		Image:    fileName,
	}

	err = s.repo.Insert(r.Context(), item)

	if err != nil {
		log.Println("Failed to insert item:", err)

		if errors.Is(err, ErrInvalidInput) {
			http.Error(w, "Invalid input", http.StatusBadRequest) //400
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError) //500
		}
		return
	}

	// return response
	response := map[string]interface{}{
		"id":       item.ID,
		"name":     item.Name,
		"category": item.Category,
		"image":    item.Image,
	}

	// return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		log.Println("failed to encode JSON: ", err)
		return
	}

	// debug
	file, err := os.Create("response.json")
	if err != nil {
		log.Printf("failed to create response.json: %v\n", err)
	} else {
		defer file.Close()
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(response); err != nil {
			log.Printf("failed to write response.json: %v\n", err)
		}
	}
}

func (s *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	// get id from URL
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/items/"))
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// call GetItemByID on infra.go
	item, err := s.repo.GetItemByID(context.Background(), id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("Database error:", err)
		return
	}
	// if there are no matching items
	if item == nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	//  return JSON response
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		log.Println("failed to encode JSON: ", err)
		return
	}

}

func (s *Handlers) GetAllItems(w http.ResponseWriter, r *http.Request) {
	// call getAll function on infra.go
	items, err := s.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("Database error:", err)
		return
	}

	// return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		log.Println("Failed to encode JSON:", err)
		return
	}
}

func (s *Handlers) Getkeyword(w http.ResponseWriter, r *http.Request) {
	//get keyword from query parameter
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "keyword parameter is required", http.StatusBadRequest)
		return
	}
	//call Getkeyword func from infra.go
	items, err := s.repo.GetKeyword(context.Background(), keyword)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("Database error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		log.Println("Failed to encode JSON:", err)
		return
	}

}

// no change in step5 from now on

// storeImage stores an image and returns the file path and an error if any.
// this method calculates the hash sum of the image as a file name to avoid the duplication of a same file
// and stores it in the image directory.
func (s *Handlers) storeImage(image []byte) (filePath string, err error) {
	s.imageDirPath = "images"
	if s.imageDirPath == "" {
		return "", fmt.Errorf("image directory path is not set")
	}

	hash := sha256.Sum256(image)
	hashStr := hex.EncodeToString(hash[:])
	fileName := hashStr + ".jpg"                       // - calc hash sum
	filePath = filepath.Join(s.imageDirPath, fileName) // - build image file path

	if _, err := os.Stat(filePath); err == nil {
		return fileName, nil //if the file already exists, just return file
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to check image file existence: %w", err)
	}

	if err := os.MkdirAll(s.imageDirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create image directory: %w", err)
	}

	if err = os.WriteFile(filePath, image, 0644); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}
	log.Println("Image successfully stored")
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
	imgPath := filepath.Join(s.imageDirPath, filepath.Clean(imageFileName))

	// to prevent directory traversal attacks
	rel, err := filepath.Rel(s.imageDirPath, imgPath)
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
