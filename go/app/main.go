package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	//use '_' for registering the driver with the 'database/sql'
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Item struct {
	ItemId    int64  `json:"item_id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

/*
getDatabasePath()

return absolute path of mercari.sqlite3 database file
which always will be located one up level of go directory.
*/
func getDatabasePath() string {
	path, err := os.Executable()
	if err != nil {
		msg := "error occured while getting path!"
		return msg
	}
	dir := filepath.Dir(path)
	dbPath := filepath.Join(dir, "..", "mercari.sqlite3")
	finalPath, err := filepath.Abs(dbPath)
	if err != nil {
		msg := "error occured while concatnating path!"
		return msg
	}
	return finalPath
}

func getFileSha256(c echo.Context, fileType string) (string, error) {
	fileHeader, err := c.FormFile(fileType)
	if err != nil {
		return "", err
	}
	f, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

/*
4-1 GET write into a database

 1. open db
 2. O(n^2). iterate over rows and colums
*/
func getItem(c echo.Context) error {
	path := getDatabasePath()
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		msg := "error occured while opening db!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	defer db.Close()
	//if table not exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        category TEXT NOT NULL,
        image_name TEXT
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		msg := "error occured while creating new DB!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	var items Items
	rows, err := db.Query("SELECT id, name, category, image_name FROM items")
	if err != nil {
		msg := "error occured while reading rows from db!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	defer rows.Close()
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ItemId, &item.Name, &item.Category, &item.ImageName); err != nil {
			msg := "error occured while scanning row from db!"
			return c.JSON(http.StatusBadRequest, msg)
		}
		items.Items = append(items.Items, item)
	}
	return c.JSON(http.StatusOK, items)
}

/*
4-1 POST write into a database

 1. open db
 2. insert item. id will be autoincremented
*/
func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	//get hashed image
	image, err := getFileSha256(c, "image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	image += ".jpg"
	c.Logger().Infof("Receive item: %s", name)
	message := fmt.Sprintf("item received: %s", name)
	// Open items file
	path := getDatabasePath()
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		msg := "error occured while opening db!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	newItem := Item{
		Name:      name,
		Category:  category,
		ImageName: image,
	}
	_, err = db.Exec("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)", &newItem.Name, &newItem.Category, &newItem.ImageName)
	if err != nil {
		msg := "error occured while inserting new item!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

/*
4-1 POST write into a database

 1. open db
 2. query row such that matches id. id index starts from 1
*/
func getItemById(c echo.Context) error {
	idStr := c.Param("itemId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		msg := "id is not valid integer!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	var item Item
	path := getDatabasePath()
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		msg := "error occured while reading db!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	defer db.Close()
	err = db.QueryRow("SELECT id, name, category, image_name FROM items WHERE id = ?", id).Scan(&item.ItemId, &item.Name, &item.Category, &item.ImageName)
	if err != nil {
		msg := "Invalid ID!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	return c.JSON(http.StatusOK, item)
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

/*
4-2 GET Search for an item

 1. open DB
 2. query DB based on keyword params
 3. add every elements that matches conditions and returns items array
*/
func getSearch(c echo.Context) error {
	keyword := c.QueryParam("keyword")
	// in sql, % keyword % will search any
	// results that contains keyword inside the word.
	keyword = "%" + keyword + "%"
	path := getDatabasePath()
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		msg := path
		return c.JSON(http.StatusBadRequest, msg)
	}
	defer db.Close()
	rows, err := db.Query("SELECT id, name, category, image_name FROM items WHERE name LIKE ? OR category LIKE ?", keyword, keyword)
	if err != nil {
		msg := "error occured while querying db!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	var items Items
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ItemId, &item.Name, &item.Category, &item.ImageName); err != nil {
			msg := "error occured while copying db to variable!"
			return c.JSON(http.StatusBadRequest, msg)
		}
		items.Items = append(items.Items, item)
	}
	return c.JSON(http.StatusBadRequest, items)
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
	e.POST("/items", addItem)
	e.GET("/items", getItem)
	e.GET("/items/:itemId", getItemById)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/search", getSearch)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
