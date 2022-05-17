import sqlite3
import os
print(os.getcwd())

dbname = 'mercari.sqlite3'
conn = sqlite3.connect(dbname)
cur = conn.cursor()

cur.execute("select * from sqlite_master where type='table'")
data = cur.fetchall()
print(data)

cur.execute("select * from items")

data = cur.fetchall()

print(data)

conn.commit()
cur.close()
conn.close()