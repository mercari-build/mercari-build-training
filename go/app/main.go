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

type Items []Item

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive item: %s", name)

	category := c.FormValue("category")
	c.Logger().Infof("Receive category: %s", category)

	// Open the JSON file
    jsonFile, err := os.Open("items.json")
    if err != nil {
        return err
    }
    defer jsonFile.Close()

    // Decode the JSON file into a Go slice
    var items Items
    byteValue, _ := ioutil.ReadAll(jsonFile)
    json.Unmarshal(byteValue, &items)

    // Add the new item to the slice
    items = append(items, Item{Name: name, Category: category})

    // Encode the slice back into JSON
    file, _ := json.MarshalIndent(items, "", " ")

    // Write the JSON back to the file
    _ = ioutil.WriteFile("items.json", file, 0644)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
    // Open the JSON file
    jsonFile, err := os.Open("items.json")
    if err != nil {
        return err
    }
    defer jsonFile.Close()

    // Decode the JSON file into a Go slice
    var items Items
    byteValue, _ := ioutil.ReadAll(jsonFile)
    json.Unmarshal(byteValue, &items)

    // Return the slice
    return c.JSON(http.StatusOK, items)
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
