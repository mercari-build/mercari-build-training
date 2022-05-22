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

	"encoding/json"
	"io/ioutil"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
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

	message := fmt.Sprintf("item received: %s, %s", name, category)
	res := Response{Message: message}

	// save to json
	raw, err := ioutil.ReadFile("./app/items.json")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	items := Items{}
	json.Unmarshal(raw, &items)
	item := Item{Name: name, Category: category}
	items.Items = append(items.Items, item)

	b_items, err := json.Marshal(items)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if err = ioutil.WriteFile("./app/items.json", b_items, os.ModePerm); err != nil {
		fmt.Println(err.Error())
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	// get from json
	raw, err := ioutil.ReadFile("./app/items.json")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	items := Items{}
	json.Unmarshal(raw, &items)
	res := items

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

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
