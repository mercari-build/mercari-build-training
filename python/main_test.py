from fastapi.testclient import TestClient
from main import app, get_db
import pytest
import sqlite3
import os
import pathlib

#STEP 6-4: uncomment this test setup
test_db = pathlib.Path(__file__).parent.resolve() / "db" / "test_mercari.sqlite3"

def override_get_db():
    conn = sqlite3.connect(test_db, check_same_thread=False)
    conn.row_factory = sqlite3.Row
    try:
        yield conn
    finally:
        conn.close()

app.dependency_overrides[get_db] = override_get_db

@pytest.fixture(autouse=True)
def db_connection():
    # Before the test is done, create a test database
    conn = sqlite3.connect(test_db)
    cursor = conn.cursor()
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS items2 (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
		    category_id INTEGER NOT NULL,
        image_name TEXT NOT NULL
	    )"""
    )
    conn.commit()
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
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
    response = client.get("/")
    # STEP 6-2: confirm the status code
    assert response.status_code == want_status_code
    # STEP 6-2: confirm response body
    response_body = response.json()
    assert response_body == want_body


# STEP 6-4: uncomment this test
@pytest.mark.parametrize(
    "args, want_status_code",
    [
        ({"name":"used iPhone 16e", "category":"phone"}, 200),
        ({"name":"", "category":"phone"}, 400),
    ],
)
def test_add_item_e2e(args,want_status_code,db_connection):
    with open("images/default.jpg", "rb") as f:
        files = {"image": ("default.jpg", f, "image/jpeg")}
        response = client.post("/items/", data=args, files = files)
        assert response.status_code == want_status_code
    
    if want_status_code >= 400:
        return
    
    
    # Check if the response body is correct
    response_data = response.json()
    assert "message" in response_data

    # Check if the data was saved to the database correctly
    cursor = db_connection.cursor()
    cursor.execute("SELECT * FROM items2 WHERE name = ?", (args["name"],))
    db_item = cursor.fetchone()
    assert db_item is not None
    assert dict(db_item)["name"] == args["name"]
    
    #Check the category is correct
    cursor.execute("SELECT id FROM categories WHERE name = ?", (args["category"],))
    db_category_id = cursor.fetchone()[0]
    assert dict(db_item)["category_id"] == db_category_id