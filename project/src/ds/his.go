// his
package main

import (
	"bytes"
	"database/sql"
	"dbs"
	"fmt"

	_ "github.com/mattn/go-oci8"
)

//var HIS_DB *sql.DB

type HIS struct {
	conn *sql.DB
}

func NewHIS() *HIS {
	strDSN := `system/manager@//130.1.10.90:1521/orcl`
	db, err := sql.Open("oci8", strDSN)
	db.Ping()
	if err != nil {
		return nil
	}
	return &HIS{conn: db}
}

func (h *HIS) staffInfo(buf *bytes.Buffer) bool {
	strSql := `select * from comm.staff_dict`

	rows, err := h.conn.Query(strSql)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer rows.Close()
	dbs.Rows2Json(rows, buf)
	return true
}
