/* ************************************************************************** */
/*   util.go
/* ************************************************************************** */

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

func getAllItems() (Items, error) {
	var items Items
	fileContent, err := os.ReadFile("items.json")
	if err != nil {
		return items, err
	}
	if err := json.Unmarshal(fileContent, &items); err != nil {
		return items, err
	}
	return items, nil
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
		return item, fmt.Errorf("category_id %d does not exitst", categoryID)
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

func addItemToFile(item Item) error {
	var items Items
	_, err := os.Stat("items.json")
	if err == nil {
		fileContent, err := os.ReadFile("items.json")
		if err != nil {
			return err
		}
		if len(fileContent) != 0 {
			if err := json.Unmarshal(fileContent, &items); err != nil {
				return err
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	items.Items = append(items.Items, item)
	updatedContent, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return os.WriteFile("items.json", updatedContent, 0644)
}
