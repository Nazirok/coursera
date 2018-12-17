package main

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func NewDbExplorer(db *sql.DB) (*dbExplorer, error) {
	tables, err := getTables(db)
	if err != nil {
		return nil, err
	}
	tablesData := make(map[string]map[string]Field)
	for _, table := range tables {
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
	Tables map[string]map[string]Field
}

type resp map[string]interface{}

type Field struct {
	Name     sql.NullString
	Type     sql.NullString
	Pri      sql.NullBool
	AutoInc  sql.NullBool
	NullAble sql.NullBool
	Default  interface{}
}

func (d *dbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.URL.Path == "/" {
		if r.Method == http.MethodGet {
			d.handleBase(w, r)
		} else {
			writeResponse(w, resp{"error": "bad method"}, http.StatusInternalServerError)
		}
	} else {
		trimPath := strings.TrimSuffix(r.URL.Path, "/")
		p := strings.Split(trimPath, "/")[1:]
		if !d.tableExist(p[0]) {
			writeResponse(w, resp{"error": "unknown table"}, http.StatusNotFound)
			return
		}

		if len(p) == 1 && r.Method == http.MethodGet {
			d.handleGetAll(w, r, p[0])
			return
		}

		if len(p) == 2 && r.Method == http.MethodGet {
			for _, f := range d.Tables[p[0]] {
				if f.Pri.Bool {
					d.handleGetByID(w, p[0], f.Name.String, p[1])
					return
				}
			}
		}

		if len(p) == 1 && r.Method == http.MethodPut {
			body := make(map[string]interface{})
			defer r.Body.Close()
			dat, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal(dat, &body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			d.handlePUT(w, p[0], body)
			return
		}

		if len(p) == 2 && r.Method == http.MethodPost {
			body := make(map[string]interface{})
			defer r.Body.Close()
			dat, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal(dat, &body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			for _, f := range d.Tables[p[0]] {
				d.handlePOST(w, p[0], f.Name.String, p[1], body)
				return
			}
		}

		if len(p) == 2 && r.Method == http.MethodDelete {
			for _, f := range d.Tables[p[0]] {
				d.handleDELETE(w, p[0], f.Name.String, p[1])
				return
			}
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
	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", table)
	rows, err := d.DB.Query(query, limit, offset)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}
	out := make([]map[string]interface{}, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		entry := make(map[string]interface{})
		for i, col := range columns {
			v := values[i]
			if v == nil {
				entry[col.Name()] = interface{}(nil)
			} else {
				switch col.ScanType().Name() {
				case "int32":
					entry[col.Name()] = v.(int64)
				case "RawBytes":
					entry[col.Name()] = string(v.([]byte))
				}
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

func (d *dbExplorer) handleGetByID(w http.ResponseWriter, table string, pri, id string) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=?", table, pri)
	rows, err := d.DB.Query(query, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	columns, err := rows.ColumnTypes()
	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		entry := make(map[string]interface{})
		for i, col := range columns {
			v := values[i]
			if v == nil {
				entry[col.Name()] = interface{}(nil)
			} else {
				switch col.ScanType().Name() {
				case "int32":
					entry[col.Name()] = v.(int64)
				case "RawBytes":
					entry[col.Name()] = string(v.([]byte))
				}
			}
		}
		res := resp{
			"response": map[string]interface{}{
				"record": entry,
			},
		}
		writeResponse(w, res, http.StatusOK)
		return
	}
	writeResponse(w, resp{"error": "record not found"}, http.StatusNotFound)
}

func (d *dbExplorer) handlePUT(w http.ResponseWriter, table string, data map[string]interface{}) {
	var insert, values string
	insert += "INSERT INTO " + table + " ("
	values += " VALUES ("
	vals := make([]interface{}, 0)
	for k, v := range data {
		f, ok := d.Tables[table][k]
		if !ok || f.Pri.Bool || f.AutoInc.Bool {
			continue
		}
		switch v.(type) {
		case int, int32, float32, float64:
			if f.Type.String != "int(11)" {
				writeResponse(w, resp{"error": fmt.Sprintf("field %s have invalid type", k)}, http.StatusBadRequest)
				return
			}
		case string:
			if !(f.Type.String == "varchar(255)" || f.Type.String == "text") {
				writeResponse(w, resp{"error": fmt.Sprintf("field %s have invalid type", k)}, http.StatusBadRequest)
				return
			}
		}
		insert += "`" + k + "`,"
		values += "?,"
		vals = append(vals, v)
	}
	insert = strings.TrimSuffix(insert, ",") + ")"
	values = strings.TrimSuffix(values, ",") + ")"
	result, err := d.DB.Exec(insert+values, vals...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeResponse(w, resp{"response": map[string]interface{}{
		"id": lastID,
	}}, http.StatusOK)
}

func (d *dbExplorer) handlePUT(w http.ResponseWriter, table string, data map[string]interface{}) {
	var insert, values string
	insert += "INSERT INTO " + table + " ("
	values += " VALUES ("
	vals := make([]interface{}, 0)
	for k, v  := range d.Tables[table] {

	}


	for k, v := range data {
		f, ok := d.Tables[table][k]
		if !ok || f.Pri.Bool || f.AutoInc.Bool {
			continue
		}
		switch v.(type) {
		case int, int32, float32, float64:
			if f.Type.String != "int(11)" {
				writeResponse(w, resp{"error": fmt.Sprintf("field %s have invalid type", k)}, http.StatusBadRequest)
				return
			}
		case string:
			if !(f.Type.String == "varchar(255)" || f.Type.String == "text") {
				writeResponse(w, resp{"error": fmt.Sprintf("field %s have invalid type", k)}, http.StatusBadRequest)
				return
			}
		}
		insert += "`" + k + "`,"
		values += "?,"
		vals = append(vals, v)
	}
	insert = strings.TrimSuffix(insert, ",") + ")"
	values = strings.TrimSuffix(values, ",") + ")"
	result, err := d.DB.Exec(insert+values, vals...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeResponse(w, resp{"response": map[string]interface{}{
		"id": lastID,
	}}, http.StatusOK)
}

func (d *dbExplorer) handlePOST(w http.ResponseWriter, table, pri, id string, data map[string]interface{}) {
	var buf bytes.Buffer
	buf.WriteString("UPDATE " + table + " SET ")
	vals := make([]interface{}, 0)
	for k, v := range data {
		f, ok := d.Tables[table][k]
		if f.Pri.Bool {
			writeResponse(w, resp{"error": fmt.Sprintf("field %s have invalid type", k)}, http.StatusBadRequest)
			return
		}
		if !f.NullAble.Bool && v == nil {
			writeResponse(w, resp{"error": fmt.Sprintf("field %s have invalid type", k)}, http.StatusBadRequest)
			return
		}
		if !ok || f.AutoInc.Bool {
			continue
		}
		switch v.(type) {
		case int, int32, float32, float64:
			if f.Type.String != "int(11)" {
				writeResponse(w, resp{"error": fmt.Sprintf("field %s have invalid type", k)}, http.StatusBadRequest)
				return
			}
		case string:
			if !(f.Type.String == "varchar(255)" || f.Type.String == "text") {
				writeResponse(w, resp{"error": fmt.Sprintf("field %s have invalid type", k)}, http.StatusBadRequest)
				return
			}
		}
		buf.WriteString(fmt.Sprintf("%s=?,", k))
		vals = append(vals, v)
	}
	vals = append(vals, id)
	query := strings.TrimSuffix(buf.String(), ",") + " WHERE " + pri + "=?"
	result, err := d.DB.Exec(query, vals...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeResponse(w, resp{"response": map[string]interface{}{
		"updated": affected,
	}}, http.StatusOK)
}

func (d *dbExplorer) handleDELETE(w http.ResponseWriter, table, pri, id string) {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s=?", table, pri)
	result, err := d.DB.Exec(query, id)
	if err != nil {
		fmt.Println("EXEC:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("AFFECTED:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeResponse(w, resp{"response": map[string]interface{}{
		"deleted": affected,
	}}, http.StatusOK)

}

func getTables(db *sql.DB) ([]string, error) {
	tables := make([]string, 0)
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}
	return tables, nil
}

func getTableColumns(db *sql.DB, table string) (map[string]Field, error) {
	rows, err := db.Query("SHOW FULL COLUMNS FROM " + table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]Field, 0)

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
		entry := Field{}
		for i, col := range columns {
			v := values[i]
			b, _ := v.([]byte)

			switch col {
			case "Field":
				entry.Name.Scan(b)
			case "Type":
				entry.Type.Scan(b)
			case "Null":
				entry.NullAble.Scan(string(b) == "YES")
			case "Key":
				entry.Pri.Scan(string(b) == "PRI")
			case "Extra":
				entry.AutoInc.Scan(string(b) == "auto_increment")
			case "Default":
				entry.Default = v
			}
		}
		out[entry.Name.String] = entry
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
