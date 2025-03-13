# STEP6: Ensure API Behavior Using Tests

In this step, you'll learn about testing.

**:book: Reference**

* (JA)[テスト駆動開発](https://www.amazon.co.jp/dp/4274217884)
* (EN)[Test Driven Development: By Example](https://www.amazon.co.jp/dp/0321146530)


## What is "testing"?

Testing is the process of evaluating and ensuring if a system or component's behavior and performance meet specifications and requirements. Let's look at a simple sayHello function example:

```go
func sayHello(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}
```

This function creates a string `Hello, ${name}!` using the `name` parameter. But does it behave correctly? Testing helps us ensure this.

In Go, we can write tests like this (don't worry about the details for now, just skim through):

```go
func TestSayHello(t *testing.T) {
    t.Run("Alice", func(t *testing.T) {
        // Expected return value is "Hello, Alice!"
        want := "Hello, Alice!"

        // Argument is "Alice" 
        arg := "Alice"
        // Actually call sayHello
        got := sayHello(arg)

        // Check if expected and actual values match
        if want != got {
            // Display error if they don't match
            t.Errorf("unexpected result of sayHello: want=%v, got=%v", want, got)
        }
    })
}
```

Running this produces:

```bash
=== RUN   TestSayHello
=== RUN   TestSayHello/Alice
--- PASS: TestSayHello (0.00s)
    --- PASS: TestSayHello/Alice (0.00s)
PASS
```

This is how we can test functionality.

## Purpose of "testing"

Tests serve several purposes:

- Finding defects
- Verifying requirement compliance
- Performance evaluation
- Reliability assessment
- Security validation
- Usability evaluation
- Maintainability assessment
etc.

A major benefit is guaranteeing expected behavior. For example, if we accidentally introduced an unwanted character ( `#` ):

```go
func sayHello(name string) string {
    return fmt.Sprintf("Hello, %s!#", name)
}
```

While this might be missed during manual review, tests would catch it:

```bash
=== RUN   TestSayHello
=== RUN   TestSayHello/Alice
    prog_test.go:20: unexpected result of sayHello: want=Hello, Alice!, got=Hello, Alice!#
--- FAIL: TestSayHello (0.00s)
    --- FAIL: TestSayHello/Alice (0.00s)
FAIL
```

By using tests to guarantee behavior, we can maintain code quality. Furthermore, when implementing complex features, writing tests for smaller components allows us to ensure working parts while development progresses. This makes it easier to identify problematic areas when unexpected bugs occur, enabling faster responses compared to not having tests.

## Types of "testing"

There are various types of tests for different purposes.

For simplicity, we'll focus on two types: Unit Tests (testing at the component level) and End-to-End Tests (E2E Tests, which simulate user operations on the integrated system). Feel free to research other types independently.

Let's consider a concrete example: testing an API for uploading images to a photo-sharing site. The image upload API would have functions/methods that receive image data and return results. We can test this using expected inputs and outputs.

However, setting up databases and servers for each test can be cumbersome. Instead, we can replace the actual database storage function/method with a mock implementation that returns fixed values. These test replacements are called mocks.

Using mocks, we can verify behavior for both successful and failed database operations without setting up an actual database. However, since mocks use values we specify, they might not perfectly match real behavior.

Tests using small functionality or mock data are called Unit Tests, while tests using actual databases and data to test complete functionality are called End-to-End tests (E2E tests).

Generally, it's recommended to have more unit tests than E2E tests. Unit tests are faster and require fewer resources, while E2E tests are slower and resource-intensive. For example, testing with real data might require preparing multiple test datasets and performing multiple save/delete operations. With large-scale data, execution times increase and resource usage grows, so it's standard practice to have fewer E2E tests and cover functionality with more unit tests. However, using only unit tests might miss environment-specific issues, so balance is important.

## Strategies for "testing"

Test approaches vary by language and framework. This section explains test strategies for Go and Python and demonstrates how to write tests.

### Go

**:book: Reference**

- (EN)[testing package - testing - Go Packages](https://pkg.go.dev/testing)
- (EN)[Add a test - The Go Programming Language](https://go.dev/doc/tutorial/add-a-test)
- (EN)[Go Wiki: Go Test Comments - The Go Programming Language](https://go.dev/wiki/TestComments)
- (EN)[Go Wiki: TableDrivenTests - The Go Programming Language](https://go.dev/wiki/TableDrivenTests)

Go provides a standard `testing` package for test functionality, and tests can be run using the `$ go test` command. For Go's testing guidelines, refer to [Go Wiki: Go Test Comments](https://go.dev/wiki/TestComments). These are language-level general guidelines that should be followed when appropriate.

Let's start by writing a unit test for our earlier code. Go recommends table-driven tests where test cases are listed and tested sequentially. Test cases are typically declared in slices or maps - maps are generally preferred unless order is important, as order-independent test cases provide stronger guarantees of functionality.

```go
func TestSayHello(t *testing.T) {
    cases := map[string]struct{
        name string
        want string
    }{
        "Alice": {
            name: "Alice",
            want: "Hello, Alice!"
        }
        "empty": {
            name: "",
            want: "Hello!"
        }
    }

    for name, tt := range cases {
        t.Run(name, func(t *testing.T) {
            got := sayHello(tt.name)

            if tt.want != got {
                t.Errorf("unexpected result of sayHello: want=%v, got=%v", tt.want, got)
            }
        })
    }
}
```

Writing test cases together like this makes it easy to see inputs and expected outputs at a glance. When reading unfamiliar code, test code can serve as a helpful reference for understanding behavior.

It's also important to consider test-friendly argument design. For example, if we modify the greeting based on time:

```go
func sayHello(name string) string {
    now := time.Now()
    currentHour := now.Hour()

    if 6 <= currentHour && currentHour < 10 {
        return fmt.Sprintf("Good morning, %s!", name)
    }
    if 10 <= currentHour && currentHour < 18 {
        return fmt.Sprintf("Hello, %s!", name)
    }
    return fmt.Sprintf("Good evening, %s!", name)
}
```

To test all time periods, we'd need to run tests at different times. This isn't ideal for testing. We can improve the design:

```go
func sayHello(name string, now time.Time) string {
    currentHour := now.Hour()

    if 6 <= currentHour && currentHour < 10 {
        return fmt.Sprintf("Good morning, %s!", name)
    }
    if 10 <= currentHour && currentHour < 18 {
        return fmt.Sprintf("Hello, %s!", name)
    }
    return fmt.Sprintf("Good evening, %s!", name)
}
```

Now we can freely set the current time and test different time periods:

```go
func TestSayHelloWithTime(t *testing.T) {
    type args struct {
        name string
        now time.Time
    }
    cases := map[string]struct{
        args
        want string
    }{
        "Morning Alice": {
            args: args{
                name: "Alice",
                now: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
            },
            want: "Good morning, Alice!",
        },
        "Hello Bob": {
            args: args{
                name: "Bob",
                now: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
            },
            want: "Hello, Bob!",
        },
        "Night Charlie": {
            args: args{
                name: "Charlie",
                now: time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
            },
            want: "Good evening, Charlie!",
        },
    }

    for name, tt := range cases {
        t.Run(name, func(t *testing.T) {
            got := sayHello(tt.name, tt.now)

            if tt.want != got {
                t.Errorf("unexpected result of sayHello: want=%v, got=%v", tt.want, got)
            }
        })
    }
}
```

This demonstrates writing code with testing in mind.

### Python

# Python Testing Strategies

**:book: Reference**

- (EN)[pytest: helps you write better programs — pytest documentation](https://docs.pytest.org/en/stable/)
- (EN)[pytest fixtures: explicit, modular, scalable — pytest documentation](https://docs.pytest.org/en/6.2.x/fixture.html)
- (EN)[Parametrizing fixtures and test functions — pytest documentation](https://docs.pytest.org/en/stable/how-to/parametrize.html)


While Python has the standard unittest library built-in for testing, the **pytest** library is widely used to write more flexible and readable tests. pytest comes with a simple API and powerful features, and can be easily installed with `pip install pytest`. You can run tests with the `$ pytest` command.

In Python, you can use the `pytest.mark.parametrize` decorator to describe multiple test cases together. Let's write a test for the say_hello function:

```python
# hello.py
def say_hello(name=""):
    if name:
        return f"Hello, {name}!"
    return "Hello!"

# test_hello.py
import pytest
from hello import say_hello

@pytest.mark.parametrize("name, expected",[
    ("Alice", "Hello, Alice!"),
    ("", "Hello!"),
]
)
def test_say_hello(name, expected):
    got = say_hello(name)

    # Check if the expected return value and the actual value are the same, and display an error if they differ
    assert got == expected, f"unexpected result of say_hello: want={expected}, got={got}"
```

The need to consider argument design with testing in mind is common to both Python and Go. Let's consider modifying the `say_hello` implementation to change the greeting based on the time of day:

```python
from datetime import datetime

def say_hello(name):
    now = datetime.now() # Directly depends on the current time, making it difficult to test
    current_hour = now.hour

    if 6 <= current_hour < 10:
        return f"Good morning, {name}!"
    if 10 <= current_hour < 18:
        return f"Hello, {name}!"
    return f"Good evening, {name}!"
```

This function is difficult to test because it directly depends on the current time. To test each time period, you would need to run the test at that specific time.

To make it more testable, we can rewrite the function as follows:

```python
# Improved code (more testable design)
from datetime import datetime

def say_hello(name, now=None):
    if now is None:
        now = datetime.now()
    
    current_hour = now.hour

    if 6 <= current_hour < 10:
        return f"Good morning, {name}!"
    if 10 <= current_hour < 18:
        return f"Hello, {name}!"
    return f"Good evening, {name}!"
```

Now we can specify the current time as an argument. By setting None as the default value, we can still omit the now parameter in normal usage.

```python
import pytest
from datetime import datetime
from greetings import say_hello

@pytest.mark.parametrize("name, now, expected", [
    ("Alice", datetime(2024, 1, 1, 9, 0, 0), "Good morning, Alice!"),
    ("Bob", datetime(2024, 1, 1, 12, 0, 0), "Hello, Bob!"),
    ("Charlie", datetime(2024, 1, 1, 20, 0, 0), "Good evening, Charlie!"),
])
def test_say_hello_simple(name, now, expected):
    got = say_hello(name, now)
    assert got == expected, f"unexpected result of say_hello: want={expected}, got={got}"
```

## 1. Writing Tests for the Item Listing API

Let's write tests for basic functionality, specifically testing item registration requests.

The expected request should require `name` and `category` fields and should return an error when this data is missing. Let's test this.

### Go

Let's look at `server_test.go`.

We want to ensure AddItem requests succeed when all values are present and fail when values are missing.
Let's write test cases for this.

**:beginner: Point**

- What does this test verify?
- What's the difference between `t.Error()` and `t.Fatal()`?

### Python (Read Only)

Python testing is implemented in [`main_test.py`](https://github.com/mercari-build/mercari-build-training/blob/main/python/main_test.py).

Unlike the Go API implementation, in Python API implementation using the FastAPI framework, developers do not need to implement HTTP Request parsing themselves. Therefore, no additional implementation is required in this chapter, but you should review the test code to deepen your understanding.

## 2. Writing Tests for the Hello Handler

Let's write handler tests.

Like in STEP 6-1, we can compare expected values with arguments.

### Go

**:book: Reference**

- (EN)[httptest package - net/http/httptest - Go Packages](https://pkg.go.dev/net/http/httptest)
- (JA)[Goのtestを理解する - httptestサブパッケージ編 - My External Storage](https://budougumi0617.github.io/2020/05/29/go-testing-httptest/)

Let's use Go's `httptest` library for testing handlers.

Unlike STEP 6-1, the comparison code isn't written yet.

- What do we want to test with this handler?
- How can we verify it's behaving correctly?

Once you have the logic figured out, implement it.

**:beginner: Point**

- Check other people's test code
- Review what the httptest package's existing code does

### Python

- (En)[FastAPI > Learn > Tutorial - User Guide / Testing](https://fastapi.tiangolo.com/tutorial/testing/)

In Python, we use FastAPI's `testclient.TestClient` to verify that the handler function `hello` works correctly. Let's edit the test function [test_hello](https://github.com/mercari-build/mercari-build-training/blob/main/python/main_test.py#L53) that's already provided and write a test.

As with Go, let's implement the test code with the following considerations in mind:

- What do you want to test with this handler?
- How can you verify that it behaves correctly?

For implementing the test, you may refer to the [official FastAPI documentation](https://fastapi.tiangolo.com/tutorial/testing/#testclient).

## 3. Writing Tests Using Mocks

Let's write tests using mocks.

As mentioned earlier, mocks replace actual logic with convenience functions that return expected data. Mocks can be used in various ways.

Consider our item registration to database. We want to test both successful and failed database operations. Intentionally creating these scenarios can be cumbersome, and using real databases might make tests flaky due to database issues.

Using mocks that return expected values instead of actual database logic allows us to test all scenarios.

### Go

**:book: Reference**

- (EN) [mock module - go.uber.org/mock - Go Packages](https://pkg.go.dev/go.uber.org/mock)

Go has various mock libraries; we'll use `gomock`.
Refer to documentation and existing blogs for basic usage.

Let's test both successful and failed persistence scenarios using mocks.

**:beginner: Point**

- Consider the benefits of using interfaces to satisfy mocks
- Think about the pros and cons of using mocks

### Python (Read Only)

**:book: Reference**

- (EN) [pytest-mock](https://github.com/pytest-dev/pytest-mock)
- (EN) [unittest.mock --- introduction](https://docs.python.org/3.13/library/unittest.mock-examples.html#)

For Python mock libraries, there are several options including the built-in standard `unittest.mock` and pytest's `pytest-mock`. Mocks become necessary when the process being tested depends on external tools or objects, such as in the following cases:

- Mocking database connections to test user authentication logic without connecting to an actual database.
- Mocking HTTP API clients to test weather forecast retrieval functions without actual network communication.
- Mocking the file system to test logging functionality without actual file operations.

In our case, we could consider implementing a test like the first example mentioned: "mocking database connections." However, the Build Python API implementation is very simple, and setting up classes like ItemRepository for mock testing would unnecessarily complicate the implementation.

Since sufficient verification can be done with the test code implemented in the chapter "4. Writing tests using actual databases," and because it would contradict Python's language philosophy of "simplicity" and "explicitness," **we have omitted Python implementations using mocks from this teaching material**.

However, in actual development environments where applications become more complex, there are many cases where tests using mocks are implemented in Python as well. If you're interested, take a look at the explanation of mock testing in the Go section, or review Python test implementations using mocks that are introduced on the internet.

## 4. Writing Tests Using Real Databases

Let's write tests replacing STEP 6-3's mocks with actual databases.

While mocks can test various scenarios, they aren't running in real environments. Sometimes code that works with mocks might fail with real databases. Let's prepare a test database and run tests with it.

### Go

In Go, we'll create a database file for testing and add operations to it.

After performing database operations, we need to verify the database state matches expectations.

- What should the database state be after item registration?
- How can we verify it's behaving correctly?

### Python

Let's write a test in Python using a test database (sqlite3). Uncomment the two places in [main_test.py](https://github.com/mercari-build/mercari-build-training/blob/main/python/main_test.py) that say "`STEP 6-4: uncomment this test setup`". ([first location](https://github.com/mercari-build/mercari-build-training/blob/main/python/main_test.py#L9-L42)/[second location](https://github.com/mercari-build/mercari-build-training/blob/main/python/main_test.py#L60-L84))

The `db_connection` function creates and sets up a new test database using sqlite3 before the test, and deletes the test database after the test is completed.

The `test_add_item_e2e` function tests the item addition functionality by sending a POST request to the API endpoint (`/items/`). This function runs with parameterized test cases (valid and invalid data). The test verifies:

1. Whether the response status code matches the expected value
2. For non-error cases, whether the response body contains a "message"
3. Whether the data was correctly saved in the database (matching name and category)

What's particularly important is that it tests end-to-end using an actual database (for testing) rather than mocks, which verifies the functionality in a way that's closer to the actual environment.

## Next

[STEP7: Run the application in a virtual environment](./07-docker.en.md)
