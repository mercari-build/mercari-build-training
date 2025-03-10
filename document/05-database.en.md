# STEP5: Database
In the previous step, we saved data into `items.json`. We will move this into a database.

**:book: Reference**

* (JA)[SQLite入門](https://www.dbonline.jp/sqlite/)
* (JA)[Udemy Business - データベース講座１：データベース論理設計](https://mercari.udemy.com/course/database-logic/)

* (EN)[https://www.sqlitetutorial.net/](https://www.sqlitetutorial.net/)
* (EN)[Udemy Business - SQLite for beginners](https://mercari.udemy.com/course/sqlite-for-beginners/)
* (EN)[Udemy Business - Relational Database Designs](https://mercari.udemy.com/course/relational-database-design/)
## 1. Write into a database
We will use a database called **SQLite**.

* Install SQlite, and make a database file called `mercari.sqlite3`.  
* Open this file and make a table called `items`. 
* Define the items table as follows and save the schema into `db/items.sql`.
  * id: int (Identifier unique for each item)
  * name: string (Name of the item)
  * category: string (Category of the item)
  * image_name: string (Filename of the image)

Change the endpoints `GET /items` and `POST /items` such that items are saved into the database and can be returned based on GET request. Add `mercari.sqlite3` to your `.gitignore` such that it is not committed. `items.sql` should be committed. 

**:beginner: Points**

* What are the advantages of saving into a database such as SQLite instead of saving into a single JSON file?

## 2. Search for an item

Make an endpoint to return a list of items that include a specified keyword called `GET /search`.

```shell
# Request a list of items containing string "jacket"
$ curl -X GET 'http://127.0.0.1:9000/search?keyword=jacket'
# Expected response for a list of items with name containing "jacket"
{"items": [{"name": "jacket", "category": "fashion"}, ...]}
```

## 8. Move the category information to a separate table

Modify the database as follows. That makes it possible to change the category names without modifying the all categories of items in the items table.
Since `GET items` should return the category name as before, **join** these two tables when returning responses.

**items table**

| id   | name   | category_id | image_name                                                       |
| :--- | :----- | :---------- | :------------------------------------------------------------------- |
| 1    | jacket | 1           | 510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg |
| 2    | ...    |             |                                                                      |

**categories table**

| id   | name    |
| :--- | :------ |
| 1    | fashion |
| ...  |         |

**:beginner: Points**
* What is database **normalization**?



---

### Next

[STEP6: Ensure API Behavior Using Tests](./06-testing.en.md)
