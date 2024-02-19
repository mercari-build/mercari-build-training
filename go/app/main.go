package main

import (
	"crypto/sha256"
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

func addItem(c echo.Context) error {
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

	// JSONファイルの既存のデータを読み込む
	allItems, err := LoadItemsFromJSON()
	if err != nil {
		res := Response{Message: "Failed to load items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	// 新しいデータを追加
	allItems.Items = append(allItems.Items, Item{Name: name, Category: category, ImageName: hashedImageName})

	// ファイルに書き込み
	jsonData, err := json.MarshalIndent(allItems, "", "  ")
	if err != nil {
		res := Response{Message: "Failed to marshal allItems"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	// 0644: パーミッション(所有者は読み書き可能)
	if err = os.WriteFile("items.json", jsonData, 0644); err != nil {
		res := Response{Message: "Failed to write items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	c.Logger().Infof("Receive item: %s", name)

	message := fmt.Sprintf("item received: name=%s,category=%s,images=%s", name, category, hashedImageName)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getAllItems(c echo.Context) error {
	allItems, err := LoadItemsFromJSON()
	if err != nil {
		res := Response{Message: "Failed to load items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, allItems)
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
	e.GET("/items", getAllItems)
	e.GET("/items/:id", getItemById)
	e.GET("/image/:imageFilename", getImg)
	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
