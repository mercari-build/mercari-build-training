package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
	ItemsJson = "items.json"
)

type Response struct {
	Message string `json:"message"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func loadItems() Items {
	// Load items.json
	file, err := os.Open(ItemsJson)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var items Items
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&items); err != nil {
		log.Fatal(err)
	}
	return items
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: name=%s, category=%s", name, category)

	// Load items.json
	items := loadItems()
	c.Logger().Infof("items: %+v", items)

	// Create objects
	new_item := new(Item)
	new_item.Name = name
	new_item.Category = category
	items.Items = append(items.Items, *new_item)

	// Convert item_obj to json
	file, err := os.Create(ItemsJson)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// Write updated items to the file
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(items); err != nil {
		log.Fatal(err)
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	// Load items.json
	items := loadItems()
	c.Logger().Infof("items: %+v", items)
	return c.JSON(http.StatusOK, items)
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
