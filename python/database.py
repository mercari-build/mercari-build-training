import sqlite3

# データベースを開く
conn = sqlite3.connect("../db/items.db")
c = conn.cursor()

#テーブルを作成
c.execute("CREATE TABLE `items` (`id` int, `name` string,`category` string);")

# sql = "INSERT INTO items(id,name,category) values(?,?,?)"

# data = [(111,"namae", "tops")]

# c.executemany(sql,data)

# 変更を確定
conn.commit()

def show_all():
    c.excute("SELECT * FROM items;")
    items = c.fetchall()

    for item in items:
        print(item)
    
def add_item(id,name,category):
    c.execute("INSERT INTO items(id,name,category) VALUES(?,?,?)", (id,name,category))
    
