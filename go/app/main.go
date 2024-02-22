package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type Items struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s", name)
	c.Logger().Infof("Receive item: %s", category)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	// open json file & data
	jsonFile, err := os.Open("items.json")
	if err != nil {
		log.Print("JSONファイルを開けません", err)
		return err
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Print("JSONデータを読み込めません", err)
		return err
	}
	// convert json into go format
	var items Items
	err = json.Unmarshal(jsonData, &items)
	if err != nil {
		log.Print("GOへの変換に失敗", err)
		return err
	}
	// add new item
	newItem := Item{Name: name, Category: category}
	items.Items = append(items.Items, newItem)

	// convert go format into json
	updatedJson, err := json.Marshal(&items)
	if err != nil {
		log.Print("JSONデータ変換に失敗", err)
		return err
	}
	// output as json file
	err = ioutil.WriteFile("items.json", updatedJson, 0644)
	if err != nil {
		log.Print("JSONファイル出力に失敗", err)
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	jsonFile, err := os.Open("items.json")
	if err != nil {
		log.Print("JSONファイルを開けません", err)
		return err
	}
	defer jsonFile.Close()
	itemsData := Items{}
	err = json.NewDecoder(jsonFile).Decode(&itemsData)
	if err != nil {
		log.Print("JSONファイルからの変換に失敗", err)
		return err
	}

	return c.JSON(http.StatusOK, itemsData)
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
