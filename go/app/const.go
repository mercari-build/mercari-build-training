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

// type FactoryFunc func() interface{}

// var dbInfos = map[string]FactoryFunc{
// 	ItemsTableName:      func() interface{} { return new(Item) },
// 	CategoriesTableName: func() interface{} { return new(Category) },
// }

// type Table struct {
// 	TableInterface
// 	TableName string
// }

// type TableInterface interface {
// 	Add(interface{})
// 	NewInstance() interface{}
// }

// func (t *Table) NewInstance() interface{} {
// 	if factory_func, ok := dbInfos[t.TableName]; ok {
// 		object := factory_func()
// 		object.
// 		return factory_func()
// 	}
// 	return nil
// }

// func (i *Items) Add(item interface{}) {
// 	i.Items = append(i.Items, item.(Item))
// }

// func (c *Categories) Add(category interface{}) {
// 	c.Categories = append(c.Categories, category.(Category))
// }
