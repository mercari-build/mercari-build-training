/* ************************************************************************** */
/*   main.go
/* ************************************************************************** */

package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const (
	ImgDir = "./images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CategoryID int    `json:"category_id"`
	Category   string `json:"category"`
	ImageName  string `json:"imagename"`
}

type Items struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func readAllDB(c echo.Context) error {
	items, err := getAllDB()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

func readKeywordItem(c echo.Context) error {
	keyword := c.Param("keyword")
	items, err := readKeywordDB(keyword)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, items)
}

func createItemDB(c echo.Context) error {
	item, err := createNewItem(c)
	if err != nil {
		if strings.Contains(err.Error(), "category_id") {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Internal Server Error",
		})
	}
	c.Logger().Infof("Receive item: %s", item.Name)
	inserted_id, err := addItemToDB(item)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to insert item into database",
		})
	}
	message := fmt.Sprintf("item received: %s and registered to database id:%d", item.Name, inserted_id)
	res := Response{Message: message}
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
		c.Logger().Errorf("Image not found: %s%s", imgPath, imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "/root/database/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

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
	e.GET("/items", readAllDB)
	e.POST("/items", createItemDB)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/search/:keyword", readKeywordItem)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
