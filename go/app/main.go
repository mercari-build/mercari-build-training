package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category_id := c.FormValue("category")
	c.Logger().Infof("Receive item: name=%s, category_id=%s", name, category_id)

	// Load items
	db, err := loadDb(DbPath)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load database")
	}
	defer db.Close()

	// Create objects
	new_item := new(Item)
	new_item.Name = name

	new_item.CategoryId, err = strconv.Atoi(category_id)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusBadRequest, "category_id must be an integer")
	}
	if _, err := loadCategoryById(db, new_item.CategoryId); err != nil {
		return httpErrorHandler(err, c, http.StatusBadRequest, "category_id not found")
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

	// Insert new item to database
	err = insertItem(db, *new_item)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to insert item")
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusCreated, res)
}

func getAllItems(c echo.Context) error {
	// Load items
	db, err := loadDb(DbPath)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load database")
	}
	defer db.Close()

	joined_items, err := joinItemsAndCategories(db)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to join items and categories")
	}
	return c.JSON(http.StatusOK, joined_items)
}

func getItemById(c echo.Context) error {
	// Load items
	db, err := loadDb(DbPath)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load database")
	}
	defer db.Close()

	// Convert id string to int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err_msg := fmt.Sprintf("id not found: '%s'. id must be an integer", c.Param("id"))
		return httpErrorHandler(err, c, http.StatusBadRequest, err_msg)
	}

	// Load item by id
	item, err := loadItemById(db, id)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load item")
	}
	if item == nil {
		err_msg := fmt.Sprintf("id not found: %d", id)
		err = fmt.Errorf(err_msg)
		return httpErrorHandler(err, c, http.StatusNotFound, err_msg)
	}

	// Join item and category name
	joined_item, err := joinItemAndCategory(db, *item)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to join item and category")

	}
	return c.JSON(http.StatusOK, joined_item)
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

func searchItems(c echo.Context) error {
	// Load items
	db, err := loadDb(DbPath)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load database")
	}
	defer db.Close()

	items, err := loadItemsByQuery(db, "SELECT * FROM items")
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load items")
	}

	// Get keyword
	// example: curl -X GET 'http://127.0.0.1:9000/search?keyword=jacket'
	keyword := c.QueryParam("keyword")
	c.Logger().Infof("keyword=%s", keyword)

	// Search items
	var res_items Items
	for _, item := range items.Items {
		if strings.Contains(item.Name, keyword) {
			res_items.Items = append(res_items.Items, item)
		}
	}
	c.Logger().Infof("items: %+v", res_items)
	return c.JSON(http.StatusOK, res_items)
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
	e.GET("/items", getAllItems)
	e.GET("/items/:id", getItemById)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/search", searchItems)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
