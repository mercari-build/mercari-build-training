/* ************************************************************************** */
/*   main.go
/* ************************************************************************** */

package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CategoryID int    `json:"category_id"`
	ImageName  string `json:"imagename"`
}

type ShowItem struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"imagename"`
}

type Items struct {
	Items []Item `json:"items"`
}

type ShowItems struct {
	Items []ShowItem `json:"showitems"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func readAll(c echo.Context) error {
	items, err := getAllItems()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

func readAllDB(c echo.Context) error {
	items, err := getAllDB()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

func readOne(c echo.Context) error {
	items, err := getAllItems()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "InValid item ID"})
	}
	if id < 0 || id >= len(items.Items) {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "InValid index"})
	}
	item := items.Items[id]
	return c.JSON(http.StatusOK, item)
}

func readKeywordItem(c echo.Context) error {
	keyword := c.Param("keyword")
	items, err := readKeywordDB(keyword)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, items)
}

func createItem(c echo.Context) error {

	item, err := createNewItem(c)
	if err != nil {
		return err
	}
	c.Logger().Infof("Receive item: %s", item.Name)
	if err := addItemToFile(item); err != nil {
		return err
	}
	message := fmt.Sprintf("item received: %s", item.Name)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

func createItemDB(c echo.Context) error {
	item, err := createNewItem(c)
	if err != nil {
		return err
	}
	c.Logger().Infof("Receive item: %s", item.Name)
	inserted_id, err := addItemToDB(item)
	if err != nil {
		return err
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
		c.Logger().Errorf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "../../db/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	//print items table
	rows, err := db.Query("SELECT * FROM items")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var id int
	var name string
	var categoryId int
	var image string
	fmt.Println("Items:")
	for rows.Next() {
		err = rows.Scan(&id, &name, &categoryId, &image)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d | %s | %d | %s\n", id, name, categoryId, image)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	//print categories table
	rows, err = db.Query("SELECT * FROM categories")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var categoryName string
	fmt.Println("\nCategories:")
	for rows.Next() {
		err = rows.Scan(&id, &categoryName)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d | %s\n", id, categoryName)
	}

	err = rows.Err()
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
	e.GET("/items", readAll)
	e.GET("/itemsdb", readAllDB)
	e.GET("/items/:id", readOne)
	e.POST("/items", createItem)
	e.POST("/itemsdb", createItemDB)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/search/:keyword", readKeywordItem)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
