from fastapi import FastAPI
from fastapi.testclient import TestClient
from fastapi import Header, HTTPException
from main import app, get_db
import pytest
import sqlite3
import pathlib

# Define test database path
test_db = pathlib.Path(__file__).parent.resolve() / "db" / "test_mercari.sqlite3"
print(f"Using Test Database Path: {test_db}")

# Override the database dependency for testing
def override_get_db():
    conn = sqlite3.connect(test_db)
    conn.row_factory = sqlite3.Row
    try:
        yield conn
    finally:
        conn.close()

app.dependency_overrides[get_db] = override_get_db  # Removed indentation issue

@pytest.fixture(autouse=True)
def db_connection():
    """Setup a test database before tests and cleanup afterward."""
    conn = sqlite3.connect(test_db)
    cursor = conn.cursor()
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name VARCHAR(255),
            category VARCHAR(255)
        )"""
    )
    conn.commit()
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries

    yield conn  # Provide the test DB connection to tests

    conn.close()
    # Instead of deleting the file, clear the table for the next test
    cursor = conn.cursor()
    cursor.execute("DELETE FROM items")
    conn.commit()

client = TestClient(app)

@pytest.mark.parametrize(
    "want_status_code, want_body",
    [
        (200, {"message": "Hello, world!"}),
    ],
)
def test_hello(want_status_code, want_body):
    response = client.get("/")  # Fix: Define `response` properly
    response_body = response.json()  # Get JSON response
    assert response.status_code == want_status_code
    assert response_body == want_body

@pytest.mark.parametrize(
    "args, want_status_code",
    [
        ({"name": "used iPhone 16e", "category": "phone"}, 200),
        ({"name": "", "category": "phone"}, 400),
    ],
)
def test_add_item_e2e(args, want_status_code, db_connection):
    response = client.post("/items/", json=args)  # Fix: Use `json=args`
    assert response.status_code == want_status_code
    
    if want_status_code >= 400:
        return  # Skip DB checks if request failed
    
    response_data = response.json()
    assert "message" in response_data

    # Check if the data was saved to the database correctly
    cursor = db_connection.cursor()
    cursor.execute("SELECT * FROM items WHERE name = ?", (args["name"],))
    db_item = cursor.fetchone()
    assert db_item is not None
    assert dict(db_item)["name"] == args["name"]
