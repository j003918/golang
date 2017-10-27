// tinydb
package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
)

func OpenDb(timeout time.Duration, driver, dsn string, maxOpen, maxIdle int) (*sql.DB, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	mydb, err := _openDb(driver, dsn, maxOpen, maxIdle)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return mydb, err
}

func _openDb(driver, dsn string, maxOpen, maxIdle int) (*sql.DB, error) {
	mydb, err := sql.Open(driver, dsn)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	mydb.SetMaxOpenConns(maxOpen)
	mydb.SetMaxIdleConns(maxIdle)

	err = mydb.Ping()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return mydb, err
}

//for insert update delete
func ModifyTab(timeout time.Duration, mydb *sql.DB, strsql string, args ...interface{}) (RowsAffected int64, ok bool) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	rowCount, ok := _modifyTab(mydb, strsql, args...)

	select {
	case <-ctx.Done():
		return -1, false
	default:
	}
	return rowCount, ok
}

func _modifyTab(mydb *sql.DB, strsql string, args ...interface{}) (RowsAffected int64, ok bool) {
	rst, err := mydb.Exec(strsql, args...)
	if err != nil {
		return 0, false
	}

	rowCount, _ := rst.RowsAffected()
	return rowCount, true
}

func SQL2Json(timeout time.Duration, mydb *sql.DB, strSql string, out_buf *bytes.Buffer) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	return _sql2Json(ctx, mydb, strSql, out_buf)
}

func _sql2Json(ctx context.Context, mydb *sql.DB, strSql string, out_buf *bytes.Buffer) error {
	if "" == strings.Trim(strSql, " ") {
		return errors.New("err msg")
	}

	rows, err := mydb.Query(strSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()

	if err != nil {
		return err
	}

	//fix bug time.Time nil
	//values := make([]sql.RawBytes, len(columns))
	values := make([]sql.NullString, len(columns))
	scans := make([]interface{}, len(columns))

	for i := range values {
		scans[i] = &values[i]
	}

	type Jitem struct {
		Item string `json:"e"`
	}
	var jitem Jitem

	out_buf.WriteByte('[')
	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			//panic(err.Error())
			fmt.Println(err.Error())
			return err
		}

		out_buf.WriteByte('{')
		var strVal string
		for i, col := range values {
			if !col.Valid {
				strVal = "null"
			} else {
				jitem.Item = col.String
				bs, _ := json.Marshal(&jitem)
				strVal = string(bs[6 : len(bs)-2])
			}

			columName := strings.ToLower(columns[i])
			cell := fmt.Sprintf(`"%v":"%v"`, columName, strVal)
			out_buf.WriteString(cell + ",")
		}
		out_buf.Bytes()[out_buf.Len()-1] = '}'
		out_buf.WriteByte(',')
	}
	if out_buf.Len() > 1 {
		out_buf.Bytes()[out_buf.Len()-1] = ']'
	} else {
		out_buf.WriteByte(']')
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
