package main

const (
	ImgDir       = "images"
	ItemsJson    = "items.json"
	DbPath       = "../db/mercari.sqlite3"
	DbSchemaPath = "../db/items.db"
)

const (
	ItemsTableName      = "items"
	CategoriesTableName = "categories"
)

type Response struct {
	Message string `json:"message"`
}

type Items struct {
	Items []Item `db:"items"`
}

type Item struct {
	Id         int    `db:"id"`
	Name       string `db:"name"`
	CategoryId int    `db:"category_id"`
	ImageName  string `db:"image_name"`
}

type Categories struct {
	Categories []Category `db:"categories"`
}

type Category struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type JoinedItems struct {
	Items []JoinedItem `json:"items"`
}

type JoinedItem struct {
	Id           int    `db:"id"`
	Name         string `db:"name"`
	CategoryName string `db:"name"`
	ImageName    string `db:"image_name"`
}
