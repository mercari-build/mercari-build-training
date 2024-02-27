import sqlite3
conn = sqlite3.connect('.tablesmercari.sqlite3')

create_table_query = """
CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    image_name TEXT NOT NULL
);
"""
cur = conn.cursor()
cur.execute(create_table_query)
conn.commit()