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
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir    = "images"
	ItemsFile = "items.json"
)

type Item struct {
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

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
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

	return &items, nil
}

func writeItemsToFile(items *Items) error {
	file, err := os.Create(ItemsFile)
	if err != nil {
		return err
	}
	defer file.Close()
	// Items to JSON
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(items); err != nil {
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

func addItem(c echo.Context) error {
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		return err
	}

	// create hash of the image
	imageHash, err := generateHashFromImage(image)
	if err != nil {
		return err
	}

	// save image file
	imagePath := filepath.Join(ImgDir, imageHash+".jpg")
	if err := saveImage(image, imagePath); err != nil {
		return err
	}

	newItem := Item{Name: name, Category: category, ImageName: imageHash + ".jpg"}

	// Read existing items from file
	items, err := readItemsFromFile()
	if err != nil && !os.IsNotExist(err) {
		return c.JSON(http.StatusInternalServerError, Response{Message: "Failed to read items from file"})
	}
	if items == nil {
		items = &Items{}
	}

	// Add new item to items list
	items.Items = append(items.Items, newItem)

	// Write updated items back to file
	if err := writeItemsToFile(items); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: "Failed to write item to file"})
	}

	c.Logger().Infof("Receive item: %s", name)

	return c.JSON(http.StatusOK, items)
}

func getItems(c echo.Context) error {
	items, err := readItemsFromFile()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, items)
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
	e.Logger.SetLevel(log.INFO)
	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
