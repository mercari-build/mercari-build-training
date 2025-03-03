-- FOR 5-1
-- CREATE TABLE IF NOT EXISTS items (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,  -- 自動的にインクリメントされるID
--     name TEXT,                    -- アイテムの名前
--     category TEXT,                -- アイテムのカテゴリー（NULL不可したければNOT NULL）
--     image_name TEXT                        -- アイテムに関連する画像のファイル名（NULL許可）
-- );

-- FOR 5-2 itemsというテーブルを削除して作り直すのが正しい方法だと思いますが、履歴を残すためnewitemsというテーブルを作ります。
CREATE TABLE IF NOT EXISTS newitems (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    category_id INTEGER,
    image_name TEXT,
    FOREIGN KEY (category_id) REFERENCES categories(id) --category_id = (table:categories) id 
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE --重複なし
);
