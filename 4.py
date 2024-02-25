import sqlite3

# Connect to the SQLite database
conn = sqlite3.connect('db/items.db')
cursor = conn.cursor()

# Endpoint to get a list of items
@app.get("/items")
def get_items():
    try:
        # Execute SQL query to fetch items from the database
        cursor.execute("SELECT * FROM items")
        items_list = cursor.fetchall()  # Fetch all rows
        return items_list
    except Exception as e:
        # Handle exceptions appropriately
        logger.error(f"Error fetching items from database: {e}")
        raise HTTPException(status_code=500, detail="Unable to fetch items from database.")

# Endpoint to add a new item
@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = Form(...)):
    logger.info(f"Received item: {name}")

    # Save image file (similar to your existing implementation)

    try:
        # Execute SQL query to insert new item into the database
        cursor.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
                       (name, category, hashed_img_name))
        conn.commit()  # Commit the transaction
        new_item_id = cursor.lastrowid  # Get the ID of the newly inserted item
        return {"id": new_item_id, "name": name, "category": category, "image_name": hashed_img_name}
    except Exception as e:
        # Handle exceptions appropriately
        logger.error(f"Error adding item to database: {e}")
        raise HTTPException(status_code=500, detail="Unable to add item to database.")
