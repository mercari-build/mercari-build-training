# STEP3: Make a listing API

## 1. Call an API

For each input,

```shell
$ curl -X GET 'http://127.0.0.1:9000'
$ curl -X POST 'http://127.0.0.1:9000/items'
$ curl -X POST \
  --url 'http://localhost:9000/items' \
  -d name=jacket
```

The Output is,
```shell
INFO:     127.0.0.1:61493 - "GET / HTTP/1.1" 200 OK

INFO:     127.0.0.1:61519 - "POST /items HTTP/1.1" 422 Unprocessable Entity

INFO:     Receive item: jacket
INFO:     127.0.0.1:61527 - "POST /items HTTP/1.1" 200 OK
```
respectively.


* Understand the difference between GET and POST requests.

A:

Google definition: GET is used to retrieve information from the server. POST is used to create or update a resource.

GET is less secure, as data is exposed in URLs and can be easily seen and logged.
POST is more secure, as data is included in the body of the request, not in the URL.

* Why do we not see `{"message": "item received: <name>"}` on accessing `http://127.0.0.1:9000/items` from your browser?

A:

Because the item is unprocessable without necessary input in a POST request.

  * What is the **HTTP Status Code** when you receive these responses?
    
    A:
    
    Wikipedia definition: Status codes are issued by a server in response to a client's request made to the server.

  * What do different types of status code mean?

    A:

    `200` stands for success.
    `422` represents that although the request format is correct, the request is unprocessable on account of some semantic mistakes.
    `405` represents that the request method is not allowed.

## 2. List a new item

The struture of items.json is like:

```shell
{
  "items": {
    "1": {
      "name": "jacket",
      "category": "fashion",
      "image": "c763dfccdfafb916d75975b8ac8de174083cf5061ca26be264bab6274d9cd2ff.jpg"
    }
  },
  "cur_max_id": 1
}
```

Each item in `items` has one number id as key, and a dict as value.
When a new item is added, a new id would be assigned to it, and `cur_max_id` would add 1.

## 4. Add an image to an item

* What is hashing?

  A:

  Hashing means that transferring the input into a string with static length by a hashing function. Using hashes as filenames avoids some security issues such as predicting filenames or overwriting existing files.

* What other hashing functions are out there except for sha256?
  
  A:

  MD5, SHA-1 Family, RIPEMD.

## 6. (Optional) Understand Loggers

* What is log level?

A:

  Google definition: A log level is set up as an indicator within your log management system that captures the importance and urgency of all entries within the logs.

* On a web server, what log levels should be displayed in a production environment?
  INFO, WARN, ERROR and FATAL.

A:

  INFO provides information of basic system activities.
  WARN provides information of potentially harmful events.
  ERROR provides information of errors and exceptions that may interrupt the normal flow of the application.
  FATAL provides information of very serious error events that may cause the application to terminate.
