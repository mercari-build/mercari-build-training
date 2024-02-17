package main

import (
	"encoding/json"
	"fmt"
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

func addItem(c echo.Context) error {
	// Get form data for name and category
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Received item: %s, Category: %s", name, category)

	// Load existing items from JSON file
	var items Items
	file, err := os.ReadFile(ItemsFile)
	if err != nil || len(file) == 0 {
	    items = Items{Items: []Item{}}
	} else {
		err = json.Unmarshal(file, &items)
		if err != nil {
			c.Logger().Errorf("JSON decode error: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Server error occurred.")
		}
	}

	// Add new item to the list
	newItem := Item{Name: name, Category: category}
	items.Items = append(items.Items, newItem)

	// Save updated items back to JSON file
	itemsData, err := json.Marshal(items)
    if err != nil {
        c.Logger().Errorf("Failed to marshal items: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process items")
    }
	err = os.WriteFile(ItemsFile, itemsData, 0644)
    if err != nil {
        c.Logger().Errorf("Failed to write items to file: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save items")
    }

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	// Read the items from the JSON file
	file, err := os.ReadFile(ItemsFile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: "Error reading items file"})
	}

	var items Items
	err = json.Unmarshal(file, &items)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: "Error parsing items file"})
	}

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
