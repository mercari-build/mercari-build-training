package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type ItemsData struct {
	Item []Item `json:"items"`
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

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image := c.FormValue("image")

	c.Logger().Infof("Receive item: %s", name)
	c.Logger().Infof("Receive category: %s", category)
	c.Logger().Infof("Receive image: %s", image)

	hash := sha256.New()
	// Write the image data to the hash function
	hash.Write([]byte(strings.Split(image, ".")[0]))

	// Get the final hash value
	hashValue := hash.Sum(nil)
	// Convert the byte slice to a hex-encoded string
	hashString := hex.EncodeToString(hashValue)
	image = hashString + ".jpg"

	var res Response
	var itemslice []Item
	// Create a new item with the next ID
	newItem := Item{
		Name:     name,
		Category: category,
		Image:    image,
	}

	if _, err := os.Stat("items.json"); err == nil {
		file1, err := os.Open("items.json") //すでにあるファイルを開く
		if err != nil {
			log.Fatal(err)
		}

		defer file1.Close()
		jsonData, err := ioutil.ReadAll(file1)
		if err != nil {
			fmt.Println("JSONデータを読み込めません", err)
		}

		json.Unmarshal(jsonData, &itemslice)

		//Print message
		message := fmt.Sprintf("item received: %s in %s category", newItem.Name, newItem.Category)
		res = Response{Message: message}

		// Append the new item to the slice
		itemslice = append(itemslice, newItem)

	} else {
		itemslice = append(itemslice, newItem)

	}

	file2, err := os.Create("items.json") // fileはos.File型
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(file2).Encode(itemslice)

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	file1, err := os.Open("items.json") //すでにあるファイルを開く
	if err != nil {
		log.Fatal(err)
	}
	defer file1.Close()
	jsonData, err := ioutil.ReadAll(file1)
	if err != nil {
		fmt.Println("JSONデータを読み込めません", err)
	}
	var itemslice []Item
	json.Unmarshal(jsonData, &itemslice)
	fmt.Println(itemslice)

	return c.JSON(http.StatusOK, itemslice)
}

func getItem(c echo.Context) error {
	itemIDStr := c.Param("item_id")

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		// エラーが発生した場合の処理
		fmt.Println("item_idを整数に変換できません:", err)
	}

	fmt.Println(itemID)

	file1, err := os.Open("items.json") //すでにあるファイルを開く
	if err != nil {
		log.Fatal(err)
	}
	defer file1.Close()
	jsonData, err := ioutil.ReadAll(file1)
	if err != nil {
		fmt.Println("JSONデータを読み込めません", err)
	}
	// var itemslice []ItemsData
	itemsData := ItemsData{}
	json.Unmarshal(jsonData, &itemsData)
	fmt.Println(itemsData)

	// res := itemsResponse{Item: itemslice[itemID-1]}
	return c.JSON(http.StatusOK, itemsData.Item[itemID-1])
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
