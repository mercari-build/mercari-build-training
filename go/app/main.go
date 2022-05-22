package main

import (
	"fmt"
	"mercari-build-training-2022/app/item_store"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"strconv"
	// _ "github.com/mattn/go-sqlite3"
)

const sqlite3_path string = "../db/mercari.sqlite3"

func InsertItem(name string, category string, image string) error {
	db, _ := sql.Open("sqlite3", sqlite3_path)
	defer db.Close()

	command := "INSERT INTO items (name, category, image) VALUESn(?, ?, ?)"
	_, err := db.Exec(command, name, category, image)
	if err != nil {
		fmt.Println(err.Error())
	}

	return err
}

/* 関数定義 */

func GetItems() (*sql.Rows, error) {
	db, _ := sql.Open("sqlite3", sqlite3_path)
	defer db.Close()

	command := "SELECT name, category, image FROM items"
	rows, err := db.Query(command)
	if err != nil {
		fmt.Println(err.Error())
	}
	return rows, err
}

func GetItemById(id int) *sql.Row {
	db, _ := sql.Open("sqlite3", "../db/mercari.sqlite3")
	defer db.Close()

	command := "SELECT name, category, image FROM items WHERE id = ?"
	row := db.QueryRow(command, id)

	return row
}

func getItemById(c echo.Context) error {
	strId := c.Param("id")
	id, _ := strconv.Atoi(strId)
	row := GetItemById(id)
	var item Item
	if err := row.Scan(&item.Name, &item.Category, &item.Image); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println(err.Error())
			message := fmt.Sprintf("No row")
			res := Response{Message: message}
			return c.JSON(http.StatusOK, res)
		} else {
			fmt.Println(err.Error())
			return err
		}
	}

	return c.JSON(http.StatusOK, item)
}

func SerchItems(keyword string) (*sql.Rows, error) {
	db, _ := sql.Open("sqlite3", "../db/mercari.sqlite3")
	defer db.Close()

	command := "SELECT name, category FROM items WHERE name LIKE ?"
	rows, err := db.Query(command, "%"+keyword+"%")
	if err != nil {
		fmt.Println(err.Error())
	}

	return rows, err
}

/* 関数定義 終わり */

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image"`
}

type Items struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s, %s", name, category)
	file, err := c.FormFile("image")
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("error")
		return err
	}
	src, err := file.Open()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer src.Close()

	fileName := strings.Split(file.Filename, ".")[0]
	sha256 := sha256.Sum256([]byte(fileName))
	hashedFileName := hex.EncodeToString(sha256[:]) + ".jpg"

	saveFile, err := os.Create(path.Join(ImgDir, hashedFileName))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer saveFile.Close()

	if _, err = io.Copy(saveFile, src); err != nil {
		fmt.Println(err.Error())
		return err
	}

	if err := InsertItem(name, category, hashedFileName); err != nil {
		fmt.Println(err.Error())
		return err
	}
	message := fmt.Sprintf("item received: %s, %s", name, category)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {

	rows, err := item_store.GetItems()
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	var items Items
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Name, &item.Category, &item.Image); err != nil {
			fmt.Println(err.Error())
			return err
		}
		items.Items = append(items.Items, item)
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

func searchNameByKeyword(c echo.Context) error {
	keyword := c.FormValue("keyword")
	rows, err := item_store.SearchItems(keyword)
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	var items Items
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Name, &item.Category); err != nil {
			fmt.Println(err.Error())
			return err
		}
		items.Items = append(items.Items, item)
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

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:id", getItemById)
	e.GET("/search", searchNameByKeyword)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
