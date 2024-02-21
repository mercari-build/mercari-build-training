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
	category_name := c.FormValue("category")
	c.Logger().Infof("Receive item: name=%s, category=%s", name, category_name)

	// Load items
	db, err := loadDb(DbPath)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load database")
	}
	defer db.Close()

	// Search or insert category
	category, err := loadCategoryByName(db, category_name)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load category")
	}
	if category == nil {
		err = insertCategory(db, category_name)
		if err != nil {
			return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to insert category")
		}
		category, err = loadCategoryByName(db, category_name)
		if err != nil {
			return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load category")
		}
	}

	// Register image
	header, err := c.FormFile("image")
	if err != nil {
		return httpErrorHandler(err, c, http.StatusBadRequest, "Image not found")
	}
	image_name, err := registerImg(header)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to register image")
	}

	// Insert new item to database
	err = insertItem(db, name, category.Id, image_name)
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

	joined_items, err := joinAll(db)
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
	// Get keyword
	keyword := c.QueryParam("keyword")
	c.Logger().Infof("keyword=%s", keyword)

	// Load items
	db, err := loadDb(DbPath)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load database")
	}
	defer db.Close()

	joined_items, err := loadJoinedItemsByKeyword(db, keyword)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to search items")
	}

	c.Logger().Infof("items: %+v", joined_items)
	return c.JSON(http.StatusOK, joined_items)
}

func addCategory(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive category: name=%s", name)

	// Load items
	db, err := loadDb(DbPath)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to load database")
	}
	defer db.Close()

	// Insert new category to database
	err = insertCategory(db, name)
	if err != nil {
		return httpErrorHandler(err, c, http.StatusInternalServerError, "Failed to insert category")
	}

	message := fmt.Sprintf("category received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusCreated, res)
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
	e.POST("/categories", addCategory)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
