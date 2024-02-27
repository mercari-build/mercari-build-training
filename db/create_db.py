import sqlite3

conn = sqlite3.connect('mercari.sqlite3')
cur = conn.cursor()

# categories テーブルを作成
cur.execute("""
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
""")

# items テーブルを作成（category_idを含む）
cur.execute("""
CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    category_id INTEGER,
    image_name TEXT NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories (id)
);
""")

conn.commit()
conn.close()
