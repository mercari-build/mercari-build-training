package main

import (
	"crypto/sha256"
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
)

const ImgDir = "images"

type Response struct {
	Message string `json:"message"`
}

type ItemsData struct {
	Items []Item `json:"items"`
}

type Item struct {
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
	var items []Item
	var itemsData ItemsData

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

	//Create new item
	newItem := Item{
		Name:     name,
		Category: category,
		Image:    imageName,
	}

	//Print message
	message := fmt.Sprintf("item received: %s in %s category", newItem.Name, newItem.Category)
	res = Response{Message: message}

	if _, err := os.Stat("items.json"); err == nil {
		//Open the exist file and get itemsdata
		itemsData, err = readFile(c, "items.json")
		if err != nil {
			errMessage(c, err, http.StatusBadRequest, "Fail to read items.json")
		}
		items = itemsData.Items
		items = append(items, newItem)

	} else {
		if os.IsNotExist(err) {
			items = append(items, newItem)
		} else {
			errMessage(c, err, http.StatusBadRequest, "Somthing went wrong")
		}

	}

	itemFile, err := os.Create("items.json")
	if err != nil {
		errMessage(c, err, http.StatusBadRequest, "Fail to create items.json")
	}
	json.NewEncoder(itemFile).Encode(ItemsData{Items: items})
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
	itemsData, err := readFile(c, "items.json")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open items.json")
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
