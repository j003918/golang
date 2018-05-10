// testdb
package main

import (
	"bytes"
	"database/sql"
	"dbs"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-oci8"
)

func mssqlInfo(buf *bytes.Buffer) error {
	strConn := `sqlserver://sa:www126.com@130.1.10.217:1433?database=rights&encrypt=disable`
	strSql := `select @@version,GETDATE() as sysdate`

	db, err := sql.Open("mssql", strConn)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	rows, err := db.Query(strSql)
	if err != nil {
		return nil
	}
	defer rows.Close()

	return dbs.Rows2Json(rows, buf)
}

func mysqlInfo(buf *bytes.Buffer) error {
	strConn := `root:root@tcp(130.1.10.230:3306)/zyyoutdoor`
	strSql := `select version(), now()`

	db, err := sql.Open("mysql", strConn)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	rows, err := db.Query(strSql)
	if err != nil {
		return nil
	}
	defer rows.Close()

	return dbs.Rows2Json(rows, buf)
}

func oracleInfo(buf *bytes.Buffer) error {
	//strConn := `dc/dc@hdc`
	strConn := `dc/dc@//130.1.10.75:1521/hdc`
	strSql := `select VERSION,sysdate from v$instance`

	db, err := sql.Open("oci8", strConn)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	rows, err := db.Query(strSql)
	if err != nil {
		return nil
	}
	defer rows.Close()

	return dbs.Rows2Json(rows, buf) //dbs.Rows2Json(rows, buf)
}
