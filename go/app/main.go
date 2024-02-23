package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	// "encoding/json"
	"strconv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"crypto/sha256"
	"io"
	"path/filepath"
	"encoding/hex"
	"mime/multipart"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)
//フォルダアクセス用宣言
const (
	ImgDir = "images"
	JsonFile = "items.json"
)

//商品情報用構造体
type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
	Images string `json:"images"`
	Id int `json:"id"`
}
//商品情報用構造体↑を一覧する用構造体
type Items struct {
	Items []Item `json:"items"`
}
//response用構造体
type Response struct {
	Message string `json:"message"`
}

type Itemdb struct {
	ID int `db:"id"`
	Name string `db:"name"`
	Category string `db:"category"`
	ImageName string `db:"image_name"`
}
type Itemsdb struct{
	Itemsdb []Itemdb `db:"itemsdb"`
}
type Categories struct{
	ID int `db:"id"`
	Name string `db:"name"`
}
type ItemWithCategory struct{
	ID int `db:"id"`
	Name string `db:"name"`
	CategoryID int `db:"category_id"`
	ImageName string `db:"image_name"`
	CategoryName string `db:"category_name"`
	// CategoriesID int `db:"categories_id`
}
type ItemsWithCategory struct{
	ItemsWithCategory []ItemWithCategory `db:"itemswithcategory"`
}

func getItemFromRequest(c echo.Context) (string,string,*multipart.FileHeader, error){
	//Get date from request
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil{
		c.Logger().Errorf("FormFile error : %v\n", err)
		return "", "", nil, err
	}
	return name, category, image, nil
}

func getNewItemsForClm(c echo.Context) (string,*multipart.FileHeader, error) {
	//Get date from request
	name := c.FormValue("name")
	image, err := c.FormFile("image")
	if err != nil{
		c.Logger().Errorf("FormFile error : %v\n", err)
		//*multipart.FileHeaderは ""でなくてnil
		return "", nil, err
	}
	return name, image, nil
}

func getCategoryID(db *sql.DB, categoryName string, c echo.Context) (int, error) {
	//prepare variable for id
	var categoryId int
	//Query.Row id
	cmd := "SELECT id FROM categories where name = ?"
	//run sql-stmt and assigning the retreived data(because got data by sql-stmt already) to variable
	//scanは一度sqlで取得したデータをプログラムが使いやすいように変数に割り当てるメソッド
	row := db.QueryRow(cmd, categoryName)
	// fmt.Printf("QueryRow !!!err: %v\n",categoryName)
	//rowにはidが入っている
	err := row.Scan(&categoryId)
	// fmt.Printf("QueryRow err!!: %v\n",categoryName)
	// if err != nil {
	// 	fmt.Printf("we haven't had its ID, so let's insert it : %v",categoryId)
	// }
	if err != nil {
		//no wanted rows error
		if err == sql.ErrNoRows {
			err := insertNewclm(db, categoryName, c)
			if err != nil {
				// fmt.Printf("cate!!!: %v", categoryName)
				return 0, fmt.Errorf("insertNewclm error : %v", err)
			}
		}
		//ex.sql-stmt error & connection error
		addItem(db)
	}
	return categoryId, nil
}

