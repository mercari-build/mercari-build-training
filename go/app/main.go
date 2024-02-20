package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"encoding/hex"

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

type itemResponse struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image := c.FormValue("image")
	c.Logger().Infof("Receive item: %s", name)
	c.Logger().Infof("Receive category: %s", category)
	c.Logger().Infof("Receive image: %s", image)

	img, err := os.Open(image)
	if err != nil {
		log.Fatal(err)
	}
	defer img.Close()
	hash := sha256.New()

	// Read the entire image file into a byte slice
	imageBytes, err := os.ReadFile(image)
	if err != nil {
		log.Fatal(err)
	}

	// Write the image data to the hash function
	hash.Write(imageBytes)

	// Get the final hash value
	hashValue := hash.Sum(nil)
	// Convert the byte slice to a hex-encoded string
	hashString := hex.EncodeToString(hashValue)
	item := Item{name, category, string(hashString)}

	message := fmt.Sprintf("item received: %s in %s category", item.Name, item.Category)
	res := Response{Message: message}

	file1, err := os.Open("item.json") //すでにあるファイルを開く
	if err != nil {
		log.Fatal(err)
	}
	defer file1.Close()
	jsonData, err := ioutil.ReadAll(file1)
	if err != nil {
		fmt.Println("JSONデータを読み込めません", err)
	}
	var itemslice []Item
	json.Unmarshal(jsonData, &itemslice)

	itemslice = append(itemslice, item)
	file2, err := os.Create("item.json") // fileはos.File型
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(file2).Encode(itemslice)

	return c.JSON(http.StatusOK, res)
}

func getItem(c echo.Context) error {
	file1, err := os.Open("item.json") //すでにあるファイルを開く
	if err != nil {
		log.Fatal(err)
	}
	defer file1.Close()
	jsonData, err := ioutil.ReadAll(file1)
	if err != nil {
		fmt.Println("JSONデータを読み込めません", err)
	}
	var itemslice []Item
	json.Unmarshal(jsonData, &itemslice)
	fmt.Println(itemslice)

	res := itemResponse{Items: itemslice}

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
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
