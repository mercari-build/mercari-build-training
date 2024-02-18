package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

type Item struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

type ItemResponse struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

var items []Item

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func saveImage(fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// SHA-256 ハッシュを計算
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)
	hashedFileName := hex.EncodeToString(hashInBytes) + ".jpg"

	// ファイルポインタをリセット
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	// 画像ファイルを保存
	dst, err := os.Create(filepath.Join("images", hashedFileName))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return hashedFileName, nil
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Received item: %s, Category: %s", name, category)

	fileHeader, err := c.FormFile("image")
	var imageName string
	if err == nil {
		imageName, err = saveImage(fileHeader)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
	} else {
		imageName = "default.jpg"
	}

	newItem := Item{
		ID:        strconv.Itoa(len(items) + 1),
		Name:      name,
		Category:  category,
		ImageName: imageName,
	}
	items = append(items, newItem)

	// 出力にIDが不要なためItemResponseスライスに変換
	var responseItems []ItemResponse
	for _, item := range items {
		responseItem := ItemResponse{
			Name:      item.Name,
			Category:  item.Category,
			ImageName: item.ImageName,
		}
		responseItems = append(responseItems, responseItem)
	}

	// ItemResponseスライスをレスポンスとして返す
	return c.JSON(http.StatusOK, echo.Map{"items": responseItems})
}

func getItem(c echo.Context) error {
	itemID := c.Param("id") // URLパラメータからitem_idを取得

	// アイテムのリストを検索して対応するアイテムを見つける
	for _, item := range items {
		if item.ID == itemID {
			response := ItemResponse{
				Name:      item.Name,
				Category:  item.Category,
				ImageName: item.ImageName,
			}
			return c.JSON(http.StatusOK, response)
		}
	}

	return c.JSON(http.StatusNotFound, echo.Map{"message": "Item not found"})
}

func getItems(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"items": items})
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
	if _, err := os.Stat("images"); os.IsNotExist(err) {
		os.Mkdir("images", os.ModePerm)
	}
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
	e.GET("/items/:id", getItem)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
