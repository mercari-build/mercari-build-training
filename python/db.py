import sqlite3


def add_item(name, category_id, image_hash):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    image = image_hash + ".jpg"
    cursor.execute("""
    INSERT INTO items (name, category, image_filename)
    VALUES (?, ?, ?)
    """, (name, category_id, image))
    connection.commit()
    connection.close()


def get_items():
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT items.id, items.name, category.name AS category_name, items.image_filename
    FROM items
    LEFT JOIN category
    ON items.category = category.id
    """)
    items = cursor.fetchall()
    connection.close()
    return items


def get_item(item_id):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT items.id, items.name, category.name AS category_name, items.image_filename
    FROM items
    LEFT JOIN category
    ON items.category = category.id
    WHERE items.id = ?
    """, (item_id,))
    item = cursor.fetchone()
    connection.close()
    return item


def search_items(keyword):
    connection = sqlite3.connect("../db/mercari.sqlite3")
    cursor = connection.cursor()
    cursor.execute("""
    SELECT items.id, items.name, items.category, category.name AS category_name, items.image_filename
    FROM items
    LEFT JOIN category
    ON items.category = category.id
    WHERE items.name LIKE ?
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
