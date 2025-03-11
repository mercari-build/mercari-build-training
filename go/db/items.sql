CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    category_id INTEGER NOT NULL,
    image_name TEXT NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

-- Insert default categories
INSERT OR IGNORE INTO categories (name) VALUES ('fashion');
INSERT OR IGNORE INTO categories (name) VALUES ('electronics');
INSERT OR IGNORE INTO categories (name) VALUES ('books'); 