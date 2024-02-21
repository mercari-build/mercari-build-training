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
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

type Item struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Conn struct {
	db *sql.DB
}

func (conn Conn) addItem() echo.HandlerFunc {
	return func(c echo.Context) error {
		// リクエストボディからデータを取得
		name := c.FormValue("name")
		category := c.FormValue("category")

		// 画像ファイルを取得
		imageFile, err := c.FormFile("image")
		if err != nil {
			res := Response{Message: "failed to get image file in addItem"}
			return c.JSON(http.StatusBadRequest, res)
		}
		src, err := imageFile.Open()
		if err != nil {
			res := Response{Message: "failed to open image file in addItem"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		defer src.Close()

		// 画像ファイルをハッシュ化
		hash := sha256.New()
		if _, err := io.Copy(hash, src); err != nil {
			res := Response{Message: "failed to hash image file in addItem"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		hashedImageName := fmt.Sprintf("%x.jpeg", hash.Sum(nil))

		// 画像ファイルを保存
		dst, err := os.Create(fmt.Sprintf("images/%s", hashedImageName))
		if err != nil {
			res := Response{Message: fmt.Sprintf("failed to create image file in addItem: image=%s", hashedImageName)}
			return c.JSON(http.StatusInternalServerError, res)
		}
		defer dst.Close()
		src.Seek(0, 0) // ファイルポインタを先頭に戻す
		//srcからdstへ内容をコピー
		if _, err := io.Copy(dst, src); err != nil {
			res := Response{Message: "failed to save image file in addItem"}
			return c.JSON(http.StatusInternalServerError, res)
		}

		// DBへの保存
		if err := conn.addItemToDB(name, category, hashedImageName); err != nil {
			res := Response{Message: fmt.Sprintf("failed to add item to DB in addItem: %s", err)}
			return c.JSON(http.StatusInternalServerError, res)
		}

		c.Logger().Infof("Receive item: %s", name)

		message := fmt.Sprintf("item received: name=%s,category=%s,images=%s", name, category, hashedImageName)
		res := Response{Message: message}

		return c.JSON(http.StatusOK, res)
	}
}

func (conn Conn) addItemToDB(name, category, imageName string) error {
	// トランザクションを開始
	tx, err := conn.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction in addItemToDB: %w", err)
	}
	// カテゴリをcategoriesテーブルに追加
	stmt1, err := tx.Prepare("INSERT INTO categories (name) VALUES (?)")
	if err != nil {
		// ロールバック
		errRollback := tx.Rollback()
		if errRollback != nil {
			return fmt.Errorf("failed to rollback :stmt1 prepare: %w", errRollback)
		}
		return fmt.Errorf("failed to prepare SQL statement1 in addItemToDB: %w", err)
	}
	defer stmt1.Close()

	result, err := stmt1.Exec(category)
	if err != nil {
		// ロールバック
		errRollback := tx.Rollback()
		if errRollback != nil {
			return fmt.Errorf("failed to rollback :stmt1 exec: %w", errRollback)
		}
		return fmt.Errorf("failed to execute SQL statement1 in addItemToDB: %w", err)
	}
	// 新しく挿入された行のIDを取得
	id, err := result.LastInsertId()
	if err != nil {
		// ロールバック
		errRollback := tx.Rollback()
		if errRollback != nil {
			return fmt.Errorf("failed to rollback :get last insert ID: %w", errRollback)
		}
		return fmt.Errorf("failed to get last insert ID in addItemToDB: %w", err)
	}

	// itemsテーブルに商品を追加
	stmt2, err := tx.Prepare("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)")
	if err != nil {
		// ロールバック
		errRollback := tx.Rollback()
		if errRollback != nil {
			return fmt.Errorf("failed to rollback :stmt2 prepare: %w", errRollback)
		}
		return fmt.Errorf("failed to prepare SQL statement2 in addItemToDB: %w", err)
	}
	defer stmt2.Close()
	if _, err := stmt2.Exec(name, id, imageName); err != nil {
		// ロールバック
		errRollback := tx.Rollback()
		if errRollback != nil {
			return fmt.Errorf("failed to rollback :stmt2 exec: %w", errRollback)
		}
		return fmt.Errorf("failed to execute SQL statement2 in addItemToDB: %w", err)
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction in addItemToDB: %w", err)
	}

	return nil
}

func (conn Conn) getAllItems() echo.HandlerFunc {
	return func(c echo.Context) error {
		// itemsテーブルとcategoriesテーブルをJOINして全てのアイテムを取得
		rows, err := conn.db.Query("SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id")
		if err != nil {
			res := Response{Message: "failed to get items from DB in getAllItems"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		defer rows.Close()

		var allItems Items
		for rows.Next() {
			var item Item
			if err := rows.Scan(&item.Name, &item.Category, &item.ImageName); err != nil {
				res := Response{Message: "failed to scan items from DB in getAllItems"}
				return c.JSON(http.StatusInternalServerError, res)
			}
			allItems.Items = append(allItems.Items, item)
		}
		return c.JSON(http.StatusOK, allItems)
	}
}

func (conn Conn) getItemsByKeyword() echo.HandlerFunc {
	return func(c echo.Context) error {
		// クエリパラメータからキーワードを取得
		keyword := c.QueryParam("keyword")

		// DBから名前にキーワードを含む商品一覧を返す
		searchKeyword := "%" + keyword + "%" // 部分一致検索
		rows, err := conn.db.Query(`
			SELECT items.name, categories.name, items.image_name 
			FROM items JOIN categories ON items.category_id = categories.id 
			WHERE items.name LIKE ?`, searchKeyword)
		if err != nil {
			res := Response{Message: fmt.Sprintf("failed to search items from DB in getItemsByKeyword: keyword=%s", keyword)}
			return c.JSON(http.StatusInternalServerError, res)
		}
		defer rows.Close()

		var keywordItems Items
		for rows.Next() {
			var item Item
			if err := rows.Scan(&item.Name, &item.Category, &item.ImageName); err != nil {
				res := Response{Message: "failed to scan items from DB in getItemsByKeyword"}
				return c.JSON(http.StatusInternalServerError, res)
			}
			keywordItems.Items = append(keywordItems.Items, item)
		}
		return c.JSON(http.StatusOK, keywordItems)
	}
}

func (conn Conn) getItemById() echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			res := Response{Message: "failed to get id in getItemById"}
			return c.JSON(http.StatusBadRequest, res)
		}

		// DBからIDに対応する商品を取得
		row := conn.db.QueryRow(`
			SELECT items.name, categories.name, items.image_name 
			FROM items JOIN categories ON items.category_id = categories.id 
			WHERE items.id = ?`, id)

		var item Item
		if err := row.Scan(&item.Name, &item.Category, &item.ImageName); err != nil {
			if err == sql.ErrNoRows { // IDに対応する商品がない場合
				res := Response{Message: fmt.Sprintf("Item not found: id=%d", id)}
				return c.JSON(http.StatusNotFound, res)
			} else {
				res := Response{Message: fmt.Sprintf("failed to scan item from DB in getItemById: id=%d", id)}
				return c.JSON(http.StatusInternalServerError, res)
			}
		}
		return c.JSON(http.StatusOK, item)
	}
}

func LoadItemsFromJSON() (*Items, error) {
	jsonFile, err := os.Open("items.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var allItems Items
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&allItems); err != nil {
		return nil, err
	}
	return &allItems, nil
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
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

	conn := Conn{db: db}

	// Routes
	e.GET("/", root)
	e.POST("/items", conn.addItem())
	e.GET("/items", conn.getAllItems())
	e.GET("/search", conn.getItemsByKeyword())
	e.GET("/items/:id", conn.getItemById())
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
