CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY,
    name TEXT,
    category_id INTEGER,
    image_name TEXT,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
