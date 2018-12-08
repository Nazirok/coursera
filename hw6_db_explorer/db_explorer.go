package main

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"fmt"
	"net/http"
)

func NewDbExplorer(db *sql.DB) (*dbExplorer, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var table string
		tableRows := make(map[string][]string)
		rows.Scan(&table)
		fmt.Println(table)
		fields, err := db.Query("SHOW FULL COLUMNS FROM `?`", table)
		if err != nil {
			return nil, err
		}
	}

	return &dbExplorer{DB: db}, nil
}

type dbExplorer struct {
	DB     *sql.DB
	Tables map[string][]string
}

type tableField struct {
	Name string
	Type string
	Null string
	Pri  string
}

func (d *dbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
