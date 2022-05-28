from multiprocessing import connection
import sqlite3

def add_item(ja_name, en_name, category_id, image_hash):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    image = image_hash + ".jpg"
    cursor.execute("""
    INSERT INTO items (ja_name, en_name, category_id, image_filename)
    VALUES (?, ?, ?, ?)
    """, (ja_name, en_name, category_id, image))
    connection.commit()
    connection.close()


def get_items():
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT items.id, items.ja_name, items.en_name, category.name AS category_name, items.image_filename
    FROM items
    LEFT JOIN category
    ON items.category_id = category.id
    """)
    items = cursor.fetchall()
    connection.close()
    return items


def get_item(item_id):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT items.id, items.en_name, items.ja_name, category.name AS category_name, items.image_filename
    FROM items
    LEFT JOIN category
    ON items.category_id = category.id
    WHERE items.id = ?
    """, (item_id,))
    item = cursor.fetchone()
    connection.close()
    return item


def search_items(keyword):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT items.id, items.en_name, items.ja_name, items.category_id, category.name AS category_name, items.image_filename
    FROM items
    LEFT JOIN category
    ON items.category_id = category.id
    WHERE items.en_name LIKE ?
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


#create items table

# connection = sqlite3.connect("../db/mercari.sqlite3")
# cursor = connection.cursor()
# cursor.execute("""
# CREATE TABLE IF NOT EXISTS items (
#     id INTEGER PRIMARY KEY AUTOINCREMENT,
#     en_name TEXT,
#     ja_name TEXT,
#     category_id INTEGER,
#     image_filename TEXT
# )
# """)
# connection.commit()
# connection.close()

#create category table

# connection = sqlite3.connect("../db/mercari.sqlite3")
# cursor = connection.cursor()
# cursor.execute("""
# CREATE TABLE IF NOT EXISTS category (
#     id INTEGER PRIMARY KEY AUTOINCREMENT,
#     name TEXT
# )
# """)
# connection.commit()
# connection.close()

#drop table
# connection = sqlite3.connect("../db/mercari.sqlite3")
# cursor = connection.cursor()
# cursor.execute("""
# DROP TABLE IF EXISTS category
# """)
# connection.commit()
# connection.close()

#insert category fashion in category table
# connection = sqlite3.connect("../db/mercari.sqlite3")
# cursor = connection.cursor()
# cursor.execute("""
# INSERT INTO category (name)
# VALUES ("post-modern")
# """)
# connection.commit()
# connection.close()



#delete category from the category table
# connection = sqlite3.connect("../db/mercari.sqlite3")
# cursor = connection.cursor()
# cursor.execute("""
# DELETE FROM categories
# WHERE id = ?
# """, (4,))
# connection.commit()
# connection.close()