func insertNewclm(db *sql.DB, categoryName string, c echo.Context) error {
		// fmt.Printf("category!?: %v", categoryName)
		// fmt.Printf("category!!!: %v\n", categoryName)
		name, image, err := getNewItemsForClm(c)
		if err != nil {
			c.Logger().Errorf("getItemFromRequest error: %v", err)
		}
		image_name,err := ImgSave(image)
		if err != nil {
			c.Logger().Errorf("ImgSave error: %v",err)
		}
		// c.Logger().Errorf("name!?: %v, %v",image_name,name)
		cmd := "INSERT INTO categories (name) VALUES(?)"
		stmt, err := db.Prepare(cmd)
		if err != nil {
			c.Logger().Errorf("db.Prepare error : %v\n",err)
		}
		addQuery, err := stmt.Exec(categoryName)
		if err != nil {
			c.Logger().Errorf("addQuery error : %v", err)
		}
		category_id, err := addQuery.LastInsertId()
		if err != nil {
			c.Logger().Errorf("getting last insert ID error : %v",err)
		}
		// c.Logger().Errorf("category_id!?: %v",category_id)

		ItemQuery := "INSERT INTO items (name,category_id,image_name) VALUES (?,?,?)"
		insertQuery, err := db.Prepare(ItemQuery)
		if err != nil {
			c.Logger().Errorf("db.Prepare error: %v", err)
			return err
		}
		defer insertQuery.Close()
		//Exec(sql-statement is executed)(get Exec's arguments into ↑?,?,?)
		//you can change param easily nothing to change sql-statement
		if _,err = insertQuery.Exec(name,category_id,image_name); err != nil{
			c.Logger().Errorf("insertQuery.Exec error", err)
			return err
		}

		return nil
}

func ImgSave(image *multipart.FileHeader)(string, error) {
		// image, err := c.FormFile("image")
		// if err != nil{
		// 	c.Logger().Errorf("FormFile error : %v\n", err)
		// 	return err
		// }
		src, err := image.Open()
		if err != nil {
			return "", fmt.Errorf("image.Open error: %v", err)
		}
		defer src.Close()
		//receive hashedImage
		hashedString, err := getHash(src)
		if err != nil{
			return "", fmt.Errorf("getHash error: %v",err)
		}
		
		image_name := hashedString + ".jpg"
		if _, err := src.Seek(0,0); err != nil {
			return "", fmt.Errorf("failed to seek file: %v", err)
		}
		//create file path
		dst, err := os.Create(filepath.Join(ImgDir, image_name))
		if err != nil{
			return "", fmt.Errorf("os.Create filepath.Join error: %v",err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst,src); err != nil {
			return "", fmt.Errorf("io.Copy error: %v",err)
		}

		return image_name, nil
} 

func addItem(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		name, category, image, err := getItemFromRequest(c)
		// fmt.Printf("getItemFrom!!! %v\n", category)
		if err != nil {
			c.Logger().Errorf("getItemFromRequest error")
			return err
		}
		image_name,err := ImgSave(image)
		if err != nil {
			c.Logger().Errorf("ImgSave error: %v\n", err)
			return err
		}
		// fmt.Printf("getItemFrom??? %v\n", category)
		category_id, err := getCategoryID(db, category,c)
		
		if err != nil {
			c.Logger().Errorf("getCategoryID error %v:",err)
			return err
		}
		//insert date into a table
		//name(categories)->id(categories)->category_id(items)
		cmd := "INSERT INTO items (name, category_id, image_name) VALUES(?, ?, ?)"
		
		//prepare sql-statement
		stmt, err := db.Prepare(cmd)
		if err != nil {
			c.Logger().Errorf("db.Prepare error: %v", err)
			return err
		}
		defer stmt.Close()
		//Exec(sql-statement is executed)(get Exec's arguments into ↑?,?,?)
		//you can change param easily nothing to change sql-statement
		if _,err = stmt.Exec(name,category_id,image_name); err != nil{
			c.Logger().Errorf("stmt.Exec error", err)
			return err
		}
		// defer db.Close()
		message := "You've successfully added items"
		res := Response{Message: message}
		return c.JSON(http.StatusOK, res)
	}
}

func sqlOpen()(*sql.DB, error){
	db, err := sql.Open("sqlite3","../db/mercari.sqlite3")
	if err != nil {
		return nil, fmt.Errorf("sql.Open error: %v",err)
	}
	return db, nil
}

