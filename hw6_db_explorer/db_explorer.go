package main

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"encoding/json"
	"net/http"
	"strings"
	"fmt"
)

func NewDbExplorer(db *sql.DB) (*dbExplorer, error) {
	defer func() {
		fmt.Println("Conns: ", db.Stats().OpenConnections)
	}()
	fmt.Println("Conns beign: ", db.Stats().OpenConnections)
	tablesData := make(map[string][]map[string]interface{})
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		rows.Scan(&table)
		columns, err := getTableColumns(db, table)
		if err != nil {
			return nil, err
		}
		tablesData[table] = columns
	}
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
		if !d.tableExist(p[0]) {
			writeResponse(w, resp{"error": "unknown table"}, http.StatusNotFound)
			return
		}
		if len(p) == 1 && r.Method == http.MethodGet {
			d.handleGetAll(w, r, p[0])
			return
		}
	}
}

func (d *dbExplorer) handleBase(w http.ResponseWriter, r *http.Request) {
	out := make([]string, 0, len(d.Tables))
	for k := range d.Tables {
		out = append(out, k)
	}
	res := resp{
		"response": map[string][]string{
			"tables": out,
		},
	}
	writeResponse(w, res, http.StatusOK)
}

func (d *dbExplorer) handleGetAll(w http.ResponseWriter, r *http.Request, table string) {
	var limit, offset string
	limit = r.URL.Query().Get("limit")
	if limit == "" {
		limit = "5"
	}
	offset = r.URL.Query().Get("offset")
	if offset == "" {
		offset = "0"
	}
	rows, err := d.DB.Query("SELECT * FROM " + table + " LIMIT " + limit + " OFFSET " + offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}
	out := make([]map[string]interface{}, 0, 5)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		entry := make(map[string]interface{})
		for i, col := range columns {
			v := values[i]
			switch v.(type) {
			case []byte:
				entry[col] = string(v.([]byte))
			case string:
				entry[col] = v.(string)
			case int:
				entry[col] = v.(int)
			case bool:
				entry[col] = v.(bool)
			case nil:
				entry[col] = interface{}(nil)
			}
		}
		out = append(out, entry)
	}
	res := resp{
		"response": map[string]interface{}{
			"records": out,
		},
	}
	writeResponse(w, res, http.StatusOK)
}

func getTableColumns(db *sql.DB, table string) ([]map[string]interface{}, error) {
	rows, err := db.Query("SHOW FULL COLUMNS FROM " + table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]interface{}, 0)

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(scanArgs...)
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
		out = append(out, entry)
	}
	return out, nil
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

// https://play.golang.org/p/kwc6sTg0SG1