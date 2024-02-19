package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"
	"path"

	"github.com/labstack/echo/v4"
)

// todo: shorten the args
func httpErrorHandler(err error, c echo.Context, code int, message string) *echo.HTTPError {
	c.Logger().Error(err)
	return echo.NewHTTPError(code, message)
}

func registerImg(header *multipart.FileHeader) (string, error) {
	// Read uploaded file
	src, err := header.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Convert src to hash
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}
	hex_hash := hex.EncodeToString(hash.Sum(nil))

	// Reset the read position of the file
	if _, err := src.Seek(0, 0); err != nil {
		return "", err
	}

	// Save file to images/
	filename := hex_hash + path.Ext(header.Filename)
	file, err := os.Create(path.Join(ImgDir, filename))
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(file, src); err != nil {
		return "", err
	}

	return filename, nil
}
