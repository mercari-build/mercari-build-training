
-- CREATE TABLE IF NOT EXISTS items (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     name TEXT NOT NULL,
--     category TEXT NOT NULL,
--     image_name TEXT
-- );
-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- Items table with foreign key to categories
CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    category_id INTEGER NOT NULL,
    image_name TEXT,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);