func getHash(src io.Reader) (string, error) {
	hash := sha256.New()
	if _,err := io.Copy(hash,src); err != nil{
		return "", fmt.Errorf("io.Copy error : %v",err)
	}
	HashedString := hex.EncodeToString(hash.Sum(nil))
	return HashedString, nil
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

func getItemByItemId(db *sql.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
        itemId, err := strconv.Atoi(c.Param("item_id"))
        if err != nil {
            // Use echo's HTTP error to return the error to the client properly.
            return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid item ID: %v", err))
        }

        var items ItemWithCategory
        cmd := `SELECT items.id, items.name, categories.name, items.category_id ,items.image_name FROM items LEFT OUTER JOIN categories ON items.category_id = categories.id WHERE items.id = ?`

        err = db.QueryRow(cmd, itemId).Scan(&items.ID, &items.Name, &items.CategoryName,&items.CategoryID,&items.ImageName)
		// fmt.Printf("item.ID %v",items.ID)
        if err != nil {
            if err == sql.ErrNoRows {
                // Item not found
                return echo.NewHTTPError(http.StatusNotFound, "Item not found")
            }
            // Internal server error for other types of errors
            return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
        }

        return c.JSON(http.StatusOK, items)
    }
}

func getItems (c echo.Context) error {
		//connect to db
		db, err := sql.Open("sqlite3","../db/mercari.sqlite3")
		if err != nil {
			c.Logger().Errorf("sql.Open error")
			return err
		}
		defer db.Close()
		//get db-date(return sql.Rows object)
		cmd := "SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id"
		// cmd := "SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id"
		rows, err := db.Query(cmd)
		if err != nil {
			c.Logger().Errorf("db.Query error")
			return err
		}
		defer rows.Close()
		//get db-date into item
		var itemsdb []Itemdb
		for rows.Next() {
			//declare item instance for storage
			var itemdb Itemdb
			//store rows-date to item-field
			if err := rows.Scan(&itemdb.Name,&itemdb.Category,&itemdb.ImageName); err != nil{
				c.Logger().Errorf("rows.Scan: rows-date storing error")
				return err
			}
			itemsdb = append(itemsdb,itemdb)
		}
		//check roup error
		if err = rows.Err(); err != nil {
			c.Logger().Errorf("for roup error(rows.Next())")
			return err
		}
		
		// return rows-date
		return c.JSON(http.StatusOK, itemsdb)
}

func searchItems(c echo.Context) error {
	//receive keywords from request
	keyword := c.QueryParam("keyword")
	//connect to db
	db, err := sqlOpen()
	if err != nil {
		c.Logger().Errorf("sql.Open error")
		return err
	}
	defer db.Close()
	//search keywords to items in db
	cmd := "SELECT items.id ,items.name , items.category_id ,categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?"
	// cmd := "SELECT categories.name FROM categories WHERE name LIKE ?"
	rows, err := db.Query(cmd,"%"+keyword+"%")
	// fmt.Printf("err! %v",rows)
	// fmt.Printf("err? %v", err)
	// fmt.Printf("err# %v", keyword)

	if err != nil {
		c.Logger().Errorf("db.Query error")
	}
	defer rows.Close()
	//get items matched it
	var itemswithcategory []ItemWithCategory
	for rows.Next() {
		var itemWithCategory ItemWithCategory
		if err := rows.Scan(&itemWithCategory.ID,&itemWithCategory.Name, &itemWithCategory.CategoryID, &itemWithCategory.CategoryName, &itemWithCategory.ImageName); err != nil {
			c.Logger().Errorf("rows.Scan error")
			return err
		}
		itemswithcategory = append(itemswithcategory, itemWithCategory)
	}
	//return its items
	return c.JSON(http.StatusOK, itemswithcategory)
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
	//db := InitDB()だとDB初期化(複雑な要件がある場合に統括的に設定)
	//sqlOpenは複雑ではないときに、エラーハンドリングができる
	db, err := sqlOpen()
	if err != nil {
		//%vはどんな型でもOK,%sはstring型
		e.Logger.Infof("Failed to open the database: %v",err)
	}
	defer db.Close()
	// e.GET("/", root)
	e.POST("/items", addItem(db))
	e.GET("/items",getItems)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:item_id", getItemByItemId(db))
	// e.GET("/items/:id", func(c echo.Context) error {
	// 	return getItemById(c, db)
	// })
	e.GET("/search",searchItems)
	// Start server
	e.Logger.Fatal(e.Start(":9000")) 
}

