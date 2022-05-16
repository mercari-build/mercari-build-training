import sqlite3

# # connect to database
# connection = sqlite3.connect("../db/mercari.sqlite3")

# # create a cursor
# cursor = connection.cursor()

# # create table
# cursor.execute("""
# CREATE TABLE IF NOT EXISTS items (
#     id INTEGER PRIMARY KEY,
#     name TEXT,
#     category TEXT,
#     image TEXT
# )
# """)

# # commit changes
# connection.commit()

# # close connection
# connection.close()


def add_item(name, category, image_hash):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    image = image_hash + ".jpg"
    cursor.execute("""
    INSERT INTO items (name, category, image)
    VALUES (?, ?, ?)
    """, (name, category, image))
    connection.commit()
    connection.close()


def get_items():
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT * FROM items
    """)
    items = cursor.fetchall()
    connection.close()
    return items


def get_item(item_id):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT * FROM items
    WHERE id = ?
    """, (item_id,))
    item = cursor.fetchone()
    connection.close()
    return item


def search_items(keyword):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT * FROM items
    WHERE name LIKE ?
    """, ("%" + keyword + "%",))
    items = cursor.fetchall()
    connection.close()
    return items

def delete_item(item_id):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    DELETE FROM items
    WHERE id = ?
    """, (item_id,))
    connection.commit()
    connection.close()

# Delete a table

# connection = sqlite3.connect("../db/mercari.sqlite3")
# cursor = connection.cursor()
# # drop a table
# cursor.execute("""
# DROP TABLE items
# """)
# # commit changes
# connection.commit()
# # close connection
# connection.close()
