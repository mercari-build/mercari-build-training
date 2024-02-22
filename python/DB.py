import sqlite3

db_name = "mercari.sqlite3"   #arbitary file extension

def sql_connection():
    try:
        con = sqlite3.connect(db_name)
        print("Connection established.")
        return con
    except Exception as e:
        print(e)

def Tuple2Dict(tuple_list):
    dict_list = []
    for data in tuple_list:
        dict_list.append({"id": data[0], "name": data[1], \
                          "category": data[2], "image_name": data[3]})
    return dict_list

#get/items in items table
def Get_all(table_name):
    con = sql_connection()
    try:
        cursor = con.cursor()
        cursor.execute(f"Select * from {table_name}")
        dict_list = Tuple2Dict(cursor.fetchall())
        return {f"{table_name}": dict_list}
    except Exception as e:
        print(e)
    finally:
        con.close()

#get/items by joinly searching in items table and categories table
def Get_all_ex():
    con = sql_connection()
    try:
        cursor = con.cursor()
        cursor.execute(f"select I.id, I.name, C.name, I.image_name " \
                        "from items as I inner join categories as C" \
                        "on I.category_id = C.id")
        dict_list = Tuple2Dict(cursor.fetchall())
        return {"items": dict_list}
    except Exception as e:
        print(e)
    finally:
        con.close()

#insert/item
def Insert_item(table_name, item: [tuple,dict]):
    if type(item) == dict:
        item = tuple(item.values())
    try:
        con = sql_connection()
        cursor = con.cursor()
        cursor.execute(f"Insert into {table_name} (id, name, category, image_name) "
                       f"values (NULL, ?, ?, ?)", item)
        con.commit()
    except Exception as e:
        print(e)
    finally:
        con.close()

def SearchByKw(table_name, kw: str):
    try:
        con = sql_connection()
        cursor = con.cursor()
        cursor.execute(f"Select * from {table_name} where name like \'%{kw}%\'")
        dict_list = Tuple2Dict(cursor.fetchall())
        return {f"{table_name}": dict_list}

    except Exception as e:
        print(e)
    finally:
        con.close()

if __name__ == '__main__':
    pass




