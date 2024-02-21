package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Item struct {
	ID string `json: "id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	ImageName string `json:"image_name"`
}


type Items struct {
	Items []Item `json:"items"`
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

func addItem(c echo.Context) error {	
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")

	// 画像ファイル取得
	imageFile, err := c.FormFile("image")

	if err != nil {
		res := Response{Message: "Failed to get image file"}
		return c.JSON(http.StatusBadRequest, res)
	}
	src, err := imageFile.Open()
	print(imageFile.Filename)
	if err != nil {
		res := Response{Message: "Failed to open image file"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer src.Close()

    // 画像ファイルの内容を読み込む
    imgData, err := io.ReadAll(src)
    if err != nil {
        res := Response{Message: "Failed to read image file"}
        return c.JSON(http.StatusInternalServerError, res)
    }

    // SHA256ハッシュを計算
    hasher := sha256.New()
    hasher.Write(imgData) // 画像データをハッシュ関数に渡す
    hash := hex.EncodeToString(hasher.Sum(nil))

    // newItemを定義
    newItem := Item {
        Name: name,
        Category: category,
		// ハッシュ値を基にファイル名を生成
        ImageName: hash + ".jpg", 
    }

	jsonFile, err := os.Open("items.json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	var items Items
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	json.Unmarshal(jsonData, &items)

	items.Items = append(items.Items, newItem)

	jsonFile, err = os.Create("items.json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	err = encoder.Encode(items)
	if err != nil {
		return err
	}

	c.Logger().Infof("Name: %s, Category: %s, ImageName: %s", name, category, newItem.ImageName)
	return c.JSON(http.StatusOK, newItem)
}

func getAllItem(c echo.Context) error {
	//JSONファイルを開く
	jsonFile, err :=os.Open("items.json")
    if err != nil {
        // ファイルが開けない場合はエラーレスポンスを返す
        return c.JSON(http.StatusInternalServerError, Response{Message: "Failed to open items.json"})
    }
    defer jsonFile.Close()
    // ファイルの内容を読み込む
    jsonData, err := io.ReadAll(jsonFile)
    if err != nil {
        // 読み込みに失敗した場合はエラーレスポンスを返す
        return c.JSON(http.StatusInternalServerError, Response{Message: "Failed to read items.json"})
    }

    // 読み込んだJSONデータをItems構造体にデコードする
    var items Items
    err = json.Unmarshal(jsonData, &items)
    if err != nil {
        // デコードに失敗した場合はエラーレスポンスを返す
        return c.JSON(http.StatusInternalServerError, Response{Message: "Failed to decode items.json"})
    }

    // デコードしたデータをレスポンスとして返す
    return c.JSON(http.StatusOK, items)
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

func getItemById(c echo.Context) error {
    // paramsからidを取得
    idStr := c.Param("id")
	_, err := strconv.Atoi(idStr)
    if err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid ID")
    }

    // items.jsonファイルを開く
    jsonFile, err := os.Open("items.json")
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Failed to open items.json")
    }
    defer jsonFile.Close()

    // ファイルの内容を読み込む
    jsonData, err := io.ReadAll(jsonFile)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Failed to read items.json")
    }

    // 読み込んだJSONデータをItems構造体にデコードする
    var items Items
    err = json.Unmarshal(jsonData, &items)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Failed to decode items.json")
    }

    // 指定されたidに一致するアイテムを探す
    for _, item := range items.Items {
        if item.ID == idStr {
            return c.JSON(http.StatusOK, item)
        }
    }

    // アイテムが見つからない場合は404エラーを返す
    return c.JSON(http.StatusNotFound, "Item not found")
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
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items", getAllItem)
	e.GET("/items/:id", getItemById)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
