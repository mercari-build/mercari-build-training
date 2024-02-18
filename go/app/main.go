package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
)

const (
	ImgDir    = "images"
	ItemsFile = "items.json"
	ImgExtension = ".jpg"
)

type Item struct {
	ID string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	ImageName string `json:"imageName"`
}

type Items struct {
	Items []Item `json:"items"`
	MaxID int `json:"maxId"`
}

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func saveImage(src multipart.File) (string, error) {
    // ファイルポインタをリセット
    _, err := src.Seek(0, io.SeekStart)
    if err != nil {
        return "", err
    }

    // SHA-256 ハッシュを計算
    hash := sha256.New()
    if _, err := io.Copy(hash, src); err != nil {
        return "", err
    }
    hashInBytes := hash.Sum(nil)
    imageName := hex.EncodeToString(hashInBytes) + ImgExtension

    // 画像ファイルを保存
    dst, err := os.Create(path.Join(ImgDir, imageName))
    if err != nil {
        return "", err
    }
    defer dst.Close()

    // ファイルの内容を再度読み込み
    _, err = src.Seek(0, io.SeekStart)
	if err != nil {
	    return "", err
	}
    if _, err := io.Copy(dst, src); err != nil {
        return "", err
    }
    return imageName, nil
}

func loadItemsFromFile(filename string) (Items, error) {
    var items Items

    file, err := os.ReadFile(filename)
    if err != nil {
        if os.IsNotExist(err) {
            return Items{Items: []Item{}, MaxID: 0}, nil
        }
        return items, fmt.Errorf("failed to read file '%s': %w", filename, err)
    }

	if len(file) == 0 {
        return Items{Items: []Item{}, MaxID: 0}, nil
    }

    err = json.Unmarshal(file, &items)
    if err != nil {
        return Items{Items: []Item{}, MaxID: 0}, fmt.Errorf("failed to unmarshal items: %w", err)
    }

    // MaxIDを更新
    maxID := 0
    for _, item := range items.Items {
        id, err := strconv.Atoi(item.ID)
        if err != nil {
            return items, fmt.Errorf("invalid item ID '%s': %w", item.ID, err)
        }
        if id > maxID {
            maxID = id
        }
    }
    items.MaxID = maxID

    return items, nil
}

func saveItemsToFile(filename string, items Items) error {
    itemsData, err := json.Marshal(items)
    if err != nil {
        return err
    }
    return os.WriteFile(filename, itemsData, 0644)
}

func addItem(c echo.Context) error {
    // Get form data for name and category
    name := c.FormValue("name")
    category := c.FormValue("category")
    c.Logger().Infof("Received item: %s, Category: %s", name, category)

    // Initialize imageName as empty
    var imageName string

    // Attempt to get the image file from the form
    file, err := c.FormFile("image")
    if err == nil {
        src, err := file.Open()
        if err != nil {
            return err
        }
        defer src.Close()

		buffer := make([]byte, 512) // MIMEタイプを検出するために最初の512バイトを読み込む
    	_, err = src.Read(buffer)
    	if err != nil {
        	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read file for MIME type detection")
    	}

		// ファイルポインタをリセット
    	_, err = src.Seek(0, io.SeekStart)
    	if err != nil {
        	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset file pointer")
    	}

    	mimeType := http.DetectContentType(buffer)
    	if !strings.HasPrefix(mimeType, "image/") {
        	return echo.NewHTTPError(http.StatusBadRequest, "The uploaded file is not an image")
    	}

        imageName, err = saveImage(src)
        if err != nil {
            c.Logger().Errorf("Failed to save image: %v", err)
            return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save image")
        }
    } else {
        // Log the absence of an image file or set a default imageName if necessary
        c.Logger().Info("No image file provided")
        // imageName can be set to a default or left empty
    }

    // Proceed with adding the item to the list
    items, err := loadItemsFromFile(ItemsFile)
    if err != nil {
        c.Logger().Errorf("Failed to load items: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load items")
    }

    // IDを更新
    items.MaxID += 1
    newItem := Item{ID: strconv.Itoa(items.MaxID), Name: name, Category: category, ImageName: imageName}
    items.Items = append(items.Items, newItem)

    // Save updated items back to JSON file
    err = saveItemsToFile(ItemsFile, items)
    if err != nil {
        c.Logger().Errorf("Failed to save items to file: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save items")
    }
	return c.JSON(http.StatusOK, map[string]interface{}{
        "items": items.Items,
    })
}

func getItem(c echo.Context) error {
    itemID := c.Param("id")

    items, err := loadItemsFromFile(ItemsFile)
    if err != nil {
        c.Logger().Errorf("Failed to load items: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load items")
    }

    for _, item := range items.Items {
        if item.ID == itemID {
            return c.JSON(http.StatusOK, item)
        }
    }

    return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Item with ID %s not found", itemID))
}

func getItems(c echo.Context) error {
    items, err := loadItemsFromFile(ItemsFile)
    if err != nil {
        c.Logger().Errorf("Failed to load items: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Error reading items file")
    }
    return c.JSON(http.StatusOK, map[string]interface{}{
        "items": items.Items,
    })
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ImgExtension) {
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

	// ログレベルをDEBUGに設定
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
	e.GET("/items/:id", getItem)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
