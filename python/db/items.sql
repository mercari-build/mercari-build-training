-- 商品情報を保存するtable
CREATE TABLE IF NOT EXISTS items(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    image_name TEXT NOT NULL
    FOREIGN KEY (category_id) REFERENCES categories(id) -- カテゴリの関連付け
);

INSERT INTO items (name, category, image_name) VALUES
('Tshirts',1,'default.jpg'),
('gloves',1,'default,jpg'),
('jacket',1,'default,jpg'),
('headphone',2,'default.jpg');


-- カテゴリー情報を保存するtable
CREATE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE -- UNIQUEを指定すると、同じカテゴリーを登録できない
);

INSERT INTO categories (name) VALUES
('fashion'),
('accessories');
