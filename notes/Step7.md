# STEP7: Implement a simple Mercari webapp as frontend

## 0. Build local environment

Clean install npm:
```shell
npm ci
```

After executing:
```shell
npm start
```
`npm` would find `package.json` and execute `start` in `scripts` parts, which is `react-scripts start` here. It is used to run React applications in my local development environment.

## 1. Add a new item

Without connecting to the backend api service, I could still input information of new items on the opened website `http://localhost:3000/`, but nothing happens.

Apparently the code in `python/main.py`
```shell
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
```
makes sense.

After connection, all items in `db/mercari.sqlite3` are listed on the website, but the pages are not my images in `images`.
I added a `cat` to the item list, with category `pet`.

## 2. Show item images

In `python/main.py`, mount the static file service on the path `/static` of React application.
```shell
app.mount("/static", StaticFiles(directory="images", html=True), name="static")
```

In `ItemList.tsx`, use
```shell
<img src={`${server}/static/${item.image_name}`} />
```
to fetch images from the static folder.

Output:
```shell
`INFO:     Uvicorn running on http://0.0.0.0:9000 (Press CTRL+C to quit)`
```
Then the API is running successfully:
```shell
curl -X POST \
  --url 'http://0.0.0.0:9000/items' \
  -F 'name=sofa' \
  -F 'category=furniture' \
  -F 'image=@/Users/xiaotongye/Desktop/images/sofa.jpg'
```
Output:
```shell
{"message":"item received: sofa"}
```

## 3. Change the styling with HTML and CSS

In `App.css`, change `.ItemList{}` part.

## 4. Change the UI for ItemList

In `App.css`, define a `Container`.
In `ItemList`, use `Container` to change the `return` part.