package main

import (
	"fmt"
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"mercari-build-training-2022/app/model"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	ImgDir   = "image"
	dbSchema = "../db/items.db"
	dbSource = "../db/mercari.sqlite3"
)

var db *sql.DB

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Items struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func handleError(c echo.Context, error_message string) error {
	c.Logger().Errorf("%s", error_message)
	res := Response{Message: error_message}
	return c.JSON(http.StatusBadRequest, res)
}

func DBConnection() error {
	// open database
	db_opened, err := sql.Open("sqlite3", dbSource)
	if err != nil {
		return err
	}
	db = db_opened

	file, err := os.OpenFile(dbSchema, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	defer file.Close()

	schema, err := os.ReadFile(dbSchema)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return err
	}
	return nil
}

func DBClose() {
	db.Close()
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	item := model.Item{name, category}
	c.Logger().Infof("Receive item: %s %s", name, category)
	// Add item to db
	err := model.AddItem(item, db)
	if err != nil {
		handleError(c, err.Error())
	}
	message := fmt.Sprintf("item added: %s %s", name, category)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

func showItems(c echo.Context) error {
	var items model.Items
	var err error
	// Get a list of items
	items.Items, err = model.GetItems(db)
	if err != nil {
		handleError(c, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

func searchItem(c echo.Context) error {
	// Get a parameter
	var items model.Items
	var err error
	name := c.QueryParam("keyword")
	fmt.Println("name is : %s", name)
	// Search items in items.db
	items.Items, err = model.SearchItem(name, db)
	if err != nil {
		handleError(c, err.Error())
	}
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
	err := DBConnection()
	if err != nil {
		fmt.Println("database error: ", err, "\n")
	}
	defer DBClose()
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
	e.GET("/items", showItems)
	e.POST("/items", addItem)
	e.GET("/search", searchItem)
	e.GET("/image/:itemImg", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
