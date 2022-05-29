package handler

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"mercari-build-training-2022/app/models/customErrors/itemsError"
	"mercari-build-training-2022/app/models/customErrors/usersError"
)

// Consts
const (
	ImgDir = "../image"
)

// Types

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type UserResponse struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Item struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Category        string `json:"category"`
	Image           string `json:"image"`
	Price           int    `json:"price"`
	PriceLowerLimit int    `json:"price_lowe_limit"`
	UserId          int    `json:"user_id"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Transaction struct {
	Id                  int `json:"id`
	DeterminedPrice     int `json:"determined_price"`
	ItemId              int `json:"item_id"`
	BuyerId             int `json:"buyer_id"`
	TransactionStatusId int `json:"transaction_status_id"`
}

type Qa struct {
	Id       int    `json:"id`
	ItemId   int    `json:"item_id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	QaTypeId int    `json:"qa_type_id"`
}

type Qas struct {
	Qas []Qa `json:"qas"`
}

type Response struct {
	Message string `json:"message"`
}

type Handler struct {
	DB *sql.DB
}

// Funcs
func getSHA256Binary(bytes []byte) []byte {
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

func (user User) Validate() error {
	return validation.ValidateStruct(&user,
		validation.Field(
			&user.Name,
			validation.Required.Error("名前は必須入力です(Name is required)"),
			validation.RuneLength(1, 20).Error("名前は 1～20 文字です"),
		),
		validation.Field(
			&user.Password,
			validation.Required.Error("パスワードは必須入力です(Email is required)"),
			validation.RuneLength(4, 20).Error("パスワードは4～20 文字です"),
			is.Alphanumeric,
		),
	)
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

func (transaction Transaction) Validate() error {
	return validation.ValidateStruct(&transaction,
		validation.Field(
			&transaction.ItemId,
			validation.Required.Error("決定した値段は必須です"),
		),
		validation.Field(
			&transaction.ItemId,
			validation.Required.Error("商品IDは必須です"),
		),
		validation.Field(
			&transaction.BuyerId,
			validation.Required.Error("買い手IDは必須です"),
		),
		validation.Field(
			&transaction.TransactionStatusId,
			validation.Required.Error("取引状態IDは必須です"),
		),
	)
}

func (qa Qa) Validate() error {
	return validation.ValidateStruct(&qa,
		validation.Field(
			&qa.ItemId,
			validation.Required.Error("商品IDは必須です"),
		),
		validation.Field(
			&qa.Question,
			validation.Required.Error("質問内容は必須です"),
		),
		validation.Field(
			&qa.Answer,
			validation.Required.Error("解答内容は必須です"),
		),
		validation.Field(
			&qa.QaTypeId,
			validation.Required.Error("質問種別IDは必須です"),
		),
	)
}

// AddUser is adding a user by BasicAuth.
// @Summary add a user
// @Description adding a user by BasicAuth.
// @Produce json
// @Success 200 {objext} any
// @Failure 500 {object} any
// @Router /users [post]
func (h Handler) AddUser(c echo.Context) error {
	// Inintialize Item
	var user User
	// Get form data
	user.Name = c.FormValue("name")
	user.Password = c.FormValue("password")

	// Validate item fields
	if err := c.Validate(user); err != nil {
		errs := err.(validation.Errors)
		for k, err := range errs {
			c.Logger().Error(k + ": " + err.Error())
		}
		return usersError.ErrPostUser.Wrap(err)
	}

	// Exec Query
	_, err := h.DB.Exec(`INSERT INTO users (name, password) VALUES (?, ?)`, user.Name, user.Password)
	if err != nil {
		c.Logger().Error(err.Error())
		return usersError.ErrPostUser.Wrap(err)
	}

	message := fmt.Sprintf("Hello, %s !!", user.Name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

// findUser is finding a user by id.
// @Summary find a user
// @Description find a user by id
// @Produce json
// @Param id path int true "User's id"
// @Success 200 {obejct} main.UserResponse
// @Failure 500 {object} any
// @Router /items/:id [get]
func (h Handler) FindUser(c echo.Context) error {
	var name string
	var id int

	// Exec Query
	userId := c.Param("id")
	err := h.DB.QueryRow("SELECT id, name FROM users WHERE id = $1", userId).Scan(&id, &name)
	if err != nil {
		c.Logger().Error(err.Error())
		return usersError.ErrFindUser.Wrap(err)
	}
	response := UserResponse{Id: id, Name: name}

	return c.JSON(http.StatusOK, response)
}

// getItems is getting items list.
// @Summary get items
// @Description get all items
// @Produce  json
// @Success 200 {array} main.Items
// @Failure 500 {object} any
// @Router /items [get]
func (h Handler) GetItems(c echo.Context) error {
	var items Items

	// Exec Query
	rows, err := h.DB.Query(`SELECT id, name, category, image, price, price_lower_limit, user_id FROM items`)
	if err != nil {
		return itemsError.ErrGetItems.Wrap(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var category string
		var price int
		var priceLowerLimit int
		var userId int
		var image sql.NullString //NULLを許容する

		// カーソルから値を取得
		if err := rows.Scan(&id, &name, &category, &image, &price, &priceLowerLimit, &userId); err != nil {
			return itemsError.ErrGetItems.Wrap(err)
		}

		items.Items = append(items.Items, Item{Name: name, Category: category, Image: image.String, Price: price, PriceLowerLimit: priceLowerLimit, UserId: userId}) // image -> {"hoge", true}
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
func (h Handler) FindItem(c echo.Context) error {
	var item Item
	var id int
	var name string
	var category string
	var image string
	var price int
	var priceLowerLimit int
	var userId int

	// Exec Query
	itemId := c.Param("id")
	c.Logger().Infof("SELECT id, name, category, image, price, price_lower_limit, user_id FROM items WHERE id = %s", itemId)
	err := h.DB.QueryRow("SELECT id, name, category, image, price, price_lower_limit, user_id FROM items WHERE id = $1", itemId).Scan(&id, &name, &category, &image, &price, &priceLowerLimit, &userId)
	if err != nil {
		return itemsError.ErrFindItem.Wrap(err)
	}
	item = Item{Name: name, Category: category, Image: image, Price: price, PriceLowerLimit: priceLowerLimit, UserId: userId}

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
func (h Handler) SearchItems(c echo.Context) error {
	var items Items

	keyWord := c.QueryParam("keyword")

	// Exec Query
	rows, err := h.DB.Query(`SELECT id, name, category, image, price, price_lower_limit, user_id FROM items WHERE name LIKE ?`, keyWord+"%")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var category string
		var image string
		var price int
		var priceLowerLimit int
		var userId int

		// カーソルから値を取得
		if err := rows.Scan(&id, &name, &category, &image, &price, &priceLowerLimit, &userId); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		items.Items = append(items.Items, Item{Name: name, Category: category, Image: image, Price: price, PriceLowerLimit: priceLowerLimit, UserId: userId})
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
func (h Handler) AddItem(c echo.Context) error {
	// Inintialize Item
	var item Item
	// Get form data
	item.Name = c.FormValue("name")
	item.Category = c.FormValue("category")
	item.Price, _ = strconv.Atoi(c.FormValue("price"))
	item.PriceLowerLimit, _ = strconv.Atoi(c.FormValue("price_lower_limit"))
	item.UserId, _ = strconv.Atoi(c.FormValue("user_id"))
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
	_, err = h.DB.Exec(
		`INSERT INTO items (name, category, image, price, price_lower_limit, user_id) VALUES (?, ?, ?, ?, ?, ?)`,
		item.Name, item.Category, item.Image, item.Price, item.PriceLowerLimit, item.UserId)
	if err != nil {
		return itemsError.ErrPostItem.Wrap(err)
	}

	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

// addItem is adding a transaction.
// @Summary post a Transaction
// @Description post a transaction to its table.
// @Produce  json
// @Param determined_price body int true "Transaction's determined_price"
// @Param item_id body int true "Item's id"
// @Param buyer_id body int false "Buyer's id"
// @Param transaction_status_id body int "Transation's status id"
// @Success 200 {object} main.Response
// @Failure 500 {object} any
// @Router /items [post]
func (h Handler) AddTransaction(c echo.Context) error {
	// Inintialize Transaction
	var transaction Transaction

	// Get form data
	transaction.DeterminedPrice, _ = strconv.Atoi(c.FormValue("determined_price"))
	transaction.ItemId, _ = strconv.Atoi(c.FormValue("item_id"))
	transaction.BuyerId, _ = strconv.Atoi(c.FormValue("buyer_id"))
	transaction.TransactionStatusId, _ = strconv.Atoi(c.FormValue("transaction_status_id"))

	// Validate item fields
	if err := c.Validate(transaction); err != nil {
		errs := err.(validation.Errors)
		for k, err := range errs {
			c.Logger().Error(k + ": " + err.Error())
		}
		return usersError.ErrPostUser.Wrap(err)
	}

	// Exec Query
	_, err := h.DB.Exec(`INSERT INTO transactions (determined_price, item_id, buyer_id, transaction_status_id) VALUES (?, ?, ?, ?)`, transaction.DeterminedPrice, transaction.ItemId, transaction.BuyerId, transaction.TransactionStatusId)
	if err != nil {
		c.Logger().Error(err.Error())
		return usersError.ErrPostUser.Wrap(err)
	}

	message := fmt.Sprintf("Transaction created: %v", transaction)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

// findTransaction is finding a transaction by item_id and buyer_id.
// @Summary find an transaction
// @Description find a transaction by item_id and buyer_id
// @Produce json
// @Param item_id path int true "Item's id"
// @Param buyer_id path int true "Buyer's id"
// @Success 200 {obejct} main.Transaction
// @Failure 500 {object} any
// @Router /transactions/:item_id/:buyer_id [get]
func (h Handler) FindTransaction(c echo.Context) error {
	var id int
	var determinedPrice int
	var itemId int
	var buyerId int
	var transactionStatusId int

	// Exec Query
	itemId, _ = strconv.Atoi(c.Param("item_id"))
	buyerId, _ = strconv.Atoi(c.Param("buyer_id"))
	c.Logger().Infof("SELECT id, name, category, image, price, price_lower_limit, user_id FROM items WHERE id = %s", itemId)
	err := h.DB.QueryRow("SELECT id, determined_price, item_id, buyer_id, transaction_status_id FROM transactions WHERE item_id = $1 AND buyer_id = $2", itemId, buyerId).Scan(&id, &determinedPrice, &itemId, &buyerId, &transactionStatusId)
	if err != nil {
		return itemsError.ErrFindItem.Wrap(err)
	}
	transaction := Transaction{Id: id, DeterminedPrice: determinedPrice, ItemId: itemId, BuyerId: buyerId, TransactionStatusId: transactionStatusId}

	return c.JSON(http.StatusOK, transaction)
}

// AddQa is adding a qa.
// @Summary add a qa
// @Description add a qa data
// @Produce json
// @Param item_id body int true "Item's id"
// @Param question body string true "Question text"
// @Param answer body string true "Answer text"
// @Param qa_type_id body int true "Item's id"
// @Success 200 {obejct} main.Response
// @Failure 500 {object} any
// @Router /qas [post]
func (h Handler) AddQa(c echo.Context) error {
	// Inintialize Qa
	var qa Qa
	// Get form data
	qa.ItemId, _ = strconv.Atoi(c.FormValue("item_id"))
	qa.Question = c.FormValue("question")
	qa.Answer = c.FormValue("answer")
	qa.QaTypeId, _ = strconv.Atoi(c.FormValue("qa_type_id"))

	// Validate qa fields
	if err := c.Validate(qa); err != nil {
		errs := err.(validation.Errors)
		for k, err := range errs {
			c.Logger().Error(k + ": " + err.Error())
		}
		return usersError.ErrPostUser.Wrap(err)
	}

	// Exec Query
	_, err := h.DB.Exec(`INSERT INTO qas (item_id, question, answer, qa_type_id) VALUES (?, ?, ?, ?)`, qa.ItemId, qa.Question, qa.Answer, qa.QaTypeId)
	if err != nil {
		c.Logger().Error(err.Error())
		return usersError.ErrPostUser.Wrap(err)
	}

	c.Logger().Infof("Receive qas")

	message := fmt.Sprintf("qas received")

	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

// GetQas is getting qas by item_id.
// @Summary getting qas by item_id
// @Description GetQas is getting qas by item_id.
// @Produce json
// @Param item_id path int true "Item's id"
// @Success 200 {obejct} main.Qas
// @Failure 500 {object} any
// @Router /qas/:item_id [get]
func (h Handler) GetQas(c echo.Context) error {

	var qas Qas

	// Exec Query
	itemId, _ := strconv.Atoi(c.Param("item_id"))
	c.Logger().Infof("SELECT id, item_id, question, answer, qa_type_id FROM items WHERE id = %s", itemId)
	rows, err := h.DB.Query("SELECT id, item_id, question, answer, qa_type_id FROM qas WHERE item_id = $1", itemId)
	if err != nil {
		return itemsError.ErrFindItem.Wrap(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var itemId int
		var question string
		var answer string
		var qaTypeId int

		// カーソルから値を取得
		if err := rows.Scan(&id, &itemId, &question, &answer, &qaTypeId); err != nil {
			return itemsError.ErrGetItems.Wrap(err)
		}

		qas.Qas = append(qas.Qas, Qa{Id: id, ItemId: itemId, Question: question, Answer: answer, QaTypeId: qaTypeId})
	}

	return c.JSON(http.StatusOK, qas)
}
