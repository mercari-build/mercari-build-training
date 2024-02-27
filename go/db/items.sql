CREATE TABLE items(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, category_id INT, image_name TEXT, FOREIGN KEY (category_id) REFERENCES categories (id));
