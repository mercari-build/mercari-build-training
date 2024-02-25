CREATE TABLE items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    category_id INTEGER,
    image_name TEXT,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
