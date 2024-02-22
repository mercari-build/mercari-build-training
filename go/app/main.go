package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

const ImgDir = "images"

type Response struct {
	Message string `json:"message"`
}

type ItemsData struct {
	Items []Item `json:"items"`
}

type Item struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func errMessage(c echo.Context, err error, status int, message string) error {
	errorMessage := fmt.Sprintf("%s:%s", message, err)
	return c.JSON(status, Response{Message: errorMessage})

}

func readFile(c echo.Context, filePath string) (ItemsData, error) {
	var items ItemsData
	file, err := os.Open(filePath)
	if err != nil {
		return items, errMessage(c, err, http.StatusBadRequest, "Unable to open the file")
	}
	defer file.Close()
	jsonData, err := ioutil.ReadAll(file)
	if err != nil {
		return items, errMessage(c, err, http.StatusBadRequest, "Unable to read json data")
	}
	err = json.Unmarshal(jsonData, &items)
	if err != nil {
		return items, errMessage(c, err, http.StatusBadRequest, "Unable to unmarshal")
	}
	return items, nil
}

func addItem(c echo.Context) error {
	var res Response
	// Get form data
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

	//Print message
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

	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO items(name,category,image_name) VALUES (?,?,?)")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, category, imageName)
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
	itemsData, err := readFile(c, "items.json")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open items.json")
	}
	return c.JSON(http.StatusOK, itemsData.Items[itemID-1])
}

func getItems(c echo.Context) error {
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT * FROM items")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to execute SQL statement")
	}
	defer rows.Close()

	var itemsData ItemsData
	for rows.Next() {
		var item Item
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

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
