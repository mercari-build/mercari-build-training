package main

const (
	ImgDir          = "images"
	ItemsJson       = "items.json"
	ItemsTablePath  = "../db/mercari.sqlite3"
	ItemsSchemaPath = "../db/items.db"
)

type Response struct {
	Message string `json:"message"`
}

type Items struct {
	Items []Item `db:"items"`
}

type Item struct {
	Id        int    `db:"id"`
	Name      string `db:"name"`
	Category  string `db:"category"`
	ImageName string `db:"image_name"`
}
