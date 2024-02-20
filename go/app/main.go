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
	"path/filepath"
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

type Item struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

type ItemsList struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func saveItemToFile(name string, category string, imageName string) error {
	currentItems, err := os.ReadFile("items.json")
	if err != nil {
		return err
	}

	var itemsList ItemsList
	json.Unmarshal(currentItems, &itemsList) //JSONデータの読み込み

	newItem := Item{Name: name, Category: category, ImageName: imageName}
	itemsList.Items = append(itemsList.Items, newItem)

	result, err := json.Marshal(itemsList) //to JSON structure
	if err != nil {
		return err
	}

	erro := os.WriteFile("items.json", result, 0666)
	if erro != nil {
		return erro
	}

	return nil
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")

	// Receive image files
	file, err := c.FormFile("imageName")
	if err != nil {
		return err
	}
	c.Logger().Infof("Receive item: %s", name)

	// Open file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Read file and calculate hash value
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return err
	}
	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	// Generate file names from hash values
	img_name := hashString + ".jpg"

	// Save images in the images directory
	dst, err := os.Create(filepath.Join(ImgDir, img_name))
	if err != nil {
		return err
	}
	defer dst.Close()

	// move the file pointer back to the beginning
	src.Seek(0, io.SeekStart)
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	erro := saveItemToFile(name, category, img_name)
	if erro != nil {
		c.Logger().Infof("Error: %s", erro)
		return erro
	}

	c.Logger().Infof("Receive item: %s", name)
	message := fmt.Sprintf("Receive item: %s; Category: %s", name, category)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	currentItems, err := os.ReadFile("items.json")
	if err != nil {
		return err
	}
	var itemsList ItemsList
	if json.Unmarshal(currentItems, &itemsList); err != nil {
		c.Logger().Infof("Error: %s", err)
		return err
	}
	return c.JSON(http.StatusOK, itemsList)
}

func getItemById(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	currentItems, err := os.ReadFile("items.json")
	if err != nil {
		return err
	}
	var itemsList ItemsList
	if json.Unmarshal(currentItems, &itemsList); err != nil {
		c.Logger().Infof("Error: %s", err)
		return err
	}
	return c.JSON(http.StatusOK, itemsList.Items[id-1])
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
	e.Logger.SetLevel(log.DEBUG)

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
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	e.GET("/items/:id", getItemById)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
