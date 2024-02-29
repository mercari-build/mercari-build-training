package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
	//dbPath = "../db/mercari.sqlite3"
	dbPath = "/app/db/mercari.sqlite3"
)

type Item struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CategoryId string `json:"category"`
	ImageName  string `json:"image_name"`
}

type Items struct {
	Items []*Item `json:"items"`
}

type Response struct {
	Message string `json:"message"`
}

var itemsMap map[string]Item

func generateHashFromImage(image *multipart.FileHeader) (string, error) {
	src, err := image.Open()
	if err != nil {
		return "", fmt.Errorf("image open failed: %w", err)
	}
	defer src.Close()

	hash := sha256.New()
	copiedBytes, err := io.Copy(hash, src)
	if err != nil {
		return "", fmt.Errorf("hash generation failed: %w", err)
	}
	if image.Size != copiedBytes {
		return "", fmt.Errorf("copied bytes (%d) don't match the expected file size (%d)", copiedBytes, image.Size)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func saveImage(image *multipart.FileHeader, imagePath string) error {
	src, err := image.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded image: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("failed to create uploaded image '%s': %w", imagePath, err)
	}
	defer dst.Close()

	newOffset, err := src.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek uploaded file file: %w", err)
	}
	if newOffset != 0 {
		return fmt.Errorf("unexpected new offset during saving image: got %d, want 0", newOffset)
	}

	copiedBytes, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy uploaded file'%s': %w", imagePath, err)
	}
	if image.Size > 0 && copiedBytes != image.Size {
		return fmt.Errorf("copied bytes (%d) do not match the expected file size (%d) for '%s'", copiedBytes, image.Size, imagePath)
	}

	return nil
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	name := c.FormValue("name")
	category := c.FormValue("category_id")
	image, err := c.FormFile("image")
	var imagePath string
	if err != nil {
		imagePath = "default.jpg"
	} else {
		imageHash, err := generateHashFromImage(image)
		if err != nil {
			c.Logger().Errorf("Image processing error: %v", err)
			return err
		}
		imagePath = filepath.Join(ImgDir, imageHash+".jpg")
		if err := saveImage(image, imagePath); err != nil {
			return err
		}
	}

	// get a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	// invoke SQL to collect all of items
	stmt, err := db.Prepare("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer stmt.Close()
	result, err := stmt.Exec(name, category, imagePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve new item ID")
	}

	newItem := Item{
		ID:         strconv.Itoa(int(newID)),
		Name:       name,
		CategoryId: category,
		ImageName:  imagePath,
	}

	itemsMap[strconv.Itoa(int(newID))] = newItem

	return c.JSON(http.StatusOK, newItem)
}

func getItem(c echo.Context) error {
	itemID := c.Param("itemID")
	if item, exists := itemsMap[itemID]; exists {
		return c.JSON(http.StatusOK, item)
	}

	// データベースからアイテムを取得する処理を追加
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database connection failed")
	}
	defer db.Close()

	var item Item
	err = db.QueryRow("SELECT id, name, category_id, image_name FROM items WHERE id = ?", itemID).Scan(&item.ID, &item.Name, &item.CategoryId, &item.ImageName)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, Response{Message: "Item not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database query failed")
	}

	// アイテムを見つけたらJSONとして返す
	return c.JSON(http.StatusOK, item)
}

func getItems(c echo.Context) error {
	// get a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	// invoke SQL to collect all of items
	rows, err := db.Query("SELECT id, name, category_id, image_name FROM items")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	// map the items to returned variable
	items := Items{Items: []*Item{}}
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.CategoryId, &item.ImageName)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		items.Items = append(items.Items, &item)
	}

	// return them as JSON
	return c.JSON(http.StatusOK, Items{Items: items.Items})
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

func searchItem(c echo.Context) error {
	// get keyword value from the query
	keyword := c.QueryParam("keyword")

	// get a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	// invoke SQL to enumerate items by filtering with the `keyword` value
	rows, err := db.Query("SELECT id, name, category_id, image_name FROM items WHERE name LIKE ?", "%"+keyword+"%")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	// map the items to returned variable
	items := Items{Items: []*Item{}}
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.CategoryId, &item.ImageName)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		items.Items = append(items.Items, &item)
	}

	// return them as JSON
	return c.JSON(http.StatusOK, items)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.DEBUG)
	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Create a map of items for quick access
	itemsMap = make(map[string]Item)

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:itemID", getItem)
	e.GET("/search", searchItem)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
