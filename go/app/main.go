package main

import (
	"fmt"
	"net/http"
	// "path/filepath"
	"os"
	"path"
	"strings"
	"encoding/json"
	"io/ioutil"
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)
const dbPath = "path/to/your/database.db"

type Item struct {
    Name         string `json:"name"`
    Category     string `json:"category"`
    ImageFilename string `json:"image_filename"`
}

type Items struct {
    Items []*Item `json:"items"`
}

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}


func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
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
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func readItemsFromFile(filename string) ([]Item, error) {
	// ファイルをオープン
	f, err := os.Open(filename)
	if err != nil {
		log.Errorf("Error opening file: %s", err)
		return nil, err
	}
	defer f.Close()

	// ファイルの内容を読み取り
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Errorf("Error reading file: %s", err)
		return nil, err
	}

	// JSONデコード
	var items []Item
	err = json.Unmarshal(bytes, &items)
	if err != nil {
		log.Errorf("Error decoding JSON: %s", err)
		return nil, err
	}

	return items, nil
}

func getItems(c echo.Context) error {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }
    defer db.Close()

    rows, err := db.Query("SELECT items.name, categories.name AS category FROM items INNER JOIN categories ON items.category_id = categories.id")
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }
    defer rows.Close()

    var items []map[string]string
    for rows.Next() {
        var itemName, categoryName string
        if err := rows.Scan(&itemName, &categoryName); err != nil {
            return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
        }
        item := map[string]string{
            "name":     itemName,
            "category": categoryName,
        }
        items = append(items, item)
    }

    return c.JSON(http.StatusOK, items)
}

func addItem(c echo.Context) error {
    // Parse request body to get item details
	name := c.FormValue("name")
	c.Logger().Infof("Reeive item:%s",name)

	message := fmt.Sprintf("item received: %s",name)

	db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }
    defer db.Close()

	stmt,err := db.Prepare("INSERT INTO items (name, category, image_name)VALUES(?,?,?)")
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }
	defer stmt.Close()
	_, err = stmt.Exec(name, "unknown", "default.jpg")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

func searchItem(c echo.Context) error {
    // Parse request body to get item details
	keyword := c.QueryParam("keyword")

	db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }
    defer db.Close()

	rows,err := db.Query("SELECT name, category, image_name FROM items WHERE name LIKE ?", "%"+keyword+"%")
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }
	defer rows.Close()

	
	items := Items{Items: []*Item{}}
	for rows.Next(){
		var item Item
		err := rows.Scan(&item.Name, &item.Category, &item.ImageFilename)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		items.Items = append(items.Items, &item)
	}
	return c.JSON(http.StatusOK,items)
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
	e.GET("/search",searchItem)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
