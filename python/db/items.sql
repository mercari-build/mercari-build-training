-- STEP 5-8: Seperate category table
CREATE TABLE IF NOT EXISTS category(
    id INTEGER PRIMARY KEY, 
    name TEXT NOT NULL 
);

-- items table modified
CREATE TABLE IF NOT EXISTS items(
    id INTEGER PRIMARY KEY AUTOINCREMENT, 
    name TEXT NOT NULL, 
    category_id INTEGER, 
    image_name TEXT,
    FOREIGN KEY (category_id) REFERENCES category(id)
);
