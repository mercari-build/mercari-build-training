package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3" // SQLite3 driver
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

type Item struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

type Items struct {
	Items []Item `json:"items"`
}

func addItem(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// リクエストボディからデータを取得
		name := c.FormValue("name")
		category := c.FormValue("category")

		// 画像ファイルを取得
		imageFile, err := c.FormFile("image")
		if err != nil {
			res := Response{Message: "Failed to get image file"}
			return c.JSON(http.StatusBadRequest, res)
		}
		src, err := imageFile.Open()
		if err != nil {
			res := Response{Message: "Failed to open image file"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		defer src.Close()

		// 画像ファイルをハッシュ化
		hash := sha256.New()
		if _, err := io.Copy(hash, src); err != nil {
			res := Response{Message: "Failed to hash image file"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		hashedImageName := fmt.Sprintf("%x.jpeg", hash.Sum(nil))

		// 画像ファイルを保存
		dst, err := os.Create("images/" + hashedImageName)
		if err != nil {
			res := Response{Message: fmt.Sprintf("Failed to create image file: %s", hashedImageName)}
			return c.JSON(http.StatusInternalServerError, res)
		}
		defer dst.Close()
		src.Seek(0, 0) // ファイルポインタを先頭に戻す
		//srcからdstへ内容をコピー
		if _, err := io.Copy(dst, src); err != nil {
			res := Response{Message: "Failed to save image file"}
			return c.JSON(http.StatusInternalServerError, res)
		}

		// 新しいアイテムをDBに追加
		stmt, err := db.Prepare("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)")
		if err != nil {
			res := Response{Message: "Failed to prepare SQL statement"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		defer stmt.Close()
		if _, err = stmt.Exec(name, category, hashedImageName); err != nil {
			res := Response{Message: "Failed to execute SQL statement"}
			return c.JSON(http.StatusInternalServerError, res)
		}

		c.Logger().Infof("Receive item: %s", name)

		message := fmt.Sprintf("item received: name=%s,category=%s,images=%s", name, category, hashedImageName)
		res := Response{Message: message}

		return c.JSON(http.StatusOK, res)
	}
}

func getAllItems(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// DBから全てのアイテムを取得
		rows, err := db.Query("SELECT name, category, image_name FROM items")
		if err != nil {
			res := Response{Message: "Failed to get items from DB"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		defer rows.Close()

		var allItems Items
		for rows.Next() {
			var item Item
			if err := rows.Scan(&item.Name, &item.Category, &item.ImageName); err != nil {
				res := Response{Message: "Failed to scan items from DB"}
				return c.JSON(http.StatusInternalServerError, res)
			}
			allItems.Items = append(allItems.Items, item)
		}
		return c.JSON(http.StatusOK, allItems)
	}
}

func getItemById(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		res := Response{Message: "Failed to get id in getItemById"}
		return c.JSON(http.StatusBadRequest, res)
	}
	allItems, err := LoadItemsFromJSON()
	if err != nil {
		res := Response{Message: "Failed to load items.json in getItemById"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	if id <= 0 || id > len(allItems.Items) {
		res := Response{Message: "Invalid id"}
		return c.JSON(http.StatusBadRequest, res)
	}

	return c.JSON(http.StatusOK, allItems.Items[id-1])
}

func LoadItemsFromJSON() (*Items, error) {
	jsonFile, err := os.Open("items.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var allItems Items
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&allItems); err != nil {
		return nil, err
	}
	return &allItems, nil
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Infof("Image not found: %s", imgPath)
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

	// DBへの接続
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		e.Logger.Infof("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem(db))
	e.GET("/items", getAllItems(db))
	e.GET("/items/:id", getItemById)
	e.GET("/image/:imageFilename", getImg)
	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
