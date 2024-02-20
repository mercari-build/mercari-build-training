package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"io/ioutil"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

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

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive item: %s", name)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
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

func readItemsFromFile(filename string) ([]Item, error) {
	// ファイルをオープン
	f, err := os.Open(filename)
	if err != nil {
		log.Errorf("Error opening file: %s", err)
		return nil, err
	}
	defer f.Close()

	// ファイルの内容を読み取り
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Errorf("Error reading file: %s", err)
		return nil, err
	}

	// JSONデコード
	var items []Item
	err = json.Unmarshal(bytes, &items)
	if err != nil {
		log.Errorf("Error decoding JSON: %s", err)
		return nil, err
	}

	return items, nil
}

func getItems(c echo.Context) error {
	items, err := readItemsFromFile("items.json")
	if err != nil {
		errMsg := fmt.Sprintf("Internal Server Error: %s", err) // エラーメッセージを作成
		fmt.Println(errMsg) // エラーメッセージを出力
		return c.JSON(http.StatusInternalServerError, Response{Message: errMsg})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"items": items})
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
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
