package main

import (
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

/*
3-3 Get a list of items

	<---- no need struct since error type return
	only requires entire json to string. if something
	has to done with seperate items or item, use this code
	---->
	items : [{name,category},,,]
	declare two structs for items and item
	type Items struct {
		Items []Item `json:"items"`
	}

	type Item struct {
		Name     string `json:"name"`
		Category string `json:"category"`
	}
*/

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

/*
3-3 Get a list of items

	GET getItem (<echo.Context> C)
	parse the json file from items.json
	return entire elements of items.json
	as array.
*/
func getItem(c echo.Context) error {
	json, err := os.Open("items.json")

	if err != nil {
		res := "items.json file not exist!"
		return c.JSON(http.StatusOK, res)
	}
	defer json.Close()

	//ioutil.ReadAll deprecated from go 1.16
	byteValue, err := io.ReadAll(json)
	if err != nil {
		res := "no items exist!"
		return c.JSON(http.StatusOK, res)
	}

	ret := string(byteValue)
	return c.JSONBlob(http.StatusOK, []byte(ret))
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
	//3-3. GET a list of items
	e.GET("/items", getItem)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
