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
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

const ItemsPath = "items.json"

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

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		return err
	}

	c.Logger().Infof("Receive item: %s, Category: %s", name, category)

	fileName, err := saveImage(image)
	if err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, res)
	}

	if err := saveItem(name, category, fileName); err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, res)
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

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

	err = os.MkdirAll("./images", 0777)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%x", hash) + ".jpg"
	imagePath := "./images/" + fileName

	_, err = os.Create(imagePath)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(imagePath, source, 0777)
	if err != nil {
		return "", err
	}

	return fileName, err
}

func getItems(c echo.Context) error {
	items, err := readItems(ItemsPath)
	if err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, items)
}

func readItems(filepath string) (Items, error) {
	var items Items

	data, err := os.ReadFile(filepath)
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
