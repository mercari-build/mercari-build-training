package main

import (
	"crypto/sha256"
	"encoding/hex"
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

type Items struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

/*
3-4 POST Add an image to an item

	returns Sha256 hashed string of File
*/
func getFileSha256(c echo.Context, fileType string) (string, error) {
	fileHeader, err := c.FormFile(fileType)
	if err != nil {
		return "", err
	}
	f, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

/*
open given files and parse json items,
return items.Items
*/
func readItemsFromFile(filePath string) ([]Item, error) {
	jsonFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var items Items
	if err := json.Unmarshal(jsonFile, &items); err != nil {
		return nil, err
	}
	return items.Items, nil
}

/*
3-3 GET get a list of items
*/
func getItem(c echo.Context) error {
	items, err := readItemsFromFile("items.json")

	if err != nil {
		msg := "items.json file not exist!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	return c.JSON(http.StatusOK, items)
}

/*
3-2, 3-4 POST add the list of item

 1. Get form data, make new <Item> newItem variable
 2. store current items in []item array.
 3. append newItem in items array, marshal item to valid Json
 4. write items.json with updated items
*/
func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	//get hashed image
	image, err := getFileSha256(c, "image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	image += ".jpg"
	c.Logger().Infof("Receive item: %s", name)
	newItem := Item{
		Name:      name,
		Category:  category,
		ImageName: image,
	}
	message := fmt.Sprintf("item received: %s", name)
	// Open items file
	items, err := readItemsFromFile("items.json")
	// append newItem in items
	if err != nil {
		msg := "error happend while reading elements in items.json!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	items = append(items, newItem)
	// marshal new items, write back to items.json
	updatedItems, err := json.Marshal(&items)
	if err != nil {
		msg := "error happend while converting new items to json structure!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	os.WriteFile("items.json", updatedItems, 0644)

	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

/*
3-5 GET Return item details

 1. Unmarshal current items in []items array
 2. return items[itemId]
*/
func getItemById(c echo.Context) error {
	itemId := c.Param("itemId")
	items, err := readItemsFromFile("items.json")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	parsedInteger, err := strconv.ParseInt(itemId, 10, 64)
	if err != nil || parsedInteger >= int64(len(items)) {
		msg := "Error occured while converting itemId to integer!"
		return c.JSON(http.StatusBadRequest, msg)
	}
	return c.JSON(http.StatusOK, items[parsedInteger])
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
	e.GET("/items", getItem)
	e.GET("/items/:itemId", getItemById)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
