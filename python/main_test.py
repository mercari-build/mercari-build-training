from fastapi.testclient import TestClient
from main import app, get_db
import pytest
import sqlite3
import os
import pathlib

# STEP 6-4: uncomment this test setup
test_db = pathlib.Path(__file__).parent.resolve() / "db" / "test_mercari.sqlite3"


def override_get_db():
    conn = sqlite3.connect(test_db)
    conn.row_factory = sqlite3.Row
    try:
        yield conn
    finally:
        conn.close()


@pytest.fixture(autouse=True)
def db_connection():
    # Before the test is done, create a test database
    conn = sqlite3.connect(test_db)
    cursor = conn.cursor()
    # Create the categories table
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL
        )"""
    )

    # Create the items table
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name VARCHAR(255),
            category_id INTEGER,
            image TEXT,
            FOREIGN KEY (category_id) REFERENCES categories(id)
        )"""
    )
    conn.commit()
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries

    yield conn

    conn.close()
    # After the test is done, remove the test database
    if test_db.exists():
        test_db.unlink() # Remove the file

app.dependency_overrides[get_db] = override_get_db

client = TestClient(app)


@pytest.mark.parametrize(
    "want_status_code, want_body",
    [
        (200, {"message": "Hello, world!"}),
    ],
)
def test_hello(want_status_code, want_body):
    response_body = client.get("/").json()
    response_status_code = client.get("/").status_code
    # STEP 6-2: confirm the status code
    assert response_status_code == want_status_code, f"expected {want_status_code}, but got {response_body.status_code}"
    # STEP 6-2: confirm response body
    assert response_body == want_body, f"expected {want_body}, but got {response_body.json()}"


# STEP 6-4: uncomment this test
@pytest.mark.parametrize(
    # 引数をmain.pyのhello関数の引数に合わせて変更しました
    "args, image, want_status_code",
    [   # success, 想定通り200が返る
        (
            {"name": "used iPhone 16e", "category": "Eletronics"}, 
            {"image": ("default.jpg", open("images/default.jpg", "rb"))}, 
            200
         ),
        # Name is empty, 想定通り400エラーが返る
        (
            {"name":"", "category":"empty"},
            {"image" : ("default.jpg", open("images/default.jpg", "rb"))},
            400
        ),
    ],
)
def test_add_item_e2e(args, image, want_status_code, db_connection):
    response = client.post("/items/", data=args, files=image)
    assert response.status_code == want_status_code
    
    if want_status_code >= 400:
        return
    
    # Check if the response body is correct
    response_data = response.json()
    assert "message" in response_data

    # Check if the data was saved to the database correctly
    cursor = db_connection.cursor()
    cursor.execute("SELECT * FROM items WHERE name = ?", (args["name"],))
    db_item = cursor.fetchone()
    assert db_item is not None
    assert dict(db_item)["name"] == args["name"]