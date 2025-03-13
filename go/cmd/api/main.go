package main

import (
	"fmt"
	"log"
	"mercari-build-training/app"
	"os"
)

const (
	port         = "9000"
	imageDirPath = "images"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Current working directory:", dir)
	// This is the entry point of the application.
	// You don't need to modify this function.
	os.Exit(app.Server{
		Port:         port,
		ImageDirPath: imageDirPath,
	}.Run())
}
