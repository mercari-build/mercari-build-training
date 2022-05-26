from multiprocessing import connection
import sqlite3
import googletrans
from googletrans import Translator
translator = Translator()

# translate and add english name and japanese name to the item
def add_item(name, category_id, image_hash):
    image = image_hash + ".jpg"
    en_name = ""
    ja_name = ""
    if translator.detect(name).lang == "ja":
        en_name = en_name + translator.translate(name, dest="en").text
        ja_name = ja_name + name
    else:
        ja_name = ja_name + translator.translate(name, dest="ja").text
        en_name = en_name + name
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
        INSERT INTO items (en_name, ja_name, category_id, image_filename)
        VALUES (?, ?, ?, ?)
    """, (en_name, ja_name, category_id, image))
    connection.commit()
    connection.close()


def get_items():
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT items.id, items.en_name, items.ja_name, category.en_name AS category_en_name, category.ja_name AS category_ja_name, items.image_filename
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
    SELECT items.id, items.en_name, items.ja_name, category.en_name AS category_en_name, category.ja_name AS category_ja_name, items.image_filename
    FROM items
    LEFT JOIN category
    ON items.category = category.id
    WHERE items.id = ?
    """, (item_id,))
    item = cursor.fetchone()
    connection.close()
    return item


# def search_items(keyword):
#     connection = sqlite3.connect("../db/mercari.sqlite3")
#     cursor = connection.cursor()
#     cursor.execute("""
#     SELECT items.id, items.name, items.category, category.name AS category_name, items.image_filename
#     FROM items
#     LEFT JOIN category
#     ON items.category = category.id
#     WHERE items.name LIKE ?
#     """, ("%" + keyword + "%",))
#     items = cursor.fetchall()
#     connection.close()
#     return items


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
#     id INTEGER PRIMARY KEY,
#     en_name TEXT,
#     ja_name TEXT
# )
# """)
# connection.commit()
# connection.close()

#drop table
# connection = sqlite3.connect("../db/mercari.sqlite3")
# cursor = connection.cursor()
# cursor.execute("""
# DROP TABLE IF EXISTS items
# """)
# connection.commit()
# connection.close()

# add category to the category table
# connection = sqlite3.connect("../db/mercari.sqlite3")
# cursor = connection.cursor()
# cursor.execute("""
# INSERT INTO category (en_name, ja_name)
# VALUES (?, ?)
# """, ("post-modern", "ポストモダン"))
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


