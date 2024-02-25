/* ************************************************************************** */
/*   util.go
/* ************************************************************************** */

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func saveImageAndHash(imageFile io.Reader) (string, error) {
	tempFile, err := os.CreateTemp("images", "*.jpg")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()
	hasher := sha256.New()
	_, err = io.Copy(io.MultiWriter(tempFile, hasher), imageFile)
	if err != nil {
		return "", err
	}
	hashedFileName := hex.EncodeToString(hasher.Sum(nil)) + ".jpg"
	hashedFilePath := filepath.Join("images", hashedFileName)
	if err := os.Rename(tempFile.Name(), hashedFilePath); err != nil {
		return "", err
	}
	return hashedFileName, nil
}

func createNewItem(c echo.Context) (Item, error) {
	var item Item
	name := c.FormValue("name")
	categoryIDStr := c.FormValue("category_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return item, err
	}
	categoryExist, err := checkCategoryIDExist(categoryID)
	if err != nil {
		return item, err
	}
	if !categoryExist {
		return item, fmt.Errorf("category_id %d does not exist", categoryID)
	}
	image, err := c.FormFile("image")
	if err != nil {
		return item, err
	}
	src, err := image.Open()
	if err != nil {
		return item, err
	}
	defer src.Close()
	hashedFileName, err := saveImageAndHash(src)
	if err != nil {
		return item, err
	}
	item = Item{
		Name:       name,
		CategoryID: categoryID,
		ImageName:  hashedFileName,
	}
	return item, nil
}
