package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"mercari-build-training-2022/app/models"
)

const (
	ImgDir = "image"
)

type Item struct {
	Name string `json:"name"`
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

func getItems(c echo.Context) error {
	var items Items
	// path := "./items.json"

	// Open JSON file and read raw binanry
	// raw, err := os.ReadFile(path)
	// if _, err := os.Stat(path); os.IsNotExist(err) {
	// 	res := Response{Message: "No items"}
	// 	return c.JSON(http.StatusOK, res)
	// }

	// Covert raw binary to the Items struct
	// err = json.Unmarshal([]byte(raw), &items)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, err)
	// }

	// Init DB
	db := db.DbConnection
	defer db.Close()

	// Exec Query
	rows, err := db.Query(`SELECT name, category FROM items`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var category string

		// カーソルから値を取得
		if err := rows.Scan(&name, &category); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		fmt.Printf("name: %d, category: %s\n", name, category)
		items.Items = append(items.Items, Item{Name: name, Category: category})
	}

	return c.JSON(http.StatusOK, items)
}

func addItem(c echo.Context) error {
	// IInintialize Items
	// var items Items
	// Inintialize Item
	var item Item
	// Get form data
	item.Name = c.FormValue("name")
	item.Category = c.FormValue("category")

	c.Logger().Infof("Receive item: %s which belongs to the category %s", item.Name, item.Category)

	message := fmt.Sprintf("item received: %s which belongs to the category %s", item.Name, item.Category)

	// Read raw data of items from items.json
	// path := "./items.json"
	// if _, err := os.Stat(path); os.IsNotExist(err) {
	// 	_, err := os.Create(path)
	// 	if err != nil {
	// 		return c.JSON(http.StatusInternalServerError, err)
	// 	}
	// }
	// raw, err := os.ReadFile(path)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, err)
	// }
	
	// if len(raw) != 0 {
	// 	// Convert raw data to the Items struct
	// 	err = json.Unmarshal([]byte(raw), &items)
	// 	if err != nil {
	// 		return c.JSON(http.StatusInternalServerError, err)
	// 	}
	// }

	// Add Item to items
	// items.Items = append(items.Items, item)

	// Convert Items to raw JSON data
	// raw, err = json.MarshalIndent(items, "", " ")
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, err)
	// }

	// Write Items to items.json
	// err = os.WriteFile("./items.json", raw, 0644)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, err)
	// }

	// Init DB
	db := db.DbConnection
	c.Logger().Infof("DB Initialized")
	defer db.Close()

	// Exec Query
	_, err := db.Exec(`INSERT INTO items (name, category) VALUES (?, ?)`, item.Name, item.Category)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("itemImg"))

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

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
