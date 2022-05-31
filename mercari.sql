-- SQLite
CREATE TABLE IF NOT EXISTS items (
        id INTEGER PRIMARY KEY AUTOINCREMENT, 
        name TEXT NOT NULL, 
        category_id INTEGER,
        image TEXT NOT NULL
);

