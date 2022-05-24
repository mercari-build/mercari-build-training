package handler

import (
	"fmt"
	"os"
	"io"
	"bytes"
	"path"
	"net/http"
	"database/sql"
	"crypto/sha256"
	"encoding/hex"

	"github.com/labstack/echo/v4"
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"mercari-build-training-2022/app/models/customErrors/itemsError"
)

// Consts
const (
	ImgDir = "../image"
)

// Types
type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
	Image string `json:"image"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Response struct {
	Message string `json:"message"`
}

type Handler struct {
	DB *sql.DB
}

// Funcs
func getSHA256Binary(bytes[]byte) []byte {
	r := sha256.Sum256(bytes)
	return r[:]
}

// Validatorの定義
type CustomValidator struct{}

func (cv *CustomValidator) Validate(i interface{}) error {
	if c, ok := i.(validation.Validatable); ok {
		return c.Validate()
	}
	return nil
}

func (item Item) Validate() error {
	return validation.ValidateStruct(&item,
		validation.Field(
			&item.Name,
			validation.Required.Error("名前は必須入力です"),
			validation.RuneLength(1, 20).Error("名前は 1～20 文字です"),
		),
		validation.Field(
			&item.Category,
			validation.Required.Error("カテゴリーは必須入力です"),
			validation.RuneLength(1, 40).Error("カテゴリーは 1～20 文字です"),
		),
	)
}

// getItems is getting items list.
// @Summary get items
// @Description get all items
// @Produce  json
// @Success 200 {array} main.Items
// @Failure 500 {object} any
// @Router /items [get]
func (h Handler)GetItems(c echo.Context) error {
	var items Items

	// Exec Query
	rows, err := h.DB.Query(`SELECT name, category, image FROM items`)
	if err != nil {
		return itemsError.ErrGetItems.Wrap(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var category string
		var image sql.NullString //NULLを許容する

		// カーソルから値を取得
		if err := rows.Scan(&name, &category, &image); err != nil {
			return itemsError.ErrGetItems.Wrap(err)
		}

		items.Items = append(items.Items, Item{Name: name, Category: category, Image: image.String}) // image -> {"hoge", true}
	}

	return c.JSON(http.StatusOK, items)
}

// findItem is finding a  item by id.
// @Summary find an item
// @Description find an item by id
// @Produce json
// @Param id path int true "Item's id"
// @Success 200 {obejct} main.Item
// @Failure 500 {object} any
// @Router /items/:id [get]
func (h Handler)FindItem(c echo.Context) error {
	var item Item
	var name string
	var category string
	var image string

	// Exec Query
	itemId := c.Param("id")
	c.Logger().Infof("SELECT name, category, image FROM items WHERE id = %s", itemId)
	err := h.DB.QueryRow("SELECT name, category, image FROM items WHERE id = $1", itemId).Scan(&name, &category, &image)
	if err != nil {
		return itemsError.ErrFindItem.Wrap(err)
	}
	item = Item{Name: name, Category: category, Image: image}

	return c.JSON(http.StatusOK, item)
}

// searchItems is searching Items by name
// @Summary search Items by name
// @Description search Items by name
// @Produce json
// @Param keyword query string true "Keyword to match Item's name"
// @Success 200 {array} main.Items
// @Failure 500 {object} any
// @Router /items/search [get]
func (h Handler)SearchItems(c echo.Context) error {
	var items Items

	keyWord := c.QueryParam("keyword")

	// Exec Query
	rows, err := h.DB.Query(`SELECT name, category FROM items WHERE name LIKE ?`, keyWord + "%")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var category string
		var image string

		// カーソルから値を取得
		if err := rows.Scan(&name, &category, &image); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		items.Items = append(items.Items, Item{Name: name, Category: category, Image: image})
	}

	return c.JSON(http.StatusOK, items)
}

// addItem is adding an item.
// @Summary post an item
// @Description post an item to its table.
// @Produce  json
// @Param name body string true "Item's name"
// @Param category body string true "Item's category"
// @Param image formData any false "Item's image"
// @Success 200 {object} main.Response
// @Failure 500 {object} any
// @Router /items [post]
func (h Handler)AddItem(c echo.Context) error {
	// Inintialize Item
	var item Item
	// Get form data
	item.Name = c.FormValue("name")
	item.Category = c.FormValue("category")
	file, err := c.FormFile("image")
	if err != nil {
		return itemsError.ErrPostItem.Wrap(err)
	}

	// Validate item fields
	if err := c.Validate(item); err != nil {
		errs := err.(validation.Errors)
		for k, err := range errs {
			c.Logger().Error(k + ": " + err.Error())
		}
		return itemsError.ErrPostItem.Wrap(err)
	}

	// Open Image File
	imageFile, err := file.Open()
	if err != nil {
		return itemsError.ErrPostItem.Wrap(err)
	}
	defer imageFile.Close()

	// Read Image Bytes
	imageBytes, err := io.ReadAll(imageFile)
	if err != nil {
		return itemsError.ErrPostItem.Wrap(err)
	}

	// Encode Image
	sha := sha256.New()
	sha.Write([]byte(imageBytes))
	item.Image = hex.EncodeToString(getSHA256Binary(imageBytes)) + ".jpg"

	c.Logger().Infof("Receive item: %s which belongs to the category %s. image name is %s", item.Name, item.Category, item.Image)

	message := fmt.Sprintf("item received: %s which belongs to the category %s. image name is %s", item.Name, item.Category, item.Image)

	// Save Image to ./image
	imgFile, err := os.Create(path.Join(ImgDir, item.Image))
	if err != nil {
		return itemsError.ErrPostItem.Wrap(err)
	}
	_, err = io.Copy(imgFile, bytes.NewReader(imageBytes))
	if err != nil {
		return itemsError.ErrPostItem.Wrap(err)
	}

	// Exec Query
	_, err = h.DB.Exec(`INSERT INTO items (name, category, image) VALUES (?, ?, ?)`, item.Name, item.Category, item.Image)
	if err != nil {
		return itemsError.ErrPostItem.Wrap(err)
	}
	
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}