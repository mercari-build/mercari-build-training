package main

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/swaggo/echo-swagger"

	_ "mercari-build-training-2022/app/docs"
	"mercari-build-training-2022/app/handler"
	"mercari-build-training-2022/app/models/db"
	"mercari-build-training-2022/app/models/customErrors/itemsError"
)

// Root.
// @Summary root
// @Description just root
// @Produce  json
// @Success 200 {array} main.Response
// @Failure 500 {object} any
// @Router /items [get]
func root(c echo.Context) error {
	res := handler.Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

// getImg is getting an image
// @Summary get an image
// @Description find an item by id
// @Produce json
// @Param id path string true "Item's image name"
// @Success 200 {obejct} File
// @Failure 500 {object} main.Response
// @Router /image/:itemImg [get]
func GetImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(handler.ImgDir, c.Param("itemImg"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := handler.Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(handler.ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

// @title Simple Mercari Items API
// @version 1.0
// @description This is a simple Mercari Items API.
// @host localhost:1313
// @BasePath /
func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = itemsError.ErrorHandler
	e.Validator = &handler.CustomValidator{}
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Initialize Handler
	handler := handler.Handler{ DB : db.DbConnection }
	defer handler.DB.Close()

	// Routes
	e.GET("/", root)
	// Users routes
	e.POST("/users", handler.AddUser)
	e.GET("/users/:id", handler.FindUser)
	// Items routes
	e.GET("/items", handler.GetItems)
	e.GET("/items/:id", handler.FindItem)
	e.POST("/items", handler.AddItem)
	e.GET("/items/search", handler.SearchItems)
	// Files routes
	e.GET("/image/:itemImg", GetImg)
	// swagger
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
