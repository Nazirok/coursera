package main

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

import (
	"database/sql"
	"encoding/json"

	_ "github.com/go-sql-driver/mysql"

	"fmt"
	"net/http"
)

func NewDbExplorer(db *sql.DB) (*dbExplorer, error) {
	tablesData := make(map[string][]map[string]interface{})
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		rows.Scan(&table)
		fields, err := db.Query("SHOW FULL COLUMNS FROM " + table)
		if err != nil {
			return nil, err
		}
		defer fields.Close()

		columns, err := fields.Columns()
		if err != nil {
			return nil, err
		}

		count := len(columns)
		values := make([]interface{}, count)
		scanArgs := make([]interface{}, count)
		for i := range values {
			scanArgs[i] = &values[i]
		}

		for fields.Next() {
			err := fields.Scan(scanArgs...)
			if err != nil {
				return nil, err
			}
			entry := make(map[string]interface{})
			for i, col := range columns {
				v := values[i]

				b, ok := v.([]byte)
				if ok {
					entry[col] = string(b)
				} else {
					entry[col] = v
				}
			}
			tablesData[table] = append(tablesData[table], entry)
		}
	}

	jsonData, err := json.Marshal(tablesData)
	if err != nil {
		fmt.Println("marshal", err)
	}

	fmt.Printf("%s\n", jsonData)

	return &dbExplorer{DB: db, Tables: tablesData}, nil
}

type dbExplorer struct {
	DB     *sql.DB
	Tables map[string][]map[string]interface{}
}

func (d *dbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
