package controllers

import (
	"database/sql"
)

func ExecuteSQLQuery(db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		rowData := make(map[string]interface{})
		values := make([]interface{}, len(columns))
		for i := range columns {
			values[i] = new(interface{})
		}
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		for i, col := range columns {
			rowData[col] = *(values[i].(*interface{}))
		}
		results = append(results, rowData)
	}
	return results, nil
}
