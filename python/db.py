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
#     category TEXT
# )
# """)

# # commit changes
# connection.commit()

# # close connection
# connection.close()

def add_item(id, name, category):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    INSERT INTO items (id, name, category)
    VALUES (?, ?, ?)
    """, (id, name, category))
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