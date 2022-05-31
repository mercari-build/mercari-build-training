# STEP3: Make a listing API

## 1. Call an API

**:book: Reference**

* (JA) [Udemy -【基礎からわかる！】Webアプリケーションの仕組み](https://www.udemy.com/course/tobach_01_webapp_structure/)
* (JA) [HTTP レスポンスステータスコード](https://developer.mozilla.org/ja/docs/Web/HTTP/Status)
* (JA) [HTTP リクエストメソッド](https://developer.mozilla.org/ja/docs/Web/HTTP/Methods)
* (JA) [APIとは？意味やメリット、使い方を世界一わかりやすく解説](https://www.sejuku.net/blog/7087)

* (EN) [API and Web Service Introduction](https://www.udemy.com/course/api-and-web-service-introduction/)
* (EN) [HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
* (EN) [HTTP request methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)

### GET request

In step2, you ran a service on your local server where you accessed the endpoint from `http://127.0.0.1:9000` using your browser.
Use the following `curl` command to access this endpoint. Install `curl` if necessary.

```shell
curl -X GET 'http://127.0.0.1:9000'
```

Check if you can see `{"message": "Hello, world!"}` on your console.

### POST request.

In the example implementation, you can see `/items` endpoint. Use `curl` to call this endpoints.

```shell
$ curl -X POST 'http://127.0.0.1:9000/items'
```

This endpoint expects to return `{"message": "item received: <name>"}`, but you should be seeing something different.

Modify the command as follows and see that you receive `{"message": "item received: jacket"}`. Investigate what causes the differences.

```shell
$ curl -X POST \
  --url 'http://localhost:9000/items' \
  -d name=jacket
```

**:beginner: Points**

* Understand the difference betweeen GET and POST requests.
* Why do we not see `{"message": "item received: <name>"}` on accessing `http://127.0.0.1:9000/items` from your browser?
  * What is the **HTTP Status Code** when you receive these responses?
  * What do different types of status code mean?

## 2. List a new item

Make an endpoint to add a new item

**:book: Reference**

* (JA)[RESTful Web API の設計](https://docs.microsoft.com/ja-jp/azure/architecture/best-practices/api-design)
* (JA)[HTTP レスポンスステータスコード](https://developer.mozilla.org/ja/docs/Web/HTTP/Status)
* (EN) [RESTful web API design](https://docs.microsoft.com/en-us/azure/architecture/best-practices/api-design)
* (EN) [HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)

The endpoint already implemented (`POST /items`) takes `name` as an argument. Modify the API such that it also accepts `category` informaton.

* name: Name of the item (string)
* category: Category of the item (string)

Since the information cannot be retained with the current implementation, save this into a `JSON` file.
Make a file called `items.json` and add new items under `items` key.

`items.json` is expected to look like the following.
```json
{"items": [{"name": "jacket", "category": "fashion"}, ...]}
```

## 3. Get a list of items

Implement a GET endpoint `/items` that returns the list of all items. The response should look like the following.

```shell
# Add a new item
$ curl -X POST \
  --url 'http://localhost:9000/items' \
  -d 'name=jacket' \
  -d 'category=fashion'
# Expected response for /items endpoint with POST request
{"message": "item received: jacket"}
# Get a list of items
$ curl -X GET 'http://127.0.0.1:9000/items'
# Expected response for /items endpoint with GET request
{"items": [{"name": "jacket", "category": "fashion"}, ...]}
```

## 4. Write into a database

In the previous step, we saved data into `items.json`. We will move this into a database.
We will use a database called **SQLite**.

**:book: Reference**

* (JA)[SQLite入門](https://www.dbonline.jp/sqlite/)
* (JA)[Udemy -【SQLiteで学ぶ】ゼロから始めるデータベースとSQL超入門](https://www.udemy.com/course/basic_database_sqlite/)
* (JA)[Udemy - はじめてのSQLserver データベース　SQL未経験者〜初心者向けコース](https://www.udemy.com/course/sqlserver-for-beginner/)
* (EN)[https://www.sqlitetutorial.net/](https://www.sqlitetutorial.net/)
* (EN)[Udemy - Intro To SQLite Databases for Python Programming](https://www.udemy.com/course/using-sqlite3-databases-with-python/)

Install SQlite, and make a database file called `mercari.sqlite3`. Open this file and make a table called `items`. Define the items table as follows and save the schema into `db/items.db`.

* id: int (Identifier unique for each item)
* name: string (Name of the item)
* category: string (Category of the item)

Change the endpoints `GET /items` and `POST /items` such that items are saved into the database and can be returned based on GET request. Add `mercari.sqlite3` to your `.gitignore` such that it is not commited. `items.db` shoud be commited. 

**:beginner: Points**

* What are the advantages of saving into a databse such as SQLite instead of saving into a single JSON file?

## 5. Search for an item

Make an endpoint to return a list of items that include a specified keyword called `GET /search`.

```shell
# Request a list of items containing string "jacket"
$ curl -X GET 'http://127.0.0.1:9000/search?keyword=jacket'
# Expected response for a list of items with name containing "jacket"
{"items": [{"name": "jacket", "category": "fashion"}, ...]}
```

## 6. Add an image to an item

Change the endpoints `GET /items` and `POST /items` such that items can have images while listing.

* Make a directory called `images`
* Hash the image using sha256, and save it with the name `<hash>.jpg`
* Modify items table such that the image file can be saved as a string

```shell
# POST the jpg file
curl -X POST \
  --url 'http://localhost:9000/items' \
  -F 'name=jacket' \
  -F 'category=fashion' \
  -F 'image=@images/local_image.jpg'
```


Items table example:

| id   | name   | category | image_filename                                                       |
| :--- | :----- | :------- | :------------------------------------------------------------------- |
| 1    | jacket | fashion  | 510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg |
| 2    | ...    |          |                                                                      |

**:beginner: Point**

* What is hashing?
* What other hashing functions are out there except for sha256?


## 7. Return item details

Make an endpoint `GET /items/<item_id>` to return item details.

```shell
$ curl -X GET 'http://127.0.0.1:9000/items/1'
{"name": "jacket", "category": "fashion", "image": "..."}
```

## 8. (Optional) Move the category information to a separate table

Modify the database as follows. This allows changes in the category names and you will not have to change all categories in the items table when they do. Since `GET items` should still return the name of the category, join these two tables when returning responses.

**items table**

| id   | name   | category_id | image_filename                                                       |
| :--- | :----- | :---------- | :------------------------------------------------------------------- |
| 1    | jacket | 1           | 510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg |
| 2    | ...    |             |                                                                      |

**category table**

| id   | name    |
| :--- | :------ |
| 1    | fashion |
| ...  |         |

**:beginner: Points**
* What is database **normalization**?


## 9. (Optional) Understand Loggers
Open `http://127.0.0.1:9000/image/no_image.jpg` on your browser.
This returns an image called `no image` but the debug log is not displayed on your console.
```
Image not found: <image path>
```
Investigate the reason why this is the case. What changes should be made to display this message?

**:beginner: Points**
* What is log level?
* On a web server, what log levels should be displayed in a production environment?

---
**:beginner: Points**

Check if you understand the following concepts.

* port number
* localhost, 127.0.0.1
* HTTP request methods (GET, POST...)
* HTTP Status Code (What does each of 1XX, 2XX, 3XX, 4XX, 5XX mean?)

---

### Next

[STEP4: Run the application in a virtual environment](step4.en.md)