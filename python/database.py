from contextlib import closing, nullcontext
import sqlite3
import hashlib

# file path of the database
filename = '../db/mercari.sqlite3'

"""
Returns the list of all items in the database
"""
def get_items():
    items = []

    with closing(sqlite3.connect(filename)) as db_connect:
        db_cursor = db_connect.cursor()
        # insert new data
        sql = 'SELECT * FROM items'
        db_cursor.execute(sql)
        items = db_cursor.fetchall()
        db_connect.commit()

    return items

"""
Search item with the given id
Return the item if present, return None otherwise
"""
def get_id_by_id(item_id):
    item = []

    with closing(sqlite3.connect(filename)) as db_connect:
        db_cursor = db_connect.cursor()
        # insert new data
        sql = 'SELECT * FROM items WHERE id = ?'
        data = (item_id,)
        db_cursor.execute(sql, data)
        item = db_cursor.fetchone()
        db_connect.commit()

    return item



"""
Add a new item with the given name, category and image to the database
"""
def add_item(name, category, image_hash):

    with closing(sqlite3.connect(filename)) as db_connect:
        db_cursor = db_connect.cursor()
        sql = 'INSERT INTO items(name, category, image) values (?, ?, ?)'
        data = [name, category, image_hash]
        db_cursor.execute(sql, data)
        db_connect.commit()

"""
Search items with the given string keyword from the database.
Returns the list of items where its name contains the keyword.
"""
def search_items(keyword):
    items = []

    with closing(sqlite3.connect(filename)) as db_connect:
        db_cursor = db_connect.cursor()
        sql = 'SELECT * FROM items WHERE name LIKE ?'
        data = ('%' + keyword + '%',)
        db_cursor.execute(sql, data)
        items = db_cursor.fetchall()
        db_connect.commit()

    return items