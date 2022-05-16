package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "image"
)

type Response struct {
	Message string `json:"message"`
}

type Items struct{
	Name string `json:"name"`
	Category string `json:"category"`
}

type ResItem struct{
	Item [] Items
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func checkFile(filename string) error {
    _, err := os.Stat(filename)
    if os.IsNotExist(err) {
        _, err := os.Create(filename)
        if err != nil {
            return err
        }
    }
    return nil
}

func readFile(filename string) []Items{
  err := checkFile(filename)
  if err != nil {
      logrus.Error(err)
  }

  file, err := ioutil.ReadFile(filename)
  if err != nil {
      logrus.Error(err)
  }

  data := []Items{}

  json.Unmarshal(file, &data)
	return data
}
// POST request and adds onto JSON file with one table
func addItem(c echo.Context) error {
	// Get form data

	filename := "items.json"
	data := readFile("items.json")

	newItem := &Items{Name: c.FormValue("name"), Category: c.FormValue("category")}

  data = append(data, *newItem)

  // Preparing the data to be marshalled and written.
  dataBytes, err := json.Marshal(data)
  if err != nil {
      logrus.Error(err)
  }

  err = ioutil.WriteFile(filename, dataBytes, 0644)
  if err != nil {
      logrus.Error(err)
  }

	message := fmt.Sprintf("item received: %s", newItem.Name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
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

// GET request and retrievs from JSON file with one table 
func getItems(c echo.Context) error {
	data := readFile("items.json")
	res := ResItem{Item: data}

	return c.JSON(http.StatusOK, res)
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
	e.GET("/image/:itemImg", getImg)
	e.GET("/items", getItems)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
