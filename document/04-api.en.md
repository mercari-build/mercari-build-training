# STEP4: Make a listing API

## 1. Call an API

**:book: Reference**

* (JA) [Udemy Business - REST WebAPI サービス 設計](https://mercari.udemy.com/course/rest-webapi-development/)
* (JA) [HTTP レスポンスステータスコード](https://developer.mozilla.org/ja/docs/Web/HTTP/Status)
* (JA) [HTTP リクエストメソッド](https://developer.mozilla.org/ja/docs/Web/HTTP/Methods)
* (JA) [APIとは？意味やメリット、使い方を世界一わかりやすく解説](https://www.sejuku.net/blog/7087)

* (EN) [Udemy Business - API and Web Service Introduction](https://mercari.udemy.com/course/api-and-web-service-introduction/)
* (EN) [HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
* (EN) [HTTP request methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)

The goal of this section is to call the API with a tool.

### Tools to Call the API
You can call APIs from a browser, but using a tool is more convenient if you want to freely send requests. There're various tools such as GUI: [Insomnia](https://insomnia.rest/) / [Postman](https://www.postman.com/) and CUI: [HTTPie](https://github.com/httpie/cli) / cURL. In this tutorial, we'll use cURL, which is often used by engineers.

### Installing cURL
You can check whether cURL is installed with the following command:

```shell
$ curl --version
```

If the version number is shown after executing the command above, cURL is installed. If not, please install the command.

### Sending a GET Request

Let's send a GET reaquest with cURL to the API server we launched in the previous section.
If you haven't started the server, run the following command:

| Python                                                                                       | Go                                                                            |
|----------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| Move to python folder before running the command <br>`uvicorn main:app --reload --port 9000` | Move to python folder before running the command <br>`go run cmd/api/main.go` |


Before sending the request with cURL, check that you can access `http://127.0.0.1:9000` in a browser and see `{"message": "Hello, world!"}` displayed. If not, refer to the section 4 of the STEP2: Run Python/Go app([Python](./02-local-env.en.md#4-run-the-python-app), [Go](./02-local-env.en.md#4-run-the-go-app)).

Now, it's time to send the request with cURL. **Open a new terminal** and run the following command:

```shell
$ curl -X GET 'http://127.0.0.1:9000'
```

You should see `{"message": "Hello, world!"}` shown in the terminal as same as in the browser.

### Sending a POST Request and Modify the Code


Next, let's send a POST request. The sample code provides an endpoint `/items`, so let's send a request to this endpoint with cURL. Run the following command:

```shell
$ curl -X POST 'http://127.0.0.1:9000/items'
```

This endpoint expects to return `{"message": "item received: <name>"}` as an successful response. However, you should receive a different response here.

Modify the command as follows and you will receive `{"message": "item received: jacket"}`. Investigate why that happens and check the `python/main.py` or `go/app/server.go` file.

```shell
$ curl -X POST \
  --url 'http://localhost:9000/items' \
  -d 'name=jacket'
```

**:beginner: Points**

* Understand the difference between GET and POST requests.
* Why do we not see `{"message": "item received: <name>"}` on accessing `http://127.0.0.1:9000/items` from your browser?
  * What is the **HTTP Status Code** when you receive these responses?
  * What do different types of status code mean?

## 2. List a new item

**:book: Reference**

* (JA)[RESTful Web API の設計](https://docs.microsoft.com/ja-jp/azure/architecture/best-practices/api-design)
* (JA)[HTTP レスポンスステータスコード](https://developer.mozilla.org/ja/docs/Web/HTTP/Status)
* (EN) [RESTful web API design](https://docs.microsoft.com/en-us/azure/architecture/best-practices/api-design)
* (EN) [HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)

The goal of this section is to extend the `POST /items` endpoint and persist data related to `items`.

The current `POST /items` endpoint can accept the `name` parameter. Let's modify it so it can also accept a `category` parameter.

* `name`: Name of the item (string)
* `category`: Category of the item (string)

Since the current implementation doesn't persist data, let's modify the code to save data in a JSON file. Let's create a file named `items.json`, and register new items under `items` key.

When a new item is added, the content should be saved in the `items.json` as follows:
```json
{
  "items": [
    {
      "name": "jacket",
      "category": "fashion"
    },
    ... (other items will follow here)
  ]
}
```

### Additional Information about Go Persistence

The Go side of the code is implemented as follows. In this case, the `Insert()` method called within the `AddItem` method is the `Insert()` method of `itemRepository`.


```go
type Handlers struct {
	imgDirPath string
	itemRepo   ItemRepository
}

func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
  // (snip)
  err = s.itemRepo.Insert(ctx, item)
  // (snip)
}

type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
}

func NewItemRepository() ItemRepository {
	return &itemRepository{fileName: "items.json"}
}

type itemRepository struct {
	fileName string
}

func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// STEP 4-2: add an implementation to store an item

	return nil
}

func (s Server) Run() int {
  // (snip)
  itemRepo := NewItemRepository()
	h := &Handlers{imgDirPath: s.ImageDirPath, itemRepo: itemRepo}
  // (snip)
}
```

You may have noticed that `s.itemRepo` is not a struct called `itemRepository`, but rather an `interface` called `ItemRepository`. This `interface` is a type that represents a collection of methods. In this case, it only has a method called `Insert`. Therefore, any structure that has an `Insert` method can be set to this `ItemRepository`. In this example, within the `Run` method, since `itemRepository` struct is set to the `itemRepo`, the `Insert` method of `itemRepository` is called.

So, why is such abstraction necessary?
There are several reasons, but one of the benefits here is that it makes it easy to replace the method of persistence.
In this case, we are using JSON as the persistence method, but it becomes easy to replace it with a database or a test implementation.
At this point, the caller in the code does not need to be aware of the underlying implementation and can call it without worrying about the specifics, which means there is no need for major changes in the code.

This concept of abstraction is also touched upon in things like the UNIX philosophy book mentioned below, so if you're interested, you might want to read about it.

**:book: Reference**

* (EN)[book - Linux and the Unix Philosophy: Operating Systems](https://www.amazon.com/dp/B001HZZSEK)

## 3. Get the List of Items

The goal of this section is to implement the `GET /items` endpoint to get the list of registered items. 

After implementing the endpoint, the response should be as follows:

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

## 4. Add Item Images

The goal of this section is to allow users to upload an image for an item. 

Modify both `GET /items` and `POST /items` endpoints for that.

* Make a directory named `images`
* Hash the image using SHA-256, and save it with the name `<hashed-value>.jpg`
* Modify `items` such that the image file can be saved as a string

```shell
# POST the jpg file
curl -X POST \
  --url 'http://localhost:9000/items' \
  -F 'name=jacket' \
  -F 'category=fashion' \
  -F 'image=@images/local_image.jpg'
```

```json
{"items": [{"name": "jacket", "category": "fashion", "image_name": "510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg"}, ...]}
```


**:beginner: Point**

* What is hashing?
* What other hashing functions are out there except for SHA-256?


## 5. Return Item Details

The goal of this section is to create an endpoint which returns the detailed information of a single product.

Make an endpoint `GET /items/<item_id>` to return item details.
The `<item_id>` represents the ID indicating the order in which the item was registered.
Let's call up the list of items from the JSON file and return the information of the item at the item_id position.

```shell
$ curl -X GET 'http://127.0.0.1:9000/items/1'
{"name": "jacket", "category": "fashion", "image_name": "..."}
```

## 6. (Optional) Understand Loggers
Open `http://127.0.0.1:9000/images/no_image.jpg` on your browser.
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

[STEP5: Database](./05-database.en.md)
