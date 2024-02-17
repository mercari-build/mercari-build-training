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
	Name     string `json:"name"`
	Category string `json:"category"`
}

type AllItems struct {
	Items []Item `json:"items"`
}

func addItem(c echo.Context) error {
	// リクエストボディからデータを取得
	name := c.FormValue("name")
	category := c.FormValue("category")

	// 既存のデータを読み込む
	allItems, err := LoadItems()
	if err != nil {
		res := Response{Message: "Failed to load items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	// 新しいデータを追加
	allItems.Items = append(allItems.Items, Item{Name: name, Category: category})

	// ファイルに書き込み
	jsonData, err := json.MarshalIndent(allItems, "", "  ")
	if err != nil {
		res := Response{Message: "Failed to marshal allItems"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	err = os.WriteFile("items.json", jsonData, 0644)
	if err != nil {
		res := Response{Message: "Failed to write items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	c.Logger().Infof("Receive item: %s", name)

	message := fmt.Sprintf("item received: name=%s,category=%s", name, category)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getAllItems(c echo.Context) error {
	allItems, err := LoadItems()
	if err != nil {
		res := Response{Message: "Failed to load items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, allItems)
}

func LoadItems() (*AllItems, error) {
	jsonFile, err := os.Open("items.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var allItems AllItems
	json.Unmarshal(jsonData, &allItems)

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
	e.GET("/image/:imageFilename", getImg)
	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
