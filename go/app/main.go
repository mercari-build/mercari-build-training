package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"io"
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
	img, err := c.FormFile("image")
	if err != nil {
		handleError(c, err.Error())
	}
	imageTitle := strings.Split(img.Filename, ".")[0]
	image_filename := hex.EncodeToString(getSHA256Binary(imageTitle)) + "." + strings.Split(img.Filename, ".")[1]
	item := model.Item{name, category, image_filename}
	c.Logger().Infof("Receive item: %s %s %s", name, category, image_filename)

	// Save a file
	newFile, err := os.Create("images/" + image_filename)
	imgFile, err := img.Open()
	_, err = io.Copy(newFile, imgFile)
	if err != nil {
		handleError(c, err.Error())
	}

	// Add item to db
	err = model.AddItem(item, db)
	if err != nil {
		handleError(c, err.Error())
	}
	message := fmt.Sprintf("item added: %s %s %s ", name, category, image_filename)
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
func searchItemById(c echo.Context) error {
	// Get Id from url
	id, err := uuid.Parse(c.Param("itemId"))

	// Get the item that corresponds to itemId
	item, err := model.SearchItemById(id, db)
	if err != nil {
		handleError(c, err.Error())
	}
	return c.JSON(http.StatusOK, item)
}

func searchItemByName(c echo.Context) error {
	// Get a parameter
	var items model.Items
	var err error
	name := c.QueryParam("keyword")
	fmt.Println("name is : %s", name)
	// Search items in items.db
	items.Items, err = model.SearchItemByName(name, db)
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

func getSHA256Binary(s string) []byte {
	r := sha256.Sum256([]byte(s))
	return r[:]
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
	e.GET("/items/:itemId", searchItemById)
	e.POST("/items", addItem)
	e.GET("/search", searchItemByName)
	e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
