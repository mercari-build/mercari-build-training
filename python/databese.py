import sqlite3

conn = sqlite3.connect('../db/mercari.sqlite3')
c = conn.cursor()
with open('../db/items.db') as f:
    schema = f.read()
    c.execute(f"""CREATE TABLE {schema}""")

# create a table
# NULL INTEGER REAL TEXT BLOB

# many_items = [('0', 'a0', 'b0'), ('1', 'a1', 'b1'), ('2', 'a2', 'b2')]
# c.executemany("INSERT INTO items VALUES (?, ?, ?)", many_items)
# print("Command executed succesefully")

# Query
# c.execute("SELECT * FROM items")
# print(c.fetchone()[0])

# Delete table
# c.execute("DROP TABLE items")
# print("Table dropped... ")

# c.execute("INSERT INTO items VALUES ('123456789', 'Jacket', 'Fashion')")
# print("Command executed succesefully")

conn.commit()
conn.close()
