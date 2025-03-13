package main

import (
	"mercari-build-training/app"
	"os"
)

const (
	port         = "9000"
	imageDirPath = "images"
)

func main() {
	// This is the entry point of the application.
	// You don't need to modify this function.
	os.Exit(app.Server{
		Port:         port,
		ImageDirPath: imageDirPath,
	}.Run())
}
