package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"io/ioutil"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Item struct {
	Name 		string `json:"name"`
	Category 	string `json:"category"`
	Image 		string `json:"image"`
	IDNumber 	string `json:"idnumber"`
}

type Items struct {
	Items []Item `json:"items"`
}

const (
	ImgDir = "images"
	ItemsJson = "app/items.json"
)

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func readItemsFromFile() ([]Item, error) {
	data, err := ioutil.ReadFile(ItemsJson)
	if err != nil {
		return nil, err
	}
	
	var items Items

	err = json.Unmarshal(data, &items)
	if err != nil {
		return nil, err
	}

	return items.Items, nil
}

func getItems(c echo.Context) error {
	items, err := readItemsFromFile()

	if err != nil {
		return err
	}
	
	return c.JSON(http.StatusOK, Items{Items:items})
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	idnumber := c.FormValue("idnumber")

	if err != nil {
		return errorHandler(err, c, http.StatusBadRequest, "Image not found")
	}
	// Making sure no two same IDs are registered
	currentItems, err := readItemsFromFile()
	if err != nil {
		return errorHandler(err, c, http.StatusInternalServerError, "Could not retrieve items")
	}

	var idList []string
	for _, item := range currentItems {
		idList = append(idList, item.IDNumber)
	}

	for _, id := range idList {
		if id == idnumber {
			res := Response{Message:"That ID already exists"}
			return c.JSON(http.StatusBadRequest, res)
		} 
	} 

	c.Logger().Infof("Receive item: %s. Category: %s", name, category)

	// hash img
	imgFile, err := image.Open()
	if err != nil {
		return errorHandler(err, c, http.StatusBadRequest, "Image not found")
	}
	defer imgFile.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, imgFile); err != nil {
		return errorHandler(err, c, http.StatusInternalServerError, "Couldn't hash")
	}
	hashed := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashed)

	hashString += ".jpg"

	err = addItemtoJson(name, category, hashString, idnumber)
	if err != nil {
		return errorHandler(err, c, http.StatusInternalServerError, "Could not add items")
	}

	message := fmt.Sprintf("item received: %s, %s, ID number %s", name, category, idnumber)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func addItemtoJson(name string, category string, image string, idnumber string) error {
	file, err := os.OpenFile(ItemsJson, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer file.Close() 

	currentItems, err := readItemsFromFile()
	if err != nil {
		return err
	}

	newItem := Item{Name: name, Category: category, Image: image, IDNumber: idnumber}

	currentItems = append(currentItems, newItem)

	file.Truncate(0)
	file.Seek(0, 0)

	err = json.NewEncoder(file).Encode(Items{Items:currentItems})
	if err != nil {
		return err
	}

	return nil
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

func getItemById(c echo.Context) error {
	receivedIDNumber := c.Param("idnumber")

	items, err := readItemsFromFile()
	if err != nil {
		errorHandler(err, c, http.StatusInternalServerError, "Could not open file")
	}

	for _, item := range items {
		if item.IDNumber == receivedIDNumber{
			return c.JSON(http.StatusOK, item)
		}
	}

	return c.JSON(http.StatusNotFound, Response{Message: "Item with that ID was not found"})
}

func errorHandler(err error, c echo.Context, code int, message string) *echo.HTTPError {
	return echo.NewHTTPError (code, message)
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
	e.GET("/items/:idnumber", getItemById)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
