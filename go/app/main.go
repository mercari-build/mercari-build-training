package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"crypto/sha256"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)
type Response struct {
	Message    string `json:"message"`
}

// Define structure
type ItemIndex struct {
	Items      []Item `json:"items"`
}
type Item struct {	
	Name       string `json:"name"`
	Category   string `json:"category"`
	Image_name string `json:"image_name"`
}

// GET "/"
func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

var itemindex ItemIndex
var item Item

// GET "/items"
func getItem(c echo.Context) error {
	//open JSON file
	file, err := os.Open("items.json")
	if err != nil {
		c.Logger().Infof("Error message: %s", err)
	}
	defer file.Close()

	var getitem ItemIndex

	// Decode
	if err := json.NewDecoder(file).Decode(&getitem); err != nil {
		c.Logger().Infof("Error message: %s", err)
	}
	defer file.Close()

	return c.JSON(http.StatusOK, getitem)
}

// POST "/items"
func addItem(c echo.Context) error {
	// Create or open JSON file
	file, err := os.OpenFile("items.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		c.Logger().Infof("Error message: %s", err)
	}
	defer file.Close()


	// Get form data
	item.Name     = c.FormValue("name")
	item.Category = c.FormValue("category")
	image, err    := c.FormFile("image")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	image_path := fmt.Sprintf("%x", image)

	// Log
	c.Logger().Infof("Receive item: %s", item.Name)
	c.Logger().Infof("Receive category: %s", item.Category)
	c.Logger().Infof("Receive image: %s", image)

	message := fmt.Sprintf("item received: %s", item.Name)
	res := Response{Message: message}

	// Hash
	hash := sha256.Sum256([]byte(image_path))
	hash_string := fmt.Sprintf("%x", hash)
	item.Image_name = hash_string + ".jpg"

	// Add Items
	itemindex.Items = append(itemindex.Items, item)
	
	// Encode JSON
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(itemindex); err != nil {
		c.Logger().Infof("Error message: %s", err)
	 }
	 return c.JSON(http.StatusOK, res)
}

//GET "/items/:id"
func showItem(c echo.Context) error {
	// Get id & debug
	id, err := strconv.Atoi(c.Param("id")) 
	if err != nil {
		c.Logger().Infof("Error message: %s", err)
	}
	if id == 0 { 
		c.Logger().Infof("Error message: Out of range")
	}

	//open JSON file
	file, err := os.Open("items.json")
	if err != nil {
		c.Logger().Infof("Error message: %s", err)
	}
	defer file.Close()

	showitem := ItemIndex{} 

	// Decode
	if err := json.NewDecoder(file).Decode(&showitem); err != nil {
		c.Logger().Infof("Error message: %s", err)
	}
	defer file.Close()
	return c.JSON(http.StatusOK, showitem.Items[id-1])
}

//GET "/image/:imageFilename
func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Infof("Image not found: %s", imgPath)
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
	e.GET("/items", getItem)
	e.POST("/items", addItem)
	e.GET("/items/:id", showItem)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))

}
