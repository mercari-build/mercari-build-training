package main

import (
	"encoding/json"
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
	ImgDir = "image"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name     string
	Category string
}

type ItemsArray struct {
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
	c.Logger().Infof("Receive item: %s", name)
	c.Logger().Infof("Receive item: %s", category)
	itemised := itemise(name, category)
	// Create File
	file, err := os.OpenFile("items.json", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		fmt.Println(err)
		filebytes, e := os.ReadFile("items.json")
		if e != nil {
			fmt.Println(e)
		}
		// Convert to str
		existingItems := string(filebytes)
		fmt.Println(existingItems)
		jsonExistingItems := decode(existingItems)
		fmt.Println(jsonExistingItems)
		jsonExistingItems.addToItemArray(itemised)
		fmt.Println(jsonExistingItems)
		jsonData, error := json.Marshal(jsonExistingItems)
		if error != nil {
			fmt.Println(error)
		}
		newfile, _ := os.Create("items.json")
		newfile.Write(jsonData)
	} else {
		jsonData := appendItem(itemised)
		file.Write(jsonData)
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func appendItem(itemised Item) []byte {
	items := []Item{}
	itemsArray := ItemsArray{items}
	itemsArray.addToItemArray(itemised)
	jsonData, err := json.Marshal(itemsArray)
	if err != nil {
		fmt.Println(err)
	}
	return jsonData
}

func decode(jsonString string) ItemsArray {
	var stcData ItemsArray
	if err := json.Unmarshal([]byte(jsonString), &stcData); err != nil {
		fmt.Println(err)
	}
	return stcData
}

func itemise(name string, category string) Item {
	item := Item{Name: name, Category: category}
	return item
}

func (itemsArray *ItemsArray) addToItemArray(item Item) []Item {
	itemsArray.Items = append(itemsArray.Items, item)
	return itemsArray.Items
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

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
