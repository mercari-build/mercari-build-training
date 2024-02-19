package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
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
	Name     string `json:"name"`
	Category string `json:"category"`
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

func addItem(c echo.Context) error {
	name := c.FormValue("name")
	category := c.FormValue("category")

	newItem := Item{Name: name, Category: category}

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

	message := fmt.Sprintf("Item : %s, Category: %s", newItem.Name, newItem.Category)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
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
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
