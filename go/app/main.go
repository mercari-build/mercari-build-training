package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3" // SQLite3 driver
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	return echo.NewHTTPError(http.StatusOK, "Hello, World!")
}

type Item struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

type Items struct {
	Items []Item `json:"items"`
}

type ServerImpl struct {
	db *sql.DB
}

func (s ServerImpl) addItem() echo.HandlerFunc {
	return func(c echo.Context) error {
		// リクエストボディからデータを取得
		name := c.FormValue("name")
		category := c.FormValue("category")

		// 画像ファイルを取得
		imageFile, err := c.FormFile("image")
		if err != nil {
			c.Logger().Errorf("Failed to get image file in addItem: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "failed to get image file")
		}
		src, err := imageFile.Open()
		if err != nil {
			c.Logger().Errorf("Failed to open image file in addItem: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to open image file")
		}
		defer src.Close()

		// 画像ファイルをハッシュ化
		hash := sha256.New()
		if _, err := io.Copy(hash, src); err != nil {
			c.Logger().Errorf("Failed to hash image file in addItem: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash image file")
		}
		hashedImageName := fmt.Sprintf("%x.jpeg", hash.Sum(nil))

		// 画像ファイルを保存
		dst, err := os.Create(fmt.Sprintf("images/%s", hashedImageName))
		if err != nil {
			c.Logger().Errorf("Failed to create image file in addItem: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create image file")
		}
		defer dst.Close()
		src.Seek(0, 0) // ファイルポインタを先頭に戻す
		//srcからdstへ内容をコピー
		if _, err := io.Copy(dst, src); err != nil {
			c.Logger().Errorf("Failed to copy image file in addItem: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to copy image file")
		}

		// DBへの保存
		if err := s.addItemToDB(name, category, hashedImageName); err != nil {
			c.Logger().Errorf("Failed to add item to DB in addItem: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to add item to DB")
		}

		c.Logger().Infof("Receive item: %s", name)

		message := fmt.Sprintf("item received: name=%s,category=%s,images=%s", name, category, hashedImageName)
		res := Response{Message: message}

		return c.JSON(http.StatusOK, res)
	}
}

func (s ServerImpl) addItemToDB(name, category, imageName string) error {
	// トランザクションを開始
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction in addItemToDB: %w", err)
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRow("SELECT id FROM categories WHERE name = ?", category).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // カテゴリが存在しない場合
			stmt1, err := tx.Prepare("INSERT INTO categories (name) VALUES (?)")
			if err != nil {
				return fmt.Errorf("failed to prepare SQL statement1 in addItemToDB: %w", err)
			}
			defer stmt1.Close()

			result, err := stmt1.Exec(category)
			if err != nil {
				return fmt.Errorf("failed to execute SQL statement1 in addItemToDB: %w", err)
			}
			// 新しく挿入された行のIDを取得
			id, err = result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get last insert ID in addItemToDB: %w", err)
			}
		} else {
			return fmt.Errorf("failed to select id from categories in addItemToDB: %w", err)
		}
	}

	// itemsテーブルに商品を追加
	stmt2, err := tx.Prepare("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement2 in addItemToDB: %w", err)
	}
	defer stmt2.Close()
	if _, err := stmt2.Exec(name, id, imageName); err != nil {
		return fmt.Errorf("failed to execute SQL statement2 in addItemToDB: %w", err)
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction in addItemToDB: %w", err)
	}

	return nil
}

func (s ServerImpl) getAllItems() echo.HandlerFunc {
	return func(c echo.Context) error {
		// itemsテーブルとcategoriesテーブルをJOINして全てのアイテムを取得
		rows, err := s.db.Query("SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id")
		if err != nil {
			c.Logger().Errorf("Failed to search items from DB in getAllItems: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search items from DB")
		}
		defer rows.Close()

		var allItems Items
		for rows.Next() {
			var item Item
			if err := rows.Scan(&item.Name, &item.Category, &item.ImageName); err != nil {
				c.Logger().Errorf("Failed to scan items from DB in getAllItems: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to scan items from DB")
			}
			allItems.Items = append(allItems.Items, item)
		}
		return c.JSON(http.StatusOK, allItems)
	}
}

func (s ServerImpl) getItemsByKeyword() echo.HandlerFunc {
	return func(c echo.Context) error {
		// クエリパラメータからキーワードを取得
		keyword := c.QueryParam("keyword")

		// DBから名前にキーワードを含む商品一覧を返す
		rows, err := s.db.Query(`
			SELECT items.name, categories.name, items.image_name 
			FROM items JOIN categories ON items.category_id = categories.id 
			WHERE items.name LIKE '%' || ? || '%'`, keyword)
		if err != nil {
			c.Logger().Errorf("Failed to search items from DB in getItemsByKeyword: %v,keyword: %v", err, keyword)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search items from DB")
		}
		defer rows.Close()

		var keywordItems Items
		for rows.Next() {
			var item Item
			if err := rows.Scan(&item.Name, &item.Category, &item.ImageName); err != nil {
				c.Logger().Errorf("Failed to scan items from DB in getItemsByKeyword: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to scan items from DB")
			}
			keywordItems.Items = append(keywordItems.Items, item)
		}
		return c.JSON(http.StatusOK, keywordItems)
	}
}

func (s ServerImpl) getItemById() echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Logger().Errorf("Failed to convert id to int in getItemById: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "failed to convert id to int")
		}

		// DBからIDに対応する商品を取得
		row := s.db.QueryRow(`
			SELECT items.name, categories.name, items.image_name 
			FROM items JOIN categories ON items.category_id = categories.id 
			WHERE items.id = ?`, id)

		var item Item
		if err := row.Scan(&item.Name, &item.Category, &item.ImageName); err != nil {
			if errors.Is(err, sql.ErrNoRows) { // IDに対応する商品がない場合
				c.Logger().Errorf("Item not found in DB: id=%d", id)
				return echo.NewHTTPError(http.StatusNotFound, "item not found")
			}
			c.Logger().Errorf("Failed to search item from DB in getItemById: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search item from DB")
		}
		return c.JSON(http.StatusOK, item)
	}
}

func LoadItemsFromJSON() (*Items, error) {
	jsonFile, err := os.Open("items.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open items.json: %w", err)
	}
	defer jsonFile.Close()

	var allItems Items
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&allItems); err != nil {
		return nil, fmt.Errorf("failed to decode items.json: %w", err)
	}
	return &allItems, nil
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		c.Logger().Error("Image path does not end with .jpg")
		return echo.NewHTTPError(http.StatusInternalServerError, "Image path does not end with .jpg")
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Infof("Image not found: %s", imgPath)
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

	// DBへの接続
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		e.Logger.Infof("Failed to open DB: %v", err)
	}
	defer db.Close()

	serverImpl := ServerImpl{db: db}

	// Routes
	e.GET("/", root)
	e.POST("/items", serverImpl.addItem())
	e.GET("/items", serverImpl.getAllItems())
	e.GET("/search", serverImpl.getItemsByKeyword())
	e.GET("/items/:id", serverImpl.getItemById())
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
