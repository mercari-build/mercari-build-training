package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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
	ImgDir    = "images"
	ItemsJson = "items.json"
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
	Id        int    `json:"id"`
}

func httpErrorHandler(err error, c echo.Context, code int, message string) *echo.HTTPError {
	c.Logger().Error(err)
	return echo.NewHTTPError(code, message)
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func loadItems() (Items, error) {
	// Load items.json
	_, err := os.Stat(ItemsJson)
	if err == nil {
		// ItemsJson exists
		file, err := os.Open(ItemsJson)
		if err != nil {
			return Items{}, err
		}
		defer file.Close()
		var items Items
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&items); err == io.EOF {
			// Empty file
			new_items := new(Items)
			return *new_items, nil
		} else if err != nil {
			return Items{}, err
		}
		return items, nil

	} else if errors.Is(err, os.ErrNotExist) {
		// ItemsJson does not exist
		new_items := new(Items)
		return *new_items, nil

	}
	// Other errors
	return Items{}, err
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	id := c.FormValue("id")
	c.Logger().Infof("Receive item: id=%s, name=%s, category=%s", id, name, category)

	// Load items.json
	items, err := loadItems()
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load items")
	}

	// Create objects
	new_item := new(Item)
	new_item.Name = name
	new_item.Category = category

	// Register id
	new_item.Id, err = strconv.Atoi(id)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusBadRequest, "id must be an integer")
	}
	if getItemIdxById(new_item.Id, items) != -1 {
		err_msg := fmt.Sprintf("id already exists: %d", new_item.Id)
		return httpErrorHandler(err, c, http.StatusBadRequest, err_msg)
	}

	// Register image
	header, err := c.FormFile("image")
	if err != nil {
		return httpErrorHandler(err, c, http.StatusBadRequest, "Image not found")
	}
	new_item.ImageName, err = registerImg(header)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to register image")
	}
	items.Items = append(items.Items, *new_item)

	// Convert item_obj to json
	file, err := os.Create(ItemsJson)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to create json")
	}
	defer file.Close()
	// Write updated items to the file
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(items); err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to write json")
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusCreated, res)
}

func registerImg(header *multipart.FileHeader) (string, error) {
	// Read uploaded file
	src, err := header.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Convert src to hash
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}
	hex_hash := hex.EncodeToString(hash.Sum(nil))

	// Reset the read position of the file
	if _, err := src.Seek(0, 0); err != nil {
		return "", err
	}

	// Save file to images/
	filename := hex_hash + path.Ext(header.Filename)
	file, err := os.Create(path.Join(ImgDir, filename))
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(file, src); err != nil {
		return "", err
	}

	return filename, nil
}

func getItems(c echo.Context) error {
	// Load items.json
	items, _ := loadItems()
	c.Logger().Infof("items: %+v", items)
	return c.JSON(http.StatusOK, items)
}

func getItemIdxById(id int, items Items) int {
	// return index if id exists, otherwise return -1
	for idx, item := range items.Items {
		if item.Id == id {
			return idx
		}
	}
	return -1
}

func getItemById(c echo.Context) error {
	// Load items
	items, err := loadItems()
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load items")
	}

	// Convert id string to int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err_msg := fmt.Sprintf("id not found: '%s'. id must be an integer", c.Param("id"))
		return httpErrorHandler(err, c, http.StatusBadRequest, err_msg)
	}
	idx := getItemIdxById(id, items)
	if idx == -1 {
		err_msg := fmt.Sprintf("id not found: %d", id)
		return httpErrorHandler(err, c, http.StatusNotFound, err_msg)
	}
	c.Logger().Infof("item: %+v", items.Items[idx])
	return c.JSON(http.StatusOK, items.Items[idx])
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath) // log level: "DEBUG"
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.Logger.SetLevel(log.INFO)
	e.Logger.SetLevel(log.DEBUG) // Print logs whose log level is no less than "DEBUG"

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
	e.GET("/items/:id", getItemById)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
