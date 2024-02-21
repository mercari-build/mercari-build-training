package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"strconv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"crypto/sha256"
	// "path/filepath"
	"io"
	"encoding/hex"
	"mime/multipart"
)
//フォルダアクセス用宣言
const (
	ImgDir = "images/"
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

//helloworld用関数
func root(c echo.Context) error {
	//構造体のfield名がjsonのキーに対応
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}
//個人的memo = サーバーはデコードされたやつがほしい。クライアントやDB/fileはエンコードされたやつがほしい
//items.jsonに追加する関数
func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	idStr := c.FormValue("id")
	// fmt.Printf("%v\n",idStr)
	c.Logger().Infof("Receive id : id=%s",idStr)
	id, err := strconv.Atoi(idStr)
	// c.Logger().Infof("Receive item: id=%s, name=%s, category=%s, idStr,name,category")
	//print/loggerでterminalみて
	if err != nil{
		fmt.Printf("strconv.Atoi error : %v\n", err)
		return err
	}
	
	//items.jsonから現在の商品一覧を読み込む
	//Items構造体に従った、jsonからのデータ格納用変数定義。ここではまだ何も入っていない
	var items Items
	//FormFileで画像受取
	//err処理
	//hash値を受取でhash化
	image, err := c.FormFile("image")
	if err != nil{
		fmt.Printf("FormFile error : %v\n", err)
		return err
	}
	imageName, err := getHash(image)
	c.Logger().Infof("Receive images : images=%s",imageName)
	//go/images/<hash化>.jpgしたい

	dst, err := os.Create("images/"+imageName+".jpg")
	//filepath.Join("images",imageName+".jpg")
	if err != nil{
		fmt.Printf("strconv.Atoi error : %v\n", err)
		return err
	}
	defer dst.Close()
	c.Logger().Infof("Receive images : images=%s",imageName)
	//Openして保存
	src, err := image.Open()
	if err != nil {
		fmt.Printf("image.Open error : %v\n",err)
		return err
	}
	defer src.Close()
	if _, err := io.Copy(dst,src); err != nil {
		log.Fatal(err)
	}
	//_値を握りつぶしたときだけ、↑;err != の書き方ができる！

	newItem := Item{Name: name, Category: category, Id: id, Images: imageName}
	fmt.Printf("Received name: %s, category: %s\n, id %s, imageName: %s", name, category, id, imageName)


	//file参照用変数。
	file, err := os.Open("items.json")
	if err != nil{
		fmt.Printf("os.Open error : %v\n", err)
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&items); err != nil{
			//error printかく
			fmt.Printf("Decode error : %v\n", err)
			return err 
	}

	//新しい商品の追加
	items.Items = append(items.Items, newItem)
	//更新された商品一覧をitems.jsonに書き込む
	file, err = os.Create("items.json")
	if err != nil{
		fmt.Printf("os.Create error : %v\n", err)
		return err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(items); err != nil{
		fmt.Printf("Encoder error : %v\n", err)
		return err
	}

	// c.Logger().Infof("Receive item: %s", name)

	// message := fmt.Sprintf("item received: %s", name)
	// res := Response{Message: message}
	return c.JSON(http.StatusOK, newItem)
}


func getHash(image *multipart.FileHeader) (string, error) {
	
	//openで読込
	//hash
	//画像保存
	//json file開封
	//json fileに保存
	src, err := image.Open()
	if err != nil {
		fmt.Printf("image.Open error : %v\n",err)
		return "",err
	}
	defer src.Close()
	//hash値計算用の箱用意
	hash := sha256.New()
	//err処理
	//hash値計算(copyでメモリ節約)
	if _,err := io.Copy(hash,src); err != nil{
		fmt.Printf("io.Copy error : %v\n",err)
		return "", err
	}
	//hash値を渡す(これを他に渡す)
	HashedValue := hex.EncodeToString(hash.Sum(nil))
	return HashedValue, nil
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

func getItemById(c echo.Context) error{
	//リクエストからid取得
	itemIdFromRequest,_ := strconv.Atoi(c.Param("item_id"))
	//jsonからfileを参照して開く
	//items.jsonを格納する変数
	//構造体に従ってreturnで返す
	var items Items
	file, err := os.Open("items.json")
	if err != nil{
		fmt.Printf("os.Open item.json error : %v\n",err)
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&items); err != nil {
		fmt.Printf("Decode error : %v\n", err)
		return err
	}
	//条件一致Item参照用変数定義
	//for でItemsスライスのItemを入れる
	//Paramでのidと一致したら出る
	var matchedItem *Item
	//rangeは2つ(index/keyと値)を返す。使わないものは_(アンダーバー)。
	for _, item := range items.Items{
		if item.Id == itemIdFromRequest {
			matchedItem = &item
			break
		}
	}
	
	if matchedItem == nil {
		return c.JSON(http.StatusNotFound, Response{Message: "Item not Found"})
	}
	return c.JSON(http.StatusOK, matchedItem)
}

func getItems (c echo.Context) error {
		//jsonfileを開く
		file, err := os.Open("items.json")
		var items Items
		if err != nil {
			items = Items{}
			fmt.Printf("os.Open json error : %v\n",err)
			return err
		} else{
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&items); err != nil{
			fmt.Printf("Decode error : %v\n", err)
			return err 
			}
		}
		return c.JSON(http.StatusOK, items)
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
	e.GET("/items", getItems)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:item_id", getItemById)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
