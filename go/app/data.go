/* ************************************************************************** */
/*   data.go
/* ************************************************************************** */

package main

func getAllDB() (Items, error) {
	var items Items
	query := `SELECT items.id, items.name, items.category_id, categories.name AS category, items.image_name
	FROM items
	JOIN categories ON items.category_id = categories.id`

	rows, err := db.Query(query)
	if err != nil {
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Item
		err = rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.Category, &item.ImageName)
		if err != nil {
			return items, err
		}
		items.Items = append(items.Items, item)
	}
	return items, nil
}

func addItemToDB(item Item) (int64, error) {
	query := `INSERT INTO items (name, category_id, image_name) VALUES (?,?,?);`
	result, err := db.Exec(query, item.Name, item.CategoryID, item.ImageName)
	if err != nil {
		return 0, err
	}
	insertedId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return insertedId, err
}

func readKeywordDB(keyword string) (Items, error) {
	var items Items
	query := `SELECT items.id,items.name,items.category_id,categories.name AS category,items.image_name
	FROM items
	JOIN categories ON items.category_id=categories.id
	WHERE items.name LIKE ?`
	rows, err := db.Query(query, "%"+keyword+"%")
	if err != nil {
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Item
		err = rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.Category, &item.ImageName)
		if err != nil {
			return items, err
		}
		items.Items = append(items.Items, item)
	}
	return items, nil
}

func checkCategoryIDExist(categoryID int) (bool, error) {
	var exist bool
	query := "SELECT EXISTS(SELECT 1 FROM categories WHERE id=?)"
	err := db.QueryRow(query, categoryID).Scan(&exist)
	if err != nil {
		return false, err
	}
	return exist, nil
}
