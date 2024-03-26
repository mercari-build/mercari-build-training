package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"io"
	"io/ioutil"
	"encoding/json"
	"crypto/sha256"
	"strconv"
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
	// DbPath = "../db/mercari.sqlite3"  // For running locally
	DbPath = "db/mercari.sqlite3" // For running in docker container
)

type Response struct {
	Message string `json:"message"`
}

type ItemList struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
	Image string `json:"image_name"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

/**
 This function handles errors; write the error to the log and send back
 the response with the error message.
 **/
func errorHandler(c echo.Context, err error, message string) error {
	c.Logger().Errorf("Error: ", err)
	res := Response{Message: message + ": " + err.Error()}
	return c.JSON(http.StatusInternalServerError, res)
}

// **************************************************************************
// In Step 3, save the data in a JSON file 
/**
 This function reads the item list from "items.json". It first checks if
 the file exists or not. If it exists, reads the data and returns the list.
 Otherwise, it returns an empty ItemList. Any error occurred is returned with
 the list.
 **/
func readFromFile(c echo.Context) (ItemList, error) {
	var list ItemList
	if _, err := os.Stat("items.json"); err == nil {
		// if "items.json" exists
		file, err := os.Open("items.json")
		if err != nil {
			return ItemList{}, errorHandler(c, err, "Error: opening a file")
		}
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&list); err != nil && err != io.EOF {
			fmt.Println("Error: ", err)
			return ItemList{}, errorHandler(c, err, "Error: decoding a file")
		}
		file.Close()
	} else {
		// if not exist, nothing is read
		list = ItemList{}
	}
	return list, nil
}

// Add an item to the list
func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive item: %s", name)
	category := c.FormValue("category")
	c.Logger().Infof("Receive category: %s", category)
	image := c.FormValue("image")

	// Hash image name
	h := sha256.New()
	h.Write([]byte(strings.Split(image, ".")[0]))
	image = fmt.Sprintf("%x", h.Sum(nil)) + ".jpg"

	new_item := Item{Name: name, Category: category, Image: image}

	list, errHandler := readFromFile(c)
	if errHandler != nil {
		return errHandler
	}
	// Append the new item to the list
	new_list := ItemList{Items: append(list.Items, new_item)}

	// Open the file again to write the new list
	file, err := os.OpenFile("items.json", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return errorHandler(c, err, "Error: opening a file")
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(new_list); err != nil {
		return errorHandler(c, err, "Error: encoding a struct")
	}

	// Response
	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}
  
	// http.StatusCreated(201) is also good choice.StatusOK
  	// but in that case, you need to implement and return a URL
  	// that returns information on the posted item.
	return c.JSON(http.StatusOK, res)
}

// Show the item list in the JSON file
func getItem(c echo.Context) error {
	// Read the item list from the JSON file
	list, errHandler := readFromFile(c)
	if errHandler != nil {
		return errHandler
	}
	return c.JSON(http.StatusOK, list)
}

func getImg(c echo.Context) error {
	imgPath := c.Param("imageFilename")

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}

	// Hash image name
	h := sha256.New()
	h.Write([]byte(strings.Split(imgPath, ".")[0]))
	imgPath = fmt.Sprintf("%x", h.Sum(nil)) + ".jpg"

	// Create image path
	imgPath = path.Join(ImgDir, imgPath)

	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().SetLevel(log.DEBUG)  // set the log level
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

// Show the item assigned to the id in the list
func getItemById(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	list, errHandler := readFromFile(c)
	if errHandler != nil {
		return errHandler
	}
	// if the id is in the list
	if len(list.Items) < id {
		return errorHandler(c, nil, "Error: id not in the list")
	}
	return c.JSON(http.StatusOK, list.Items[id-1])
}
// **************************************************************************

// From here are functions working with the SQLite database

/*
 * Given rows of a table, extract the data and store it in the ItemList struct
 */
func storeInItemStruct(c echo.Context, rows *sql.Rows) (ItemList, error) {
	list := ItemList{}
	for rows.Next() {
		var item Item
        err := rows.Scan(&item.Name, &item.Category, &item.Image)
        if err != nil {
            return ItemList{}, errorHandler(c, err, "Error: rows.Scan")
        }
		list.Items = append(list.Items, item)
    }
	return list, nil
}

/*
 * Insert a new item on the two tables in the database
 * items: (id, name, category_id, image_name)
 * categories: (id, name)
 * curl -X POST --url http://localhost:9000/items -F name=jacket -F category=fashion -F image=@images/jacket.jpg
 */
func addItemDatabase(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive item: %s", name)
	category := c.FormValue("category")
	c.Logger().Infof("Receive category: %s", category)
	image, _ := c.FormFile("image")
	image_name := image.Filename

	// Open file for reading
	imageFile, err := image.Open()
	if err != nil {
		return errorHandler(c, err, "Error: opening an image file")
	}
	defer imageFile.Close()

	// Hash image name
	image_name = strings.Replace(image_name, "@images/", "", 1)
	h := sha256.New()
	h.Write([]byte(strings.Split(image_name, ".")[0]))
	image_name = fmt.Sprintf("%x", h.Sum(nil)) + ".jpg"

	// Create image file
	fmt.Println("image_path: ", path.Join(ImgDir, image_name))
	storedImageFile, err := os.Create(path.Join(ImgDir, image_name))
	if err != nil {
		return errorHandler(c, err, "Error: creating an image file")
	}
	defer storedImageFile.Close()

	// Copy image to file
	_, err = io.Copy(storedImageFile, imageFile); 
	if err != nil {
		return errorHandler(c, err, "Error: copying the image file")
	}

	// Connect to the database
	db, err := sql.Open("sqlite3", DbPath)
	if err != nil {
		return errorHandler(c, err, "Error: sql.Open")
	}
	defer db.Close()

	// Get the category id; create one if not exists
	var category_id int
	row := db.QueryRow("SELECT id FROM categories WHERE name=?", category)
	err = row.Scan(&category_id)
    if err == sql.ErrNoRows {
		cmd := "INSERT INTO categories (name) VALUES (?)"
		res, err := db.Exec(cmd, category)
		if err != nil {
			return errorHandler(c, err, "Error: db.Exec")
		}
		id, err := res.LastInsertId()
		if err != nil {
            return errorHandler(c, err, "Error: LastInsertId")
        }
		category_id = int(id)
    } else if err != nil {
        return errorHandler(c, err, "Error: row.Scan")
    }

	// Insert the new item to the database
	cmd := "INSERT INTO items (name, category_id, image_name) VALUES ($1, $2, $3)"
	_, err = db.Exec(cmd, name, category_id, image_name)
	if err != nil {
		return errorHandler(c, err, "Error: db.Exec")
	}

	// Response
	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

/*
 * Combine the two tables in the database and show the item list 
 */
func getItemDatabase(c echo.Context) error {
	// Connect to the database
	db, err := sql.Open("sqlite3", DbPath)
	if err != nil {
		return errorHandler(c, err, "Error: sql.Open")
	}
	defer db.Close()

	// Get the item list
	rows, err := db.Query("SELECT i.name, c.name, i.image_name FROM items AS i INNER JOIN categories AS c ON i.category_id=c.id")
    defer rows.Close()
    if err != nil {
		return errorHandler(c, err, "Error: Database Query")
    }

	// Store in the ItemList struct
	list, errHandler := storeInItemStruct(c, rows) 
	if (err != nil) {
		return errHandler
	}
	return c.JSON(http.StatusOK, list)
}

/*
 * Search items that match the keyword given
 */
func search(c echo.Context) error {
	keyword := c.QueryParam("keyword")

	// Connect to the database
	db, err := sql.Open("sqlite3", DbPath)
	if err != nil {
		return errorHandler(c, err, "Error: sql.Open")
	}
	defer db.Close()

	// Get the item list based on the keyword
	cmd := "SELECT i.name, c.name, i.image_name FROM items AS i INNER JOIN categories AS c ON i.category_id=c.id WHERE i.name LIKE $1 OR c.name LIKE $1"
	rows, err := db.Query(cmd, "%" + keyword + "%")
    defer rows.Close()
    if err != nil {
		return errorHandler(c, err, "Error: Database Query")
    }

	// Store in the ItemList struct
	list, errHandler := storeInItemStruct(c, rows) 
	if (err != nil) {
		return errHandler
	}
	return c.JSON(http.StatusOK, list)
}

// Show the item assigned to the id in the list
func getItemByIdDatabase(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	
	// Connect to the database
	db, err := sql.Open("sqlite3", DbPath)
	if err != nil {
		return errorHandler(c, err, "Error: sql.Open")
	}
	defer db.Close()

	// Get the item list based on the keyword
	cmd := "SELECT i.name, c.name, i.image_name FROM items AS i INNER JOIN categories AS c ON i.category_id=c.id WHERE i.id=?"
	row := db.QueryRow(cmd, id)
    if err != nil {
		return errorHandler(c, err, "Error: Database Query")
    }

	// Store in the Item struct
	var item Item
	err = row.Scan(&item.Name, &item.Category, &item.Image)
	if err == sql.ErrNoRows {
		return errorHandler(c, err, "Error: id is not found")
	} else if err != nil {
		return errorHandler(c, err, "Error: row.Scan")
	}

	return c.JSON(http.StatusOK, item)
}

// **************************************************************************

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Create database tables if not exists
	db, err := sql.Open("sqlite3", DbPath)
    if err != nil {
        fmt.Println(err)
        return 
    }
	defer db.Close()

	// Read schema from items.db
	sqlCommands, err := ioutil.ReadFile("db/items.db")
    if err != nil {
        fmt.Println("Error reading SQL file:", err)
        return 
    }
    // Split SQL commands by semicolon
    commands := strings.Split(string(sqlCommands), ";")

    // Execute SQL commands
    for _, cmd := range commands {
        cmd = strings.TrimSpace(cmd)
        if cmd == "" {
            continue
        }
        _, err := db.Exec(cmd)
        if err != nil {
            fmt.Println("Error executing SQL command:", err)
            return 
        }
    }

	// Routes
	e.GET("/", root)
	// e.POST("/items", addItem)
	// e.GET("/items", getItem)
	e.POST("/items", addItemDatabase)
	e.GET("/items", getItemDatabase)
	e.GET("/image/:imageFilename", getImg)
	// e.GET("/items/:id", getItemById)
	e.GET("/items/:id", getItemByIdDatabase)
	e.GET("/search", search)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
