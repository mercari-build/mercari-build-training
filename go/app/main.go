package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"database/sql"
	// "sqlite3"

	// "./././db/items.db"

	_ "github.com/mattn/go-sqlite3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)



const (
	ImgDir = "image"
)

type Response struct {
	Message string `json:"items"`
}

type Items struct {
	Items Data `json:"items"`
}

type Data struct {
	Name     string `json:"name"`
	Category string `json:"string"`
}

func root(c echo.Context) error {
	res := "error"
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive item: %s", name)
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s", category)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	// data := Items{Items: Data{name, category}}
	// stmt, err := db.Prepare("INSERT INTO items(name, category) values(?,?)")
	// data, err := stmt.Exec(name,category)

	//write items.json
	// f, err := os.Create("./items.json")
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()

	// err = json.NewEncoder(f).Encode(data)
	// if err != nil {
	// 	return err
	// }

	return c.JSON(http.StatusOK, res)
}

func showItem(c echo.Context) error {
	// JSONファイル読み込み
    bytes, err := ioutil.ReadFile("./items.json")
    if err != nil {
        log.Fatal(err)
    }
    // JSONデコード
    var res Items
    if err := json.Unmarshal(bytes, &res); err != nil {
        log.Fatal(err)
    }
    // デコードしたデータを表示
	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("itemImg"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		// res := Response{Items: "Image path does not end with .jpg"}
		res := "Image path does not end with .jpg"
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	// db接続
	db, err := sql.Open("sqlite3", "db/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
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
	e.GET("/items", showItem)
	e.POST("/items", addItem)
	e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
