package main

import (
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"io"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strconv"

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

type Item struct{
	ID int `json:"id"`
	Name string `json:"name"`
	Category string `json:"category"`
	ImageName string `json:"hashedFilename"`
}

type Items struct{
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func saveImageAndHash(imageFile io.Reader)(string,error){
	tempFile,err:=os.CreateTemp("images","*.jpg")
	if err!=nil{
		return "",err
	}
	defer tempFile.Close()
	hasher:=sha256.New()
	_,err=io.Copy(io.MultiWriter(tempFile,hasher),imageFile)
	if err!=nil{
		return "",err
	}
	hashedFileName:=hex.EncodeToString(hasher.Sum(nil))+".jpg"
	hashedFilePath:=filepath.Join("images",hashedFileName)

	if err:=os.Rename(tempFile.Name(),hashedFilePath);err!=nil{
		return "",err
	}
	return hashedFileName,nil
}

func showAll(c echo.Context)error{
	items,err:=GetAllItems();
	if err!=nil{
		return c.JSON(http.StatusInternalServerError,err)
	}
	return c.JSON(http.StatusOK,items)
}


func showOne(c echo.Context)error{
	items,err:=GetAllItems();
	if err!=nil{
		return c.JSON(http.StatusInternalServerError,echo.Map{"error": err.Error()})
	}
	idStr:=c.Param("id")
	id,err:=strconv.Atoi(idStr)
	if err!=nil{
		return c.JSON(http.StatusBadRequest,echo.Map{"error":"InValid item ID"})
	}
	if id<0 ||id>=len(items.Items){
		return c.JSON(http.StatusNotFound,echo.Map{"error":"Item not found"})
	}
	item:=items.Items[id]
	// for _,item :=range items.Items{
	// 	if item.ID==id{
	// 		return c.JSON(http.StatusOK,item)
	// 	}
	// }
	// return c.JSON(http.StatusNotFound,echo.Map{"error":"Item not found"})
	return c.JSON(http.StatusOK,item)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category:=c.FormValue("category")
	image,err:=c.FormFile("image")
	if err!=nil{
		return err
	}
	src,err:=image.Open()
	if err!=nil{
		return err
	}
	defer src.Close()
	hashedFileNanme,err:=saveImageAndHash(src)
	if err!=nil{
		return err
	}
	item:=Item{
		Name:name,
		Category: category,
		ImageName: hashedFileNanme,
	}
	if err:=AddItemtoFile(item);err!=nil{
		return err
	}
	c.Logger().Infof("Receive item: %s,category: %s", name,category)
	// message := fmt.Sprintf("item received: %s,category: %s", name,category)
	// res := Response{Message: message}

	return c.JSON(http.StatusOK, item)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Errorf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func AddItemtoFile(item Item)error{
	var items Items
	if _,err:=os.Stat("items.json");err==nil{
		fileContent,err:=os.ReadFile("items.json")
		if err!=nil{
			return err
		}
		json.Unmarshal(fileContent,&items)
	}
	items.Items=append(items.Items,item)
	updatedContent,err:=json.Marshal(items)
	if err!=nil{
		return err
	}
	return os.WriteFile("items.json",updatedContent,0644)
}

func GetAllItems()(Items,error){
	var items Items
	fileContent,err:=os.ReadFile("items.json")
	if err!=nil{
		return items,err
	}
	json.Unmarshal(fileContent,&items)
	return items,nil
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
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
	e.GET("/items",showAll)
	e.GET("/items/:id",showOne)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
