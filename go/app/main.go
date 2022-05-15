package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	//"encoding/json"
	//"io/ioutil"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"mercari-build-training-2022/app/model"
)

const (
	ImgDir = "image"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
}

type Items struct {
	Items []Item `json:"items"` 
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func handleError(c echo.Context, error_message string) error {
	c.Logger().Errorf("%s", error_message)
	res := Response{Message: error_message} 
 	return c.JSON(http.StatusBadRequest, res) 
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	item := model.Item{name, category}
	c.Logger().Infof("Receive item: %s %s", name, category)
	// Add item to db
	err := model.AddItem(item)
	if err != nil {
		handleError(c, err.Error())
	}
	message := fmt.Sprintf("item added: %s %s", name, category)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

func showItems(c echo.Context) error {
	var items model.Items
	var err error
	// Get a list of items
	items.Items, err = model.GetItems()
	if err != nil {
		handleError(c, err.Error())
	}
	return c.JSON(http.StatusOK, items)
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

func main() {
	sqlDB, err := model.DBConnection()
	if err != nil {
		fmt.Println("database error: ",err,"\n")
	}
	fmt.Println(sqlDB)
	// defer sqlDB.Close()
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
	e.GET("/items", showItems)
	e.POST("/items", addItem)
	e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
