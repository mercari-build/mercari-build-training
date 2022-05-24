package main

import (
	"os"
	"io/ioutil"
	"bytes"
	"net/http"
	"mime/multipart"
	"testing"
	"net/http/httptest"
	"crypto/sha256"
	"encoding/hex"

	"github.com/labstack/echo/v4"

	"mercari-build-training-2022/app/handler"
	"mercari-build-training-2022/app/models/db"
)

// Types
type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
	Image string `json:"image"`
}

// Funcs
func getSHA256Binary(bytes[]byte) []byte {
	r := sha256.Sum256(bytes)
	return r[:]
}

// Test for GetItems
func TestGetItems(t *testing.T) {
	// 環境変数の設定
	os.Setenv("ENV", "test")

	// echoの初期化
	e := echo.New()

	// DBの初期化
	handler := handler.Handler{ DB: db.DbConnection }

	// First Request
    req := httptest.NewRequest(http.MethodGet, "/items", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.SetPath("/")
	
	if err := handler.GetItems(c); err != nil {
		t.Errorf("First: Couldn't get items")
	} 
	first_items := rec.Body.String()

	// Second Request
    req = httptest.NewRequest(http.MethodGet, "/items", nil)
    rec = httptest.NewRecorder()
    c = e.NewContext(req, rec)
    c.SetPath("/")
	
	if err := handler.GetItems(c); err != nil {
		t.Errorf("Second: Couldn't get items")
	} 
	second_items := rec.Body.String()
	if first_items != second_items {
        t.Errorf("expected response %s, got %s", first_items, second_items)
    }
}

// Test for AddItems
func TestAddItem(t *testing.T) {
	// Set env variables
	os.Setenv("ENV", "test")

	// Init echo
	e := echo.New()
	e.Validator = &handler.CustomValidator{}

	// Init DB
	handler := handler.Handler{ DB: db.DbConnection }

	// Item for Test
	item := Item{ Name: "テスト太郎", Category: "テストカテゴリ", Image: "../image/default.jpg"}

	file, err := os.Open(item.Image)
	if err != nil {
		t.Errorf("Couldn't find a File: " + err.Error())
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("Couldn't Read File: " + err.Error())
	}

	// Encode Image
	sha := sha256.New()
	sha.Write([]byte(fileContents))
	shaImageName := hex.EncodeToString(getSHA256Binary(fileContents)) + ".jpg"

	fi, err := file.Stat()
	if err != nil {
		t.Errorf("Couldn't Stat File: " + err.Error())
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", fi.Name())
	if err != nil {
		t.Errorf("Couldn't create FormFile: " + err.Error())
	}
	part.Write(fileContents)
	_ = writer.WriteField("name", item.Name)
	_ = writer.WriteField("category", item.Category)
	writer.Close()

    req := httptest.NewRequest(http.MethodPost, "/items", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.SetPath("/")

	if err = handler.AddItem(c); err != nil {
		t.Errorf("Couldn't post item: " + err.Error())
	}

	var resItem Item
	err = handler.DB.QueryRow("SELECT name, category, image FROM items DESC LIMIT 1").Scan(&resItem.Name, &resItem.Category, &resItem.Image)
	if err != nil {
		t.Errorf("Couldn't get an item from DB: " + err.Error())
	}
	if (resItem.Name != item.Name || resItem.Category != item.Category || resItem.Image != shaImageName) {
		t.Errorf("got %v", resItem)
	}
}