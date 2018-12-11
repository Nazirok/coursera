package main

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"encoding/json"
	"net/http"
	"strings"
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

	//jsonData, err := json.Marshal(tablesData)
	//if err != nil {
	//	fmt.Println("marshal", err)
	//}
	//
	//fmt.Printf("%s\n", jsonData)

	return &dbExplorer{DB: db, Tables: tablesData}, nil
}

type dbExplorer struct {
	DB     *sql.DB
	Tables map[string][]map[string]interface{}
}

type resp map[string]interface{}

func (d *dbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.URL.Path == "/" {
		if r.Method == http.MethodGet {
			d.handleBase(w, r)
		} else {
			writeResponse(w, resp{"error": "bad method"}, http.StatusInternalServerError)
		}
	} else {
		p := strings.Split(r.URL.Path, "/")[1:]
		//l := len(p)
		if !d.tableExist(p[0]) {
			writeResponse(w, resp{"error": "unknown table"}, http.StatusNotFound)
		}
	}
}

func (d *dbExplorer) handleBase(w http.ResponseWriter, r *http.Request) {
	out := make([]string, 0, len(d.Tables))
	for k := range d.Tables {
		out = append(out, k)
	}
	res := resp{
		"error": "",
		"response": map[string][]string{
			"tables": out,
		},
	}
	writeResponse(w, res, http.StatusOK)
}

func writeResponse(w http.ResponseWriter, res resp, status int) {
	data, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Write(data)
	return
}

func (d *dbExplorer) tableExist(table string) bool {
	_, ok := d.Tables[table]
	return ok
}
