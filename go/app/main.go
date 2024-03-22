package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

	_ "github.com/mattn/go-sqlite3"
)

const ImgDir = "../images"
const dbPath="../db/mercari.sqlite3"

type Response struct {
	Message string `json:"message"`
}

type ItemsData struct {
	Items []Item `json:"items"`
}

type Item struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category int    `json:"category"`
	Image    string `json:"image"`
}

type ItemDisplay struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image"`
}

type ItemsDisplay struct {
	Items []ItemDisplay `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func errMessage(c echo.Context, err error, status int, message string) error {
	errorMessage := fmt.Sprintf("%s:%s", message, err)
	return c.JSON(status, Response{Message: errorMessage})

}

func addItem(c echo.Context) error {
	var res Response
	

	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		errMessage(c, err, http.StatusBadRequest, "Unable to get image")
	}

	imageName, err := hashFile(c, image)
	if err != nil {
		errMessage(c, err, http.StatusBadRequest, "Fail to convert image to hash string")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}

	var categoryID int
	if err := db.QueryRow("SELECT id FROM categories WHERE name==?", category).Scan(&categoryID); err != nil {
		errMessage(c, err, http.StatusBadRequest, "Unable to get categoryID from categoryName")
	}

	
	message := fmt.Sprintf("item received: %s in %s category", name, category)
	res = Response{Message: message}

	// Save image file to ImgDir
	imageFile, err := image.Open()
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open the image")
	}
	defer imageFile.Close()

	// Create the file in ImgDir with the hashed name
	savedImagePath := path.Join(ImgDir, imageName)
	savedImageFile, err := os.Create(savedImagePath)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to create the image file")
	}
	defer savedImageFile.Close()

	// Copy the image data to the saved file
	if _, err := io.Copy(savedImageFile, imageFile); err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to save the image file")
	}

	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO items(name,category,image_name) VALUES (?,?,?)")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, categoryID, imageName)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}

	return c.JSON(http.StatusOK, res)

}

func hashFile(c echo.Context, image *multipart.FileHeader) (string, error) {
	//Create hash
	hash := sha256.New()

	//Open the image file
	imageFile, err := image.Open()
	if err != nil {
		errMessage(c, err, http.StatusBadRequest, "Unable to open the image")
	}
	defer imageFile.Close()

	// Read the file data and write it to the hash function
	if _, err := io.Copy(hash, imageFile); err != nil {
		errMessage(c, err, http.StatusBadRequest, "Unable to copy imagefile to hash")
	}

	// Get the final hash value
	hashValue := hash.Sum(nil)
	// Convert the byte slice to a hex-encoded string
	hashString := hex.EncodeToString(hashValue)
	imageName := hashString + ".jpg"

	return imageName, err
}

func getItem(c echo.Context) error {
	itemIDStr := c.Param("item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		errMessage(c, err, http.StatusBadRequest, "Unable to conveert item_id to int")
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT items.id,items.name,categories.name,items.image_name FROM items LEFT JOIN categories ON items.category=categories.id")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to execute SQL statement")
	}
	defer rows.Close()

	var itemsData ItemsDisplay
	for rows.Next() {
		var item ItemDisplay
		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
		if err != nil {
			return errMessage(c, err, http.StatusInternalServerError, "Unable to scan rows")
		}
		itemsData.Items = append(itemsData.Items, item)
	}

	if err := rows.Err(); err != nil {
		return errMessage(c, err, http.StatusInternalServerError, "Error iterating over rows")
	}
	return c.JSON(http.StatusOK, itemsData.Items[itemID-1])
}

func getItems(c echo.Context) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT items.id,items.name,categories.name,items.image_name FROM items LEFT JOIN categories ON items.category=categories.id")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to execute SQL statement")
	}
	defer rows.Close()

	var itemsData ItemsDisplay
	for rows.Next() {
		var item ItemDisplay
		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
		if err != nil {
			return errMessage(c, err, http.StatusInternalServerError, "Unable to scan rows")
		}
		itemsData.Items = append(itemsData.Items, item)
	}

	if err := rows.Err(); err != nil {
		return errMessage(c, err, http.StatusInternalServerError, "Error iterating over rows")
	}
	return c.JSON(http.StatusOK, itemsData)
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

func searchItem(c echo.Context) error {
	//Get Query param
	keyword := c.QueryParam("keyword")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT items.id,items.name,categories.name,items.image_name FROM items LEFT JOIN categories ON items.category==categories.name HAVING name LIKE CONCAT('%',?,'%')")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer stmt.Close()
	rows, err := stmt.Query(keyword)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to execute SQL statement")
	}
	defer rows.Close()

	var itemsData ItemsDisplay
	for rows.Next() {
		var item ItemDisplay
		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
		if err != nil {
			return errMessage(c, err, http.StatusInternalServerError, "Unable to scan rows")
		}
		itemsData.Items = append(itemsData.Items, item)
	}

	if err := rows.Err(); err != nil {
		return errMessage(c, err, http.StatusInternalServerError, "Error iterating over rows")
	}
	return c.JSON(http.StatusOK, itemsData)
}

func addCategory(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	//Print message
	message := fmt.Sprintf("category received: %s", name)
	res := Response{Message: message}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO categories(name) VALUES (?)")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer stmt.Close()
	_, err = stmt.Exec(name)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to execute sql command")
	}

	return c.JSON(http.StatusOK, res)

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
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:item_id", getItem)
	e.GET("/search", searchItem)
	e.POST("/categories", addCategory)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
