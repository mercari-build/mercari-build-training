# Build@Mercari Training Program 

## 1. Call an API

For following inputs, output is as follows:

```shell
$ curl -X GET 'http://127.0.0.1:9000'
$ curl -X POST 'http://127.0.0.1:9000/items'
$ curl -X POST \
  --url 'http://localhost:9000/items' \
  -d name=jacket
```

The Output is,
```shell
INFO:     127.0.0.1:63469 - "GET / HTTP/1.1" 200 OK
INFO:     127.0.0.1:63512 - "POST /items HTTP/1.1" 422 Unprocessable Entity
INFO:     Receive item: jacket
INFO:     127.0.0.1:63743 - "POST /items HTTP/1.1" 200 OK
```
respectively.
* Q1. Understand the difference between GET and POST requests.
What I learned from my working is GET requests are used to retrieve data from a specified resource and 
POST requests are used to submit data to be processed to a specified resource. 
When we want to retrieve data from the server without modifying it, we use GET requests and we use POST requests when we need to send data to the server to create or update resources.

* Q2. Why do we not see `{"message": "item received: <name>"}` on accessing `http://127.0.0.1:9000/items` from your browser?
  It happens as the item is unprocessable without necessary input in a POST request and the correct message is typically returned in response to a successful POST request where the server receives an item with a name.

* Q3. What do different types of status code mean?
    200 stands for success.
    422 represents unprocessable Entity, to indicate that the request is invalid or unprocessable due to missing input data.
    405 represents that the request method is not allowed.

## 4. Add an image to an item

* What is hashing?
    Hashing is the process of converting an input (or 'message') into a fixed-size string of bytes, typically in the form of a 'digest'.
* What other hashing functions are out there except for sha256?
  MD5, SHA-1 Family, RIPEMD.

## 6. (Optional) Understand Loggers

* What is log level?
  A log level is set up as an indicator within your log management system that captures the importance and urgency of all entries within the logs.They help in filtering and prioritizing logs, making it easier to focus on critical events and create meaningful alerts
* On a web server, what log levels should be displayed in a production environment?
  INFO, WARN, ERROR and FATAL.
  INFO provides information of basic system activities.
  WARN provides information of potentially harmful events.
  ERROR provides information of errors and exceptions that may interrupt the normal flow of the application.
  FATAL provides information of very serious error events that may cause the application to terminate.