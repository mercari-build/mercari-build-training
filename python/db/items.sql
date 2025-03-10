CREATE TABLE IF NOT EXISTS categories (
    id  INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
);

CREATE TABLE IF NOT EXISTS items (
    id  INTEGER PRIMARY KEY AUTO_INCREMENT,
    name TEXT NOT NULL,
    category_id INTEGER,
    image_name TEXT,
    FOREIGN KEY (category_id) REFERENCES categories (category_id)

);

