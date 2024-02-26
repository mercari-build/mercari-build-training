package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir    = "images"
	ItemsPath = "items.json"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

type Items struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!!"}
	return c.JSON(http.StatusOK, res)
}

// addItem processes form data and saves item information.
func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		c.Logger().Errorf("Error while retrieving image: %v", err)
		res := Response{Message: "Error while retrieving image"}
		return echo.NewHTTPError(http.StatusBadRequest, res)
	}

	c.Logger().Infof("Receive item: %s, Category: %s", name, category)

	fileName, err := saveImage(image)
	if err != nil {
		c.Logger().Errorf("Error while hashing and saving image: %v", err)
		res := Response{Message: "Error while hashing and saving image"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	if err := saveItem(name, category, fileName); err != nil {
		c.Logger().Errorf("Error while saving item information: %v", err)
		res := Response{Message: "Error while saving item information"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

// saveItem writes the item information into the JSON file.
func saveItem(name, category, fileName string) error {

	item := Item{Name: name, Category: category, ImageName: fileName}

	items, err := readItems(ItemsPath)
	if err != nil {
		err := fmt.Errorf("error while reading or unmarshaling file: %w", err)
		return err
	}

	items.Items = append(items.Items, item)

	newData, err := json.Marshal(items)
	if err != nil {
		err := fmt.Errorf("error while marshaling file: %w", err)
		return err
	}

	if err = os.WriteFile(ItemsPath, newData, 0644); err != nil {
		err := fmt.Errorf("error while writing file: %w", err)
		return err
	}

	return nil
}

// saveImage hashes the image, saves it, and returns its file name.
func saveImage(image *multipart.FileHeader) (string, error) {

	img, err := image.Open()
	if err != nil {
		return "", err
	}
	source, err := io.ReadAll(img)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(source)

	err = os.MkdirAll("./images", 0750)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%x.jpg", hash)
	imagePath := fmt.Sprintf("./images/%s", fileName)

	_, err = os.Create(imagePath)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(imagePath, source, 0644)
	if err != nil {
		return "", err
	}

	return fileName, err
}

// getItems gets all the item information from the JSON file.
func getItems(c echo.Context) error {
	items, err := readItems(ItemsPath)
	if err != nil {
		c.Logger().Errorf("Error while reading item information: %v", err)
		res := Response{Message: "Error while reading item information"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, items)
}

// readItems reads and unmarshals the JSON file.
func readItems(filePath string) (Items, error) {
	var items Items

	data, err := os.ReadFile(filePath)
	if err != nil {
		err := fmt.Errorf("error while reading file: %w", err)
		return Items{}, err
	}

	if err = json.Unmarshal(data, &items); err != nil {
		err := fmt.Errorf("error while unmarshaling file: %w", err)
		return Items{}, err
	}

	return items, nil
}

// getImg gets the designated image by file name.
func getImg(c echo.Context) error {
	// Create image path
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

// getInfo gets information of the designeted item by id.
func getInfo(c echo.Context) error {
	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Logger().Errorf("Invalid ID: %v", err)
		res := Response{Message: "Invalid ID"}
		return echo.NewHTTPError(http.StatusBadRequest, res)
	}

	items, err := readItems(ItemsPath)
	if err != nil {
		c.Logger().Errorf("Error while reading item information: %v", err)
		res := Response{Message: "Error while reading item information"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	if itemId <= 0 || itemId > len(items.Items) {
		c.Logger().Errorf("Invalid ID: %v", err)
		res := Response{Message: "Invalid ID"}
		return echo.NewHTTPError(http.StatusBadRequest, res)
	}

	item := items.Items[(itemId - 1)]
	return c.JSON(http.StatusOK, item)
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
	e.GET("/items/:id", getInfo)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
