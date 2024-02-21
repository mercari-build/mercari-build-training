package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir    = "images"
	ItemsFile = "items.json"
)

type Item struct {
	SeqID     string `json:"seq_id"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Response struct {
	Message string `json:"message"`
}

var itemsMap map[string]Item

func GenerateUUID() string {
	return uuid.New().String()
}

func mapToItemsSlice() []Item {
	items := make([]Item, 0, len(itemsMap))
	for _, item := range itemsMap {
		items = append(items, item)
	}
	return items
}

func readItemsFromFile() (*Items, error) {
	file, err := os.Open(ItemsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Items{Items: []Item{}}, nil
		}
		return nil, err
	}
	defer file.Close()

	var items Items
	// JSON to Items
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&items); err != nil {
		if err == io.EOF {
			return &Items{Items: []Item{}}, nil
		}
		return nil, err
	}

	// Create a map of items for quick access
	itemsMap = make(map[string]Item)
	for _, item := range items.Items {
		itemsMap[item.ID] = item
	}

	return &items, nil
}

func writeItemsToFile() error {
	items := mapToItemsSlice()

	file, err := os.Create(ItemsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(Items{Items: items}); err != nil {
		return err
	}
	return nil
}

func generateHashFromImage(image *multipart.FileHeader) (string, error) {
	src, err := image.Open()
	if err != nil {
		return "", fmt.Errorf("image open failed: %w", err)
	}
	defer src.Close()

	hash := sha256.New()
	copiedBytes, err := io.Copy(hash, src)
	if err != nil {
		return "", fmt.Errorf("hash generation failed: %w", err)
	}
	if image.Size != copiedBytes {
		return "", fmt.Errorf("copied bytes (%d) don't match the expected file size (%d)", copiedBytes, image.Size)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func saveImage(image *multipart.FileHeader, imagePath string) error {
	src, err := image.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded image: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("failed to create uploaded image '%s': %w", imagePath, err)
	}
	defer dst.Close()

	newOffset, err := src.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek uploaded file file: %w", err)
	}
	if newOffset != 0 {
		return fmt.Errorf("unexpected new offset during saving image: got %d, want 0", newOffset)
	}

	copiedBytes, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy uploaded file'%s': %w", imagePath, err)
	}
	if image.Size > 0 && copiedBytes != image.Size {
		return fmt.Errorf("copied bytes (%d) do not match the expected file size (%d) for '%s'", copiedBytes, image.Size, imagePath)
	}

	return nil
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	name := c.FormValue("name")
	category := c.FormValue("category")

	var imagePath string
	contentType := c.Request().Header.Get("Content-Type")
	// Check if the request contains a file
	if strings.Contains(contentType, "multipart/form-data") {
		image, err := c.FormFile("image")
		if err == nil {
			// image file exists
			imageHash, err := generateHashFromImage(image)
			if err != nil {
				c.Logger().Errorf("Image processing error: %v", err)
				return err
			}
			imagePath = filepath.Join(ImgDir, imageHash+".jpg")
			if err := saveImage(image, imagePath); err != nil {
				return err
			}
		}
	} else {
		// image file does not exist
		imagePath = "default.jpg"
	}

	uuid := GenerateUUID()
	seqID := strconv.Itoa(len(itemsMap) + 1)

	newItem := Item{
		SeqID:     seqID,
		ID:        uuid,
		Name:      name,
		Category:  category,
		ImageName: imagePath,
	}

	// Add new item to the map
	itemsMap[uuid] = newItem

	// Write updated items back to file
	if err := writeItemsToFile(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: "Failed to write item to file"})
	}

	return c.JSON(http.StatusOK, newItem)
}

func getItem(c echo.Context) error {
	itemID := c.Param("itemID")
	item, exists := itemsMap[itemID]
	if !exists {
		return c.JSON(http.StatusNotFound, Response{Message: "Item not found"})
	}
	return c.JSON(http.StatusOK, item)
}

func getItems(c echo.Context) error {
	items := mapToItemsSlice()
	return c.JSON(http.StatusOK, Items{Items: items})
}

func getImg(c echo.Context) error {
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.DEBUG)
	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Read existing items from file at startup
	items, err := readItemsFromFile()
	if err != nil && !os.IsNotExist(err) {
		e.Logger.Fatal("Failed to read items from file:", err)
	}
	if items == nil {
		items = &Items{}
	}

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:itemID", getItem)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
