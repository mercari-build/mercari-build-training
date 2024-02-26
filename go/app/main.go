package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
	DBPath = "../db/mercari.sqlite3"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name,omitempty"`
}

type Items struct {
	Items []Item `json:"items"`
}

// scanRowsToItems is a method for type Items.
// It scans *sql.rows and turns it into type Items that has item name and category name.
func (items *Items) ScanRowsToItems(rows *sql.Rows) error {
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Name, &item.Category)
		if err != nil {
			return err
		}
		items.Items = append(items.Items, item)
	}
	return nil
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!!"}
	return c.JSON(http.StatusOK, res)
}

// addItem processes form data and saves item information.
func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		c.Logger().Errorf("Error while retrieving image: %s", err)
		res := Response{Message: "Error while retrieving image"}
		return echo.NewHTTPError(http.StatusBadRequest, res)
	}

	c.Logger().Infof("Receive item: %s, Category: %s", name, category)

	fileName, err := saveImage(image)
	if err != nil {
		c.Logger().Errorf("Error while hashing and saving image: %s", err)
		res := Response{Message: "Error while hashing and saving image"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	if err := saveItem(name, category, fileName); err != nil {
		c.Logger().Errorf("Error while saving item information: %s", err)
		res := Response{Message: "Error while saving item information"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

// saveItem writes the item information into the database.
func saveItem(name, category, fileName string) error {

	dbCon, err := connectDB(DBPath)
	if err != nil {
		return err
	}
	defer dbCon.Close()

	// Transaction starts.
	tx, err := dbCon.Begin()
	if err != nil {
		err = fmt.Errorf("error while beginning transaction: %s", err)
		return err
	}

	categoryId, err := searchCategoryId(category, tx)
	if err != nil {
		tx.Rollback()
		err = fmt.Errorf("error while searching for category id: %s", err)
		return err
	}

	const insertItem = "INSERT INTO items (name, image_name, category_id) VALUES (?, ?, ?)"
	_, err = tx.Exec(insertItem, name, fileName, categoryId)
	if err != nil {
		tx.Rollback()
		err = fmt.Errorf("error while adding a new item: %s", err)
		return err
	} else {
		if err := tx.Commit(); err != nil {
			err = fmt.Errorf("error while commiting transaction: %s", err)
			return err
		}
	}

	return nil
}

// searchCategoryId look for category ID, if it doesn't exist, register a new category.
func searchCategoryId(category string, tx *sql.Tx) (int, error) {
	var categoryId int

	const searchForKey = "SELECT id FROM categories WHERE name = ?"
	for {
		rows, err := tx.Query(searchForKey, category)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&categoryId)
			if err != nil {
				return 0, err
			}
			break
		} else {
			makeNewCategory := "INSERT INTO categories (name) values (?)"
			_, err := tx.Exec(makeNewCategory, category)
			if err != nil {
				return 0, err
			}
		}

	}
	return categoryId, nil
}

// saveImage hashes the image, saves it, and returns its file name.
func saveImage(image *multipart.FileHeader) (string, error) {

	img, err := image.Open()
	if err != nil {
		return "", err
	}
	source, err := io.ReadAll(img)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(source)

	err = os.MkdirAll("./images", 0750)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%x.jpg", hash)
	imagePath := fmt.Sprintf("./images/%s", fileName)

	_, err = os.Create(imagePath)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(imagePath, source, 0644)
	if err != nil {
		return "", err
	}

	return fileName, err
}

// getItems gets all the item information.
func getItems(c echo.Context) error {
	items, err := readItems()
	if err != nil {
		c.Logger().Errorf("Error while reading item information: %s", err)
		res := Response{Message: "Error while reading item information"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, items)
}

// readItems reads database and returns all the item information.
func readItems() (Items, error) {

	dbCon, err := connectDB(DBPath)
	if err != nil {
		return Items{}, err
	}
	defer dbCon.Close()

	const selectAllItems = "SELECT items.name, categories.name FROM items JOIN categories ON items.category_id = categories.id"
	rows, err := dbCon.Query(selectAllItems)
	if err != nil {
		return Items{}, err
	}

	items := new(Items)
	err = items.ScanRowsToItems(rows)
	if err != nil {
		return Items{}, err
	}

	return *items, nil
}

func searchItems(c echo.Context) error {
	keyword := c.QueryParam("keyword")
	key := "%" + keyword + "%"

	dbCon, err := connectDB(DBPath)
	if err != nil {
		c.Logger().Errorf("Error while connecting to database: %s", err)
		res := Response{Message: "Error while connecting to database"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}
	defer dbCon.Close()

	const searchWithKey = "SELECT items.name, categories.name FROM items JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?"
	rows, err := dbCon.Query(searchWithKey, key)
	if err != nil {
		c.Logger().Errorf("Error while searching with keyword: %s", err)
		res := Response{Message: "Error while searching with keyword"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	resultItems := new(Items)
	err = resultItems.ScanRowsToItems(rows)
	if err != nil {
		c.Logger().Errorf("Error while scanning rows to items: %s", err)
		res := Response{Message: "Error while scanning rows to items"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, resultItems)
}

// getImg gets the designated image by file name.
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

// getInfo gets detailed information of the designeted item by id.
func getInfoById(c echo.Context) error {
	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Logger().Errorf("Invalid ID: %s", err)
		res := Response{Message: "Invalid ID"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	dbCon, err := connectDB(DBPath)
	if err != nil {
		c.Logger().Errorf("Error while connecting to database: %s", err)
		res := Response{Message: "DError while connecting to database"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}
	defer dbCon.Close()

	var item Item
	selectById := "SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id WHERE items.id = ?"
	rows := dbCon.QueryRow(selectById, itemId)
	err = rows.Scan(&item.Name, &item.Category, &item.ImageName)
	if err != nil {
		c.Logger().Errorf("Error while searching item with ID: %s", err)
		res := Response{Message: "Error while searching item with ID"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, item)
}

// 　connectDB opens database connection.
func connectDB(dbPath string) (*sql.DB, error) {
	dbCon, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return dbCon, nil
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
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:id", getInfoById)
	e.GET("/search", searchItems)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
