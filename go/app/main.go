package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
	"bytes"
	"path"
	"strings"
	"database/sql"
	"crypto/sha256"
	"encoding/hex"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"mercari-build-training-2022/app/models"
)

const (
	ImgDir = "../image"
)

type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
	Image string `json:"image"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Response struct {
	Message string `json:"message"`
}

func getSHA256Binary(bytes[]byte) []byte {
	r := sha256.Sum256(bytes)
	return r[:]
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	var items Items

	// Init DB
	db := db.DbConnection

	// Exec Query
	rows, err := db.Query(`SELECT name, category, image FROM items`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var category string
		var image sql.NullString //NULLを許容する

		// カーソルから値を取得
		if err := rows.Scan(&name, &category, &image); err != nil {
			c.Logger().Error("Read Item from DB Error: %s", err)
			return c.JSON(http.StatusInternalServerError, err)
		}

		fmt.Printf("name: %d, category: %s, image: %s\n", name, category, image.String)
		items.Items = append(items.Items, Item{Name: name, Category: category, Image: image.String}) // image -> {"hoge", true}
	}

	return c.JSON(http.StatusOK, items)
}

func findItem(c echo.Context) error {
	var item Item
	var name string
	var category string
	var image string

	// Init DB
	db := db.DbConnection

	// Exec Query
	itemId := c.Param("id")
	c.Logger().Infof("SELECT name, category, image FROM items WHERE id = %s", itemId)
	err := db.QueryRow("SELECT name, category, image FROM items WHERE id = $1", itemId).Scan(&name, &category, &image)
	if err != nil {
		c.Logger().Error("Couldn't find data from DB: %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	item = Item{Name: name, Category: category, Image: image}

	return c.JSON(http.StatusOK, item)
}

func addItem(c echo.Context) error {
	// Inintialize Item
	var item Item
	// Get form data
	item.Name = c.FormValue("name")
	item.Category = c.FormValue("category")
	file, err := c.FormFile("image")
	if err != nil {
		c.Logger().Error("Read Image File Error: %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	// Open Image File
	imageFile, err := file.Open()
	if err != nil {
		c.Logger().Error("Open Image File Error: : %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer imageFile.Close()

	// Read Image Bytes
	imageBytes, err := io.ReadAll(imageFile)
	if err != nil {
		c.Logger().Error("Open Image Bytes Error: : %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	// Encode Image
	sha := sha256.New()
	sha.Write([]byte(imageBytes))
	item.Image = hex.EncodeToString(getSHA256Binary(imageBytes)) + ".jpg"

	c.Logger().Infof("Receive item: %s which belongs to the category %s. image name is %s", item.Name, item.Category, item.Image)

	message := fmt.Sprintf("item received: %s which belongs to the category %s. image name is %s", item.Name, item.Category, item.Image)

	// Save Image to ./image
	imgFile, err := os.Create(path.Join(ImgDir, item.Image))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	_, err = io.Copy(imgFile, bytes.NewReader(imageBytes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// Init DB
	db := db.DbConnection

	// Exec Query
	_, err = db.Exec(`INSERT INTO items (name, category, image) VALUES (?, ?, ?)`, item.Name, item.Category, item.Image)
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

func searchItems(c echo.Context) error {
	var items Items

	keyWord := c.QueryParam("keyword")
	db := db.DbConnection

	// Exec Query
	rows, err := db.Query(`SELECT name, category FROM items WHERE name LIKE ?`, keyWord + "%")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var category string
		var image string

		// カーソルから値を取得
		if err := rows.Scan(&name, &category, &image); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		items.Items = append(items.Items, Item{Name: name, Category: category, Image: image})
	}

	return c.JSON(http.StatusOK, items)
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

	// Initialize DB
	db := db.DbConnection
	defer db.Close()

	// Routes
	e.GET("/", root)
	e.GET("/items", getItems)
	e.GET("/items/:id", findItem)
	e.POST("/items", addItem)
	e.GET("/image/:itemImg", getImg)
	e.GET("/items/search", searchItems)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
