package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "modernc.org/sqlite"
)

const (
	ImgDir = "images"
	dbPath = "../../sqlite3/mercari.sqlite3"
)

type Response struct {
	Message string `json:"message"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image_name"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	// get form file
	file, err := c.FormFile("image")
	if err != nil {
		log.Print("画像ファイルの受け取りに失敗", err)
		return err
	}

	c.Logger().Infof("Receive item: %s", name)
	c.Logger().Infof("Receive item: %s", category)
	c.Logger().Infof("Receive item: %s", file)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	// open image file
	src, err := file.Open()
	if err != nil {
		log.Print("画像ファイルの読み取りに失敗", err)
		return err
	}
	defer src.Close()

	// create hash
	h := sha256.New()
	if _, err = io.Copy(h, src); err != nil {
		log.Print("hash生成に失敗", err)
		return err
	}
	imageName := fmt.Sprintf("%x.jpg", h.Sum(nil))

	// select directory path for new image file
	filePath := filepath.Join("images/", imageName)
	// destination to store image file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Print("新規画像ファイル作成に失敗", err)
		return err
	}
	defer dst.Close()

	// copy image to created image file
	if _, err = io.Copy(dst, src); err != nil {
		log.Print("新規画像の保存に失敗", err)
		return err
	}
	// connect to db
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Print("db接続に失敗")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO ITEMS (name, category, image_name) VALUES (?,?,?)")
	if err != nil {
		log.Print("INSERTクエリ失敗")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	_, err = stmt.Exec(name, category, imageName)
	if err != nil {
		log.Print("INSERT失敗")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)

}

func getItems(c echo.Context) error {
	// connect to db
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Print("db接続に失敗")
		return err
	}
	defer db.Close()
	// get data from db
	rows, err := db.Query("SELECT name, category, image_name FROM items")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	items := Items{}

	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Name, &item.Category, &item.Image)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		items.Items = append(items.Items, item)
	}
	return c.JSON(http.StatusOK, items)

	// jsonFile, err := os.Open("items.json")
	// if err != nil {
	// 	log.Print("JSONファイルを開けません", err)
	// 	return err
	// }
	// defer jsonFile.Close()
	// itemsData := Items{}
	// err = json.NewDecoder(jsonFile).Decode(&itemsData)
	// if err != nil {
	// 	log.Print("JSONファイルからの変換に失敗", err)
	// 	return err
	// }

	// return c.JSON(http.StatusOK, itemsData)
}

func getItemById(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	file, err := os.Open("items.json")
	if err != nil {
		c.Logger().Errorf("Error opening file: %s", err)
		res := Response{Message: "Error opening file"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer file.Close()

	itemsData := Items{}
	err = json.NewDecoder(file).Decode(&itemsData)
	if err != nil {
		log.Print("JSONファイルからの変換に失敗", err)
		return err
	}
	return c.JSON(http.StatusOK, itemsData.Items[id-1])

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
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/items/:id", getItemById)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
