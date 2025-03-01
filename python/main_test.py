from fastapi.testclient import TestClient
import pytest
import sqlite3
import os
from main import app
from main import get_db

TEST_DATABASE = "test_fastapi.db"

def override_get_db():
    conn = sqlite3.connect(TEST_DATABASE)
    conn.row_factory = sqlite3.Row
    try:
        yield conn
    finally:
        conn.close()

@pytest.fixture(autouse=True)
def db_connection():
    # Before the test is done, create a test database
    conn = sqlite3.connect(TEST_DATABASE)
    cursor = conn.cursor()
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255),
		category VARCHAR(255)
	)"""
    )
    conn.commit()
    conn.row_factory = sqlite3.Row # Return rows as dictionaries
    

    yield conn

    conn.close()
    # After the test is done, remove the test database
    if os.path.exists(TEST_DATABASE):
        os.remove(TEST_DATABASE)


app.dependency_overrides[get_db] = override_get_db

client = TestClient(app)

def test_add_item(db_connection):
    test_item = {"name": "test"}
    response = client.post("/items/", data=test_item)
    assert response.status_code == 200
    response_data = response.json()
    assert "message" in response_data
    
    # Check if the data was saved to the database correctly
    cursor = db_connection.cursor()
    cursor.execute("SELECT * FROM items WHERE name = ?", (test_item["name"],))
    db_item = cursor.fetchone()
    assert db_item is not None
    assert dict(db_item)["name"] == test_item["name"]